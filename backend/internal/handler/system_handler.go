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

// ErrorCodes handles GET /api/v1/error-codes: the full list of business codes.
func (h *SystemHandler) ErrorCodes(c *gin.Context) {
	response.OK(c, "Error codes retrieved successfully", errs.Catalogue)
}
