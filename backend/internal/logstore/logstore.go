// Package logstore persists the realtime request log.
//
// Every finished request is fanned out to three sinks:
//   - an in-process hub, used by the SSE endpoint (/api/v1/logs/stream);
//   - a capped Redis list + pub/sub channel, for realtime consumers;
//   - the PostgreSQL request_logs table, for durable history.
//
// Writes happen on a background worker so request latency is never affected.
package logstore

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/nugroho/siswa/backend/internal/config"
	"github.com/nugroho/siswa/backend/internal/dto"
	"github.com/nugroho/siswa/backend/internal/model"
)

const (
	// RedisListKey holds the most recent entries (capped at REQUEST_LOG_MAX_ENTRIES).
	RedisListKey = "siswa:request_logs"
	// RedisChannel is the pub/sub channel realtime consumers subscribe to.
	RedisChannel = "siswa:request_logs:stream"
)

// Store writes and reads request logs.
type Store struct {
	cfg  *config.Config
	rdb  *redis.Client
	pool *pgxpool.Pool

	queue  chan model.RequestLog
	hub    *hub
	wg     sync.WaitGroup
	closed chan struct{}
	once   sync.Once
}

// New builds the store and starts its background worker. Both rdb and pool may
// be nil; the store simply skips the sink that is unavailable.
func New(cfg *config.Config, rdb *redis.Client, pool *pgxpool.Pool) *Store {
	s := &Store{
		cfg:    cfg,
		rdb:    rdb,
		pool:   pool,
		queue:  make(chan model.RequestLog, 2048),
		hub:    newHub(),
		closed: make(chan struct{}),
	}

	s.wg.Add(1)
	go s.worker()
	return s
}

// Record enqueues a finished request. It never blocks: when the queue is full
// the entry is dropped and counted in the server log, because logging must not
// slow down or fail the API.
func (s *Store) Record(entry model.RequestLog) {
	select {
	case <-s.closed:
		return
	default:
	}

	select {
	case s.queue <- entry:
	default:
		slog.Warn("request log queue full, entry dropped", "url", entry.URL)
	}
}

// Subscribe returns a channel of live entries plus its cancel function.
func (s *Store) Subscribe(ctx context.Context) (<-chan model.RequestLog, func()) {
	return s.hub.subscribe(ctx)
}

// RedisAvailable reports whether Redis is wired in.
func (s *Store) RedisAvailable() bool { return s.rdb != nil }

func (s *Store) worker() {
	defer s.wg.Done()

	for entry := range s.queue {
		// A short, independent context: the request that produced the entry is
		// already finished and its context cancelled.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if s.rdb != nil && s.cfg.RequestLogToRedis {
			s.writeRedis(ctx, entry)
		}
		if s.pool != nil && s.cfg.RequestLogToDB {
			s.writeDB(ctx, entry)
		}
		cancel()
	}
}

func (s *Store) writeRedis(ctx context.Context, entry model.RequestLog) {
	payload, err := json.Marshal(entry)
	if err != nil {
		slog.Error("marshal request log", "error", err)
		return
	}

	pipe := s.rdb.TxPipeline()
	pipe.LPush(ctx, RedisListKey, payload)
	pipe.LTrim(ctx, RedisListKey, 0, s.cfg.RequestLogMaxEntries-1)
	pipe.Publish(ctx, RedisChannel, payload)
	if _, err := pipe.Exec(ctx); err != nil {
		slog.Error("push request log to redis", "error", err)
	}
}

func (s *Store) writeDB(ctx context.Context, entry model.RequestLog) {
	const q = `
		INSERT INTO request_logs
			(id, method, api_url, payload, executed_time_ms, status, response_code,
			 client_ip, user_agent, response_bytes, datetime)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO NOTHING`

	_, err := s.pool.Exec(ctx, q,
		entry.ID, entry.Method, entry.URL, entry.Payload, entry.ExecutedTime,
		entry.Status, entry.ResponseCode, entry.ClientIP, entry.UserAgent,
		entry.ResponseBytes, entry.Datetime,
	)
	if err != nil {
		slog.Error("insert request log", "error", err)
	}
}

// Broadcast publishes an entry to the in-process SSE hub. It is called by the
// middleware synchronously so subscribers see the event immediately.
func (s *Store) Broadcast(entry model.RequestLog) {
	s.hub.broadcast(entry)
}

// StartRedisBridge mirrors the Redis pub/sub channel into the local hub, so an
// SSE client connected to one replica also receives entries logged by another.
func (s *Store) StartRedisBridge(ctx context.Context) {
	if s.rdb == nil || !s.cfg.RequestLogToRedis {
		return
	}

	sub := s.rdb.Subscribe(ctx, RedisChannel)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer func() { _ = sub.Close() }()

		ch := sub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var entry model.RequestLog
				if err := json.Unmarshal([]byte(msg.Payload), &entry); err != nil {
					continue
				}
				// Entries produced by this replica were already broadcast
				// locally; the hub de-duplicates them by id.
				s.hub.broadcastOnce(entry)
			}
		}
	}()
}

// List returns a page of request logs. Source "db" reads PostgreSQL (accurate
// filtering + total), source "redis" (default when available) reads the capped
// Redis list.
func (s *Store) List(ctx context.Context, q dto.ListLogQuery) ([]model.RequestLog, int64, string, error) {
	source := q.Source
	if source == "" {
		if s.rdb != nil && s.cfg.RequestLogToRedis {
			source = "redis"
		} else {
			source = "db"
		}
	}

	if source == "redis" && s.rdb != nil {
		entries, total, err := s.listRedis(ctx, q)
		if err == nil {
			return entries, total, "redis", nil
		}
		slog.Warn("redis log read failed, falling back to database", "error", err)
	}

	entries, total, err := s.listDB(ctx, q)
	return entries, total, "db", err
}

func (s *Store) listRedis(ctx context.Context, q dto.ListLogQuery) ([]model.RequestLog, int64, error) {
	raw, err := s.rdb.LRange(ctx, RedisListKey, 0, s.cfg.RequestLogMaxEntries-1).Result()
	if err != nil {
		return nil, 0, err
	}

	all := make([]model.RequestLog, 0, len(raw))
	for _, item := range raw {
		var entry model.RequestLog
		if err := json.Unmarshal([]byte(item), &entry); err != nil {
			continue
		}
		if !matches(entry, q) {
			continue
		}
		all = append(all, entry)
	}

	total := int64(len(all))
	start := q.Offset()
	if start >= len(all) {
		return []model.RequestLog{}, total, nil
	}
	end := start + q.Limit
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], total, nil
}

func matches(entry model.RequestLog, q dto.ListLogQuery) bool {
	if q.Method != "" && !strings.EqualFold(entry.Method, q.Method) {
		return false
	}
	if q.Status != nil && entry.Status != *q.Status {
		return false
	}
	return true
}

func (s *Store) listDB(ctx context.Context, q dto.ListLogQuery) ([]model.RequestLog, int64, error) {
	if s.pool == nil {
		return []model.RequestLog{}, 0, nil
	}

	where := "WHERE 1 = 1"
	args := []any{}
	if q.Method != "" {
		args = append(args, q.Method)
		where += " AND method = $1"
	}
	if q.Status != nil {
		args = append(args, *q.Status)
		where += " AND status = $" + strconv.Itoa(len(args))
	}

	var total int64
	if err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM request_logs "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, q.Limit, q.Offset())
	listQ := `
		SELECT id, method, api_url, payload, executed_time_ms, status, response_code,
		       client_ip, user_agent, response_bytes, datetime
		FROM request_logs ` + where + `
		ORDER BY datetime DESC
		LIMIT $` + strconv.Itoa(len(args)-1) + ` OFFSET $` + strconv.Itoa(len(args))

	rows, err := s.pool.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	entries := make([]model.RequestLog, 0, q.Limit)
	for rows.Next() {
		var e model.RequestLog
		if err := rows.Scan(&e.ID, &e.Method, &e.URL, &e.Payload, &e.ExecutedTime,
			&e.Status, &e.ResponseCode, &e.ClientIP, &e.UserAgent, &e.ResponseBytes, &e.Datetime); err != nil {
			return nil, 0, err
		}
		entries = append(entries, e)
	}
	return entries, total, rows.Err()
}

// Close drains the queue and waits for the worker to finish.
func (s *Store) Close() {
	s.once.Do(func() {
		close(s.closed)
		close(s.queue)
	})
	s.wg.Wait()
	s.hub.close()
}
