package handler

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nugroho/siswa/backend/internal/dto"
	"github.com/nugroho/siswa/backend/internal/errs"
	"github.com/nugroho/siswa/backend/internal/logstore"
	"github.com/nugroho/siswa/backend/internal/response"
)

// LogHandler serves the realtime request log endpoints.
type LogHandler struct {
	store *logstore.Store
}

func NewLogHandler(store *logstore.Store) *LogHandler {
	return &LogHandler{store: store}
}

// List handles GET /api/v1/logs.
func (h *LogHandler) List(c *gin.Context) {
	q := dto.ListLogQuery{
		Page:   queryInt(c, "page", 1),
		Limit:  queryInt(c, "limit", 25),
		Method: c.Query("method"),
		Source: c.Query("source"),
	}

	if raw := strings.TrimSpace(c.Query("status")); raw != "" {
		status, err := strconv.Atoi(raw)
		if err != nil {
			response.Error(c, errs.Validation("Invalid query parameter",
				errs.FieldError{Field: "status", Reason: "Must be an integer HTTP status"}))
			return
		}
		q.Status = &status
	}

	q.Sanitize()

	entries, total, source, err := h.store.List(c.Request.Context(), q)
	if err != nil {
		response.Error(c, errs.Cache(err))
		return
	}

	c.Writer.Header().Set("X-Log-Source", source)
	response.OKWithMeta(c, "Request logs retrieved successfully", entries,
		response.NewMeta(q.Page, q.Limit, total))
}

// Stream handles GET /api/v1/logs/stream: a Server-Sent Events feed pushing
// every request log entry the moment the request that produced it finishes.
func (h *LogHandler) Stream(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(200)

	ctx := c.Request.Context()
	events, unsubscribe := h.store.Subscribe(ctx)
	defer unsubscribe()

	_, _ = c.Writer.WriteString("event: connected\ndata: {\"source\":\"siswa\"}\n\n")
	c.Writer.Flush()

	keepAlive := time.NewTicker(20 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-keepAlive.C:
			if _, err := c.Writer.WriteString(": keep-alive\n\n"); err != nil {
				return
			}
			c.Writer.Flush()
		case entry, ok := <-events:
			if !ok {
				return
			}
			payload, err := json.Marshal(entry)
			if err != nil {
				continue
			}
			if _, err := c.Writer.WriteString("event: request\ndata: " + string(payload) + "\n\n"); err != nil {
				return
			}
			c.Writer.Flush()
		}
	}
}
