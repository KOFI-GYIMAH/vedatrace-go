// Package chi provides VedaTrace request-logging middleware for the Chi
// router (github.com/go-chi/chi).
package chi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	vedatrace "github.com/KOFI-GYIMAH/vedatrace-go"
)

// Middleware returns an http.Handler middleware that logs each request using
// the provided VedaTrace logger.
func Middleware(logger *vedatrace.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			next.ServeHTTP(ww, r)

			logger.Info("http request", vedatrace.LogMetadata{
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      ww.Status(),
				"latency_ms":  time.Since(start).Milliseconds(),
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
			})
		})
	}
}
