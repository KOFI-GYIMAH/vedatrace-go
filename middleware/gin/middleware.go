// Package gin provides VedaTrace request-logging middleware for the Gin
// web framework (github.com/gin-gonic/gin).
package gin

import (
	"time"

	"github.com/gin-gonic/gin"

	vedatrace "github.com/KOFI-GYIMAH/vedatrace-go"
)

// Middleware returns a gin.HandlerFunc that logs each request using the
// provided VedaTrace logger.
func Middleware(logger *vedatrace.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		logger.Info("http request", vedatrace.LogMetadata{
			"method":      c.Request.Method,
			"path":        c.FullPath(),
			"status":      c.Writer.Status(),
			"latency_ms":  time.Since(start).Milliseconds(),
			"remote_addr": c.ClientIP(),
			"user_agent":  c.Request.UserAgent(),
		})
	}
}
