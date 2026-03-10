package vedatrace

import (
	"context"
	"fmt"
	"maps"
	"time"
)

// Logger is the main VedaTrace logging instance. Create one with New or NewDev
// and keep it for the lifetime of your application.
type Logger struct {
	cfg   Config
	trans transport
	bat   *batcher // nil when using consoleTransport
	meta  LogMetadata
}

// New creates a Logger that sends logs to the VedaTrace HTTP ingest endpoint.
// cfg.APIKey and cfg.Service are required.
func New(cfg Config) (*Logger, error) {
	if cfg.APIKey == "" && !cfg.DisableHTTP {
		return nil, fmt.Errorf("vedatrace: APIKey is required (or set DisableHTTP: true for console-only mode)")
	}
	if cfg.Service == "" {
		return nil, fmt.Errorf("vedatrace: Service name is required")
	}
	cfg = cfg.withDefaults()

	var t transport
	var b *batcher

	if cfg.DisableHTTP {
		t = &consoleTransport{}
	} else {
		t = newHTTPTransport(cfg)
		b = newBatcher(cfg, t)
	}

	return &Logger{cfg: cfg, trans: t, bat: b}, nil
}

// NewDev creates a console-only Logger suitable for local development.
// No API key is required.
func NewDev(service string) *Logger {
	cfg := Config{Service: service, DisableHTTP: true}.withDefaults()
	return &Logger{cfg: cfg, trans: &consoleTransport{}}
}

// Child returns a new Logger that inherits this logger's config and merges
// meta into every log entry it emits.
func (l *Logger) Child(meta LogMetadata) *Logger {
	merged := make(LogMetadata, len(l.meta)+len(meta))
	maps.Copy(merged, l.meta)
	maps.Copy(merged, meta)
	return &Logger{cfg: l.cfg, trans: l.trans, bat: l.bat, meta: merged}
}

// Debug logs a message at debug level.
func (l *Logger) Debug(msg string, meta ...LogMetadata) {
	l.emit(LevelDebug, msg, nil, meta...)
}

// Info logs a message at info level.
func (l *Logger) Info(msg string, meta ...LogMetadata) {
	l.emit(LevelInfo, msg, nil, meta...)
}

// Warn logs a message at warn level.
func (l *Logger) Warn(msg string, meta ...LogMetadata) {
	l.emit(LevelWarn, msg, nil, meta...)
}

// Error logs a message at error level with an optional error value.
func (l *Logger) Error(msg string, err error, meta ...LogMetadata) {
	l.emit(LevelError, msg, CaptureError(err), meta...)
}

// Fatal logs a message at fatal level, flushes all pending logs, then panics.
func (l *Logger) Fatal(msg string, err error, meta ...LogMetadata) {
	l.emit(LevelFatal, msg, CaptureErrorWithStack(err), meta...)
	l.Flush()
	panic(fmt.Sprintf("vedatrace: fatal — %s", msg))
}

// Flush blocks until all buffered log entries have been delivered.
// It is a no-op when using console transport.
func (l *Logger) Flush() {
	if l.bat != nil {
		l.bat.flush()
	}
}

// Stop flushes and shuts down the background batcher goroutine.
// After Stop returns the Logger must not be used.
func (l *Logger) Stop() {
	if l.bat != nil {
		l.bat.Drain()
	}
}

// emit builds a LogEntry and hands it to the batcher (HTTP) or the transport
// directly (console).
func (l *Logger) emit(level Level, msg string, errInfo *ErrorInfo, extra ...LogMetadata) {
	merged := make(LogMetadata, len(l.meta))
	maps.Copy(merged, l.meta)
	for _, m := range extra {
		maps.Copy(merged, m)
	}
	if len(merged) == 0 {
		merged = nil
	}

	entry := LogEntry{
		Level:     level,
		Message:   msg,
		Service:   l.cfg.Service,
		Timestamp: time.Now().UTC(),
		Metadata:  merged,
		Error:     errInfo,
	}

	if l.bat != nil {
		l.bat.Enqueue(entry)
	} else {
		// Console transport: apply redaction inline, ignore error.
		entry.Metadata = redact(entry.Metadata, l.cfg.RedactFields)
		_ = l.trans.Send(context.TODO(), []LogEntry{entry})
	}
}
