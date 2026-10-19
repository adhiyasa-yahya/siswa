// Package cache owns the Redis client used by the realtime request logger.
package cache

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/nugroho/siswa/backend/internal/config"
)

// Connect dials Redis and pings it. Redis is an optional dependency: when it is
// unreachable the caller keeps running and request logs fall back to PostgreSQL,
// so this returns (nil, err) instead of panicking.
func Connect(ctx context.Context, cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
	})

	var err error
	for attempt := 1; attempt <= 5; attempt++ {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err = client.Ping(pingCtx).Err()
		cancel()
		if err == nil {
			slog.Info("redis connected", "addr", cfg.RedisAddr, "db", cfg.RedisDB)
			return client, nil
		}

		slog.Warn("redis not ready, retrying", "attempt", attempt, "error", err)
		select {
		case <-ctx.Done():
			_ = client.Close()
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}

	_ = client.Close()
	return nil, err
}
