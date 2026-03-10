// Package nethttp provides VedaTrace request-logging middleware for the
// standard library net/http package.
package nethttp

import (
	"fmt"
	"net/http"
	"time"

	vedatrace "github.com/KOFI-GYIMAH/vedatrace-go"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) statusCode() int {
	if rw.status == 0 {
		return http.StatusOK
	}
	return rw.status
}

// Middleware returns an http.Handler middleware that logs each request using
// the provided VedaTrace logger. Panics are caught, logged at Fatal level, and
// then re-panicked.
func Middleware(logger *vedatrace.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &responseWriter{ResponseWriter: w}
			start := time.Now()

			defer func() {
				if p := recover(); p != nil {
					logger.Fatal(
						fmt.Sprintf("panic: %v", p),
						nil,
						vedatrace.LogMetadata{
							"method":      r.Method,
							"path":        r.URL.Path,
							"remote_addr": r.RemoteAddr,
						},
					)
					panic(p) // re-panic after logging
				}
			}()

			next.ServeHTTP(rw, r)

			logger.Info("http request", vedatrace.LogMetadata{
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      rw.statusCode(),
				"latency_ms":  time.Since(start).Milliseconds(),
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
			})
		})
	}
}
