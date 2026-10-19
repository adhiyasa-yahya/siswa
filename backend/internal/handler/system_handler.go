package handler

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/nugroho/siswa/backend/internal/errs"
	"github.com/nugroho/siswa/backend/internal/response"
)

// SystemHandler serves the health check and the error code catalogue.
type SystemHandler struct {
	pool      *pgxpool.Pool
	rdb       *redis.Client
	appName   string
	version   string
	startedAt time.Time
}

func NewSystemHandler(pool *pgxpool.Pool, rdb *redis.Client, appName, version string) *SystemHandler {
	return &SystemHandler{
		pool:      pool,
		rdb:       rdb,
		appName:   appName,
		version:   version,
		startedAt: time.Now(),
	}
}

// Health handles GET /health.
func (h *SystemHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	postgres := "up"
	if h.pool == nil {
		postgres = "disabled"
	} else if err := h.pool.Ping(ctx); err != nil {
		postgres = "down"
	}

	redisStatus := "disabled"
	if h.rdb != nil {
		redisStatus = "up"
		if err := h.rdb.Ping(ctx).Err(); err != nil {
			redisStatus = "down"
		}
	}

	response.OK(c, "Service is healthy", gin.H{
		"app":            h.appName,
		"version":        h.version,
		"uptime_seconds": int(time.Since(h.startedAt).Seconds()),
		"postgres":       postgres,
		"redis":          redisStatus,
		"time":           time.Now().UTC().Format(time.RFC3339),
	})
}

// Status handles GET /api/v1/status.
//
// This is the readiness probe the frontend gate polls: it is deliberately cheap
// (two pings with a short timeout, no queries) and it is excluded from the
// request log so the polling does not flood the realtime stream.
//
// PostgreSQL is required for the app to be usable, so a database that does not
// answer returns **503 / code 2001** and the frontend blocks the UI. Redis only
// backs the realtime log, so losing it degrades without blocking.
func (h *SystemHandler) Status(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 1500*time.Millisecond)
	defer cancel()

	postgres := "up"
	if h.pool == nil {
		postgres = "down"
	} else if err := h.pool.Ping(ctx); err != nil {
		postgres = "down"
	}

	redisStatus := "disabled"
	if h.rdb != nil {
		redisStatus = "up"
		if err := h.rdb.Ping(ctx).Err(); err != nil {
			redisStatus = "down"
		}
	}

	payload := gin.H{
		"app":            h.appName,
		"version":        h.version,
		"status":         "ok",
		"usable":         true,
		"postgres":       postgres,
		"redis":          redisStatus,
		"uptime_seconds": int(time.Since(h.startedAt).Seconds()),
		"server_time":    time.Now().UTC().Format(time.RFC3339),
	}

	if postgres != "up" {
		payload["status"] = "down"
		payload["usable"] = false
		response.Error(c, errs.Database(nil).WithFields(
			errs.FieldError{Field: "postgres", Reason: "Database is not reachable"},
		))
		return
	}

	if redisStatus == "down" {
		payload["status"] = "degraded"
	}

	response.OK(c, "Backend is reachable", payload)
}

// ErrorCodes handles GET /api/v1/error-codes: the full list of business codes.
func (h *SystemHandler) ErrorCodes(c *gin.Context) {
	response.OK(c, "Error codes retrieved successfully", errs.Catalogue)
}
