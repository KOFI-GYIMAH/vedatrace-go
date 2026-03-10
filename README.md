# vedatrace-go

> Official Go SDK for the [VedaTrace](https://vedatrace.dev) logging platform.

Type-safe, lightweight, and developer-friendly structured logging for Go applications — with background batching, automatic retries, and field redaction.

**Zero external dependencies. Requires Go 1.21+.**

---

## Installation

```bash
go get github.com/KOFI-GYIMAH/vedatrace-go
```

## Quick start

```go
package main

import (
    "log"
    vedatrace "github.com/KOFI-GYIMAH/vedatrace-go"
)

func main() {
    logger, err := vedatrace.New(vedatrace.Config{
        APIKey:  "vt_your_api_key",
        Service: "my-service",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer logger.Stop()

    logger.Info("server started", vedatrace.LogMetadata{"port": 8080})
    logger.Warn("high memory usage", vedatrace.LogMetadata{"used_mb": 512})
}
```

## Development mode (no API key)

```go
logger := vedatrace.NewDev("my-service")
logger.Debug("only printed to stderr")
```

## Log levels

| Method | Level |
|--------|-------|
| `logger.Debug(msg, meta...)` | debug |
| `logger.Info(msg, meta...)` | info |
| `logger.Warn(msg, meta...)` | warn |
| `logger.Error(msg, err, meta...)` | error |
| `logger.Fatal(msg, err, meta...)` | fatal — flushes then panics |

## Child loggers

```go
reqLog := logger.Child(vedatrace.LogMetadata{"request_id": "abc-123"})
reqLog.Info("processing") // automatically includes request_id
```

## Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `APIKey` | `string` | — | **Required** unless `DisableHTTP` is true |
| `Service` | `string` | — | **Required**. Attached to every log entry |
| `Endpoint` | `string` | `https://ingest.vedatrace.dev/v1/logs` | Ingest URL |
| `BatchSize` | `int` | `100` | Flush when buffer reaches this size |
| `FlushInterval` | `time.Duration` | `5s` | Flush at least this often |
| `MaxRetries` | `int` | `3` | Extra attempts after a failed send |
| `RetryDelay` | `time.Duration` | `1s` | Base delay between retries (linear back-off) |
| `RedactFields` | `[]string` | — | Dot-notation paths to redact, e.g. `"card.cvv"` |
| `OnSuccess` | `func([]LogEntry)` | — | Called after each successful batch delivery |
| `OnError` | `func(error, []LogEntry)` | — | Called after all retry attempts are exhausted |
| `DisableHTTP` | `bool` | `false` | Force console-only output |

## Field redaction

```go
logger, _ := vedatrace.New(vedatrace.Config{
    APIKey:       "vt_key",
    Service:      "payments",
    RedactFields: []string{"password", "card.cvv"},
})
logger.Info("checkout", vedatrace.LogMetadata{
    "password": "s3cr3t",      // → "[REDACTED]"
    "card": map[string]any{
        "cvv": "123",          // → "[REDACTED]"
        "last4": "4242",       // unchanged
    },
})
```

## HTTP request logging

The SDK ships no middleware — wire it up directly in a few lines:

### net/http

```go
func loggingMiddleware(logger *vedatrace.Logger, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        rw := &statusWriter{ResponseWriter: w}
        next.ServeHTTP(rw, r)
        logger.Info("http request", vedatrace.LogMetadata{
            "method":     r.Method,
            "path":       r.URL.Path,
            "status":     rw.status,
            "latency_ms": time.Since(start).Milliseconds(),
        })
    })
}

type statusWriter struct {
    http.ResponseWriter
    status int
}

func (sw *statusWriter) WriteHeader(code int) {
    sw.status = code
    sw.ResponseWriter.WriteHeader(code)
}
```

### Gin

```go
r.Use(func(c *gin.Context) {
    start := time.Now()
    c.Next()
    logger.Info("http request", vedatrace.LogMetadata{
        "method":     c.Request.Method,
        "path":       c.FullPath(),
        "status":     c.Writer.Status(),
        "latency_ms": time.Since(start).Milliseconds(),
    })
})
```

### Echo

```go
e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        start := time.Now()
        err := next(c)
        logger.Info("http request", vedatrace.LogMetadata{
            "method":     c.Request().Method,
            "path":       c.Request().URL.Path,
            "status":     c.Response().Status,
            "latency_ms": time.Since(start).Milliseconds(),
        })
        return err
    }
})
```

### Chi

```go
r.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        rw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
        next.ServeHTTP(rw, r)
        logger.Info("http request", vedatrace.LogMetadata{
            "method":     r.Method,
            "path":       r.URL.Path,
            "status":     rw.Status(),
            "latency_ms": time.Since(start).Milliseconds(),
        })
    })
})
```

## Flushing & shutdown

```go
logger.Flush() // block until current buffer is delivered
logger.Stop()  // flush + stop background goroutine (call on shutdown)
```

## Error capture helpers

```go
info := vedatrace.CaptureError(err)              // message + type
info := vedatrace.CaptureErrorWithStack(err)     // + goroutine stack trace
```

## License

MIT
