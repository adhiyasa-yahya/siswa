package middleware

import (
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/nugroho/siswa/backend/internal/errs"
	"github.com/nugroho/siswa/backend/internal/response"
)

// Recovery turns a panic into the standard error envelope (code 2003) instead of
// an empty 500, and keeps the stack trace on the server side only.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("panic recovered",
					"request_id", requestIDOf(c),
					"path", c.Request.URL.Path,
					"panic", fmt.Sprint(recovered),
					"stack", string(debug.Stack()),
				)
				response.Error(c, errs.Internal(fmt.Errorf("panic: %v", recovered)))
			}
		}()
		c.Next()
	}
}
