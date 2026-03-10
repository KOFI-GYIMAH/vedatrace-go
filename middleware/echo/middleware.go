// Package echo provides VedaTrace request-logging middleware for the Echo
// web framework (github.com/labstack/echo).
package echo

import (
	"time"

	"github.com/labstack/echo/v4"

	vedatrace "github.com/KOFI-GYIMAH/vedatrace-go"
)

// Middleware returns an echo.MiddlewareFunc that logs each request using the
// provided VedaTrace logger.
func Middleware(logger *vedatrace.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)

			status := c.Response().Status
			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					status = he.Code
				}
			}

			logger.Info("http request", vedatrace.LogMetadata{
				"method":      c.Request().Method,
				"path":        c.Request().URL.Path,
				"status":      status,
				"latency_ms":  time.Since(start).Milliseconds(),
				"remote_addr": c.RealIP(),
				"user_agent":  c.Request().UserAgent(),
			})
			return err
		}
	}
}
