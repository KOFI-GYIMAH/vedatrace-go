# vedatrace-go — Implementation Plan

> Go SDK for the VedaTrace logging platform, mirroring the feature-set of
> [vedatrace-npm](https://github.com/kurtiz/vedatrace-npm).
>
> **Module:** `github.com/KOFI-GYIMAH/vedatrace-go`
> **Target Go version:** 1.21+

---

## Current state

| File | Status | Notes |
|------|--------|-------|
| `go.mod` | ✅ done | module path set |
| `types.go` | ✅ done | `Level`, `LogEntry`, `LogMetadata`, `ErrorInfo`, `IngestPayload` |
| `config.go` | ✅ done | `Config` struct + `withDefaults()` |

---

## Step 1 — Core logger (`logger.go`)

Create the central `Logger` type and the public constructor functions.

```
Logger struct {
    cfg       Config
    batcher   *batcher       // nil when DisableHTTP + no batcher needed
    transport transport      // interface (http or console)
    metadata  LogMetadata    // inherited by child loggers
}
```

- `New(cfg Config) *Logger` — validates `cfg.APIKey` (required unless `DisableHTTP`),
  applies `cfg.withDefaults()`, wires up transport and batcher, returns `*Logger`.
- `NewDev(service string) *Logger` — convenience constructor: sets `DisableHTTP = true`,
  console-only output, no API key required.
- Log-level methods, each building a `LogEntry` and handing it to the batcher/transport:
  - `Debug(msg string, meta ...LogMetadata)`
  - `Info(msg string, meta ...LogMetadata)`
  - `Warn(msg string, meta ...LogMetadata)`
  - `Error(msg string, err error, meta ...LogMetadata)`
  - `Fatal(msg string, err error, meta ...LogMetadata)` — calls `log.Fatal` after flushing
- `Child(meta LogMetadata) *Logger` — returns a new `*Logger` that inherits the parent's
  config and merges `meta` into every log entry it emits.
- `Flush() error` — blocks until the batcher drains.
- `Stop()` — stops the background flush timer and performs a final flush.

---

## Step 2 — Transport interface + HTTP transport (`transport.go`)

Define the internal `transport` interface:

```go
type transport interface {
    Send(ctx context.Context, entries []LogEntry) error
}
```

Implement `httpTransport`:

- Serialises `[]LogEntry` into `IngestPayload` JSON.
- Sets `Authorization: Bearer <apiKey>` and `Content-Type: application/json`.
- Executes up to `cfg.MaxRetries + 1` attempts with linear back-off
  (`attempt * cfg.RetryDelay`).
- Returns the last error after all attempts are exhausted.

---

## Step 3 — Console transport (`console.go`)

Implement `consoleTransport` (used when `DisableHTTP = true` or no API key):

- Writes each entry as a human-readable, colour-coded line to `os.Stderr`
  (or `os.Stdout` for debug/info).
- Format: `[LEVEL] timestamp service — message {metadata}`.
- No batching delay; each entry is printed immediately.

---

## Step 4 — Batcher (`batcher.go`)

Implement the background batcher used by `httpTransport`:

```
batcher struct {
    mu       sync.Mutex
    buf      []LogEntry
    cfg      Config
    trans    transport
    ticker   *time.Ticker
    quit     chan struct{}
    wg       sync.WaitGroup
}
```

- `newBatcher(cfg Config, t transport) *batcher` — starts the background goroutine.
- `Enqueue(e LogEntry)` — appends to the buffer; flushes immediately if
  `len(buf) >= cfg.BatchSize`.
- `flush()` (internal) — drains the buffer, calls `transport.Send`, invokes
  `cfg.OnSuccess` or `cfg.OnError` callbacks.
- Background goroutine wakes on `ticker.C` and calls `flush()`.
- `Drain() error` — stops the ticker, waits for the goroutine, flushes remaining entries.

---

## Step 5 — Field redaction (`redact.go`)

Implement redaction for `cfg.RedactFields`:

- Each element is a dot-notation path, e.g. `"card.cvv"` or `"password"`.
- Walk `LogMetadata` (which is `map[string]any`) and replace matching leaf values
  with `"[REDACTED]"`.
- Apply redaction in `batcher.flush()` / `consoleTransport.Send()` before any
  serialisation.

---

## Step 6 — Error capture helper (`errors.go`)

Expose a helper for building `*ErrorInfo` from a Go `error`:

```go
func CaptureError(err error) *ErrorInfo
func CaptureErrorWithStack(err error) *ErrorInfo  // includes runtime.Stack output
```

These are used internally by `Error()` / `Fatal()` and can also be used by callers
to attach error info to metadata manually.

---

## Step 7 — Middleware: `net/http` (`middleware/nethttp/middleware.go`)

Separate sub-package `middleware/nethttp`:

- `Middleware(logger *vedatrace.Logger) func(http.Handler) http.Handler`
- Logs every request at `Info` level with metadata:
  `method`, `path`, `status`, `latency_ms`, `remote_addr`, `user_agent`.
- Captures panics, logs them at `Fatal`, re-panics.

---

## Step 8 — Middleware: Gin (`middleware/gin/middleware.go`)

Separate sub-package `middleware/gin`:

- `Middleware(logger *vedatrace.Logger) gin.HandlerFunc`
- Same fields as the `net/http` middleware; uses `gin.Context` for status code
  capture.

---

## Step 9 — Middleware: Echo (`middleware/echo/middleware.go`)

Separate sub-package `middleware/echo`:

- `Middleware(logger *vedatrace.Logger) echo.MiddlewareFunc`

---

## Step 10 — Middleware: Chi (`middleware/chi/middleware.go`)

Separate sub-package `middleware/chi`:

- `Middleware(logger *vedatrace.Logger) func(http.Handler) http.Handler`
- Wraps `chi/middleware.WrapResponseWriter` to capture status.

---

## Step 11 — Tests

| File | What to cover |
|------|---------------|
| `logger_test.go` | `New`, `NewDev`, level methods, `Child`, `Flush`, `Stop` |
| `batcher_test.go` | batch-size trigger, flush-interval trigger, drain, callbacks |
| `transport_test.go` | HTTP send success, retry on 5xx, retry exhaustion |
| `redact_test.go` | shallow and nested dot-notation redaction |
| `errors_test.go` | `CaptureError`, `CaptureErrorWithStack` |
| `middleware/nethttp/middleware_test.go` | request/response logging, panic capture |

Use `httptest.NewServer` for transport tests; no real network calls in tests.

---

## Step 12 — Documentation & examples

- `README.md` — quick-start, configuration table, middleware usage.
- `example/basic/main.go` — `New` + all five log levels + `Stop`.
- `example/dev/main.go` — `NewDev` console logger.
- `example/http-middleware/main.go` — stdlib HTTP server with middleware.

---

## Directory layout (final)

```
vedatrace-go/
├── go.mod
├── config.go          ✅ done
├── types.go           ✅ done
├── logger.go          step 1
├── transport.go       step 2
├── console.go         step 3
├── batcher.go         step 4
├── redact.go          step 5
├── errors.go          step 6
├── logger_test.go     step 11
├── batcher_test.go    step 11
├── transport_test.go  step 11
├── redact_test.go     step 11
├── errors_test.go     step 11
├── middleware/
│   ├── nethttp/
│   │   ├── middleware.go
│   │   └── middleware_test.go
│   ├── gin/
│   │   └── middleware.go
│   ├── echo/
│   │   └── middleware.go
│   └── chi/
│       └── middleware.go
├── example/
│   ├── basic/main.go
│   ├── dev/main.go
│   └── http-middleware/main.go
└── README.md
```

---

## Implementation order

1. `logger.go` + `console.go` — gets a working `NewDev` logger end-to-end.
2. `transport.go` — HTTP send without retries first, then add retry loop.
3. `batcher.go` — wire batcher into `New`.
4. `redact.go` — plug into batcher flush.
5. `errors.go` — plug into `Error`/`Fatal`.
6. Tests for core (steps 1–5).
7. Middleware packages (steps 7–10), each with its own test.
8. Examples + README.
