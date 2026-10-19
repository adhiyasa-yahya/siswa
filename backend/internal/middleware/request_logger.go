// Package middleware holds the cross cutting HTTP concerns: request id,
// realtime request logging, panic recovery and CORS.
package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/nugroho/siswa/backend/internal/config"
	"github.com/nugroho/siswa/backend/internal/logstore"
	"github.com/nugroho/siswa/backend/internal/model"
)

// RequestIDHeader is echoed back on every response for correlation.
const RequestIDHeader = "X-Request-ID"

// ContextRequestID is the gin context key holding the request id.
const ContextRequestID = "request_id"

// RequestID assigns (or reuses) a request id and exposes it to handlers and
// response headers.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.GetHeader(RequestIDHeader))
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(ContextRequestID, id)
		c.Writer.Header().Set(RequestIDHeader, id)
		c.Next()
	}
}

// responseRecorder counts the bytes written so the log can report payload size.
type responseRecorder struct {
	gin.ResponseWriter
	bytes int
}

func (w *responseRecorder) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

func (w *responseRecorder) WriteString(s string) (int, error) {
	n, err := w.ResponseWriter.WriteString(s)
	w.bytes += n
	return n, err
}

// RequestLogger records the full lifecycle of every request:
// method | api url | payload | executed time | status | datetime.
//
// The entry is broadcast to SSE subscribers immediately and persisted
// asynchronously to Redis and PostgreSQL by the log store.
func RequestLogger(cfg *config.Config, store *logstore.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.RequestLogEnabled || isSkipped(c.Request.URL.Path) {
			c.Next()
			return
		}

		start := time.Now()
		payload := capturePayload(c, cfg.RequestLogPayloadLimit)

		recorder := &responseRecorder{ResponseWriter: c.Writer}
		c.Writer = recorder

		c.Next()

		entry := model.RequestLog{
			ID:            requestIDOf(c),
			Method:        c.Request.Method,
			URL:           fullURL(c),
			Payload:       payload,
			ExecutedTime:  float64(time.Since(start).Microseconds()) / 1000.0,
			Status:        c.Writer.Status(),
			ResponseCode:  c.GetInt("response_code"),
			ClientIP:      c.ClientIP(),
			UserAgent:     c.Request.UserAgent(),
			ResponseBytes: recorder.bytes,
			Datetime:      start.UTC(),
		}

		store.Broadcast(entry)
		store.Record(entry)
	}
}

// isSkipped keeps the log free of noise and avoids a feedback loop where the
// log stream logs its own traffic.
func isSkipped(path string) bool {
	switch {
	case path == "/health", path == "/favicon.ico":
		return true
	case strings.HasPrefix(path, "/swagger"):
		return true
	case strings.HasPrefix(path, "/api/v1/logs"):
		return true
	case path == "/api/v1/status":
		// Polled by the frontend gate every few seconds; keeps the log readable.
		return true
	default:
		return false
	}
}

func requestIDOf(c *gin.Context) string {
	if id, ok := c.Get(ContextRequestID); ok {
		if s, ok := id.(string); ok && s != "" {
			return s
		}
	}
	return uuid.NewString()
}

func fullURL(c *gin.Context) string {
	if c.Request.URL.RawQuery != "" {
		return c.Request.URL.Path + "?" + c.Request.URL.RawQuery
	}
	return c.Request.URL.Path
}

// capturePayload reads the request body (restoring it for the handler), trims it
// to the configured limit and never records multipart uploads.
func capturePayload(c *gin.Context, limit int) string {
	if c.Request.Body == nil || c.Request.Method == http.MethodGet || c.Request.Method == http.MethodDelete {
		return ""
	}
	if strings.HasPrefix(c.ContentType(), "multipart/form-data") {
		return "[multipart/form-data omitted]"
	}

	body, err := io.ReadAll(io.LimitReader(c.Request.Body, int64(limit)+1))
	if err != nil {
		return ""
	}

	// Anything beyond the limit is not buffered, so put back what was read and
	// let the handler stream the rest.
	c.Request.Body = io.NopCloser(io.MultiReader(bytes.NewReader(body), c.Request.Body))

	if len(body) > limit {
		return string(body[:limit]) + "...[truncated]"
	}
	return string(body)
}
