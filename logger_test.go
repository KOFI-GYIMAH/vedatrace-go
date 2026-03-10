package vedatrace

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// captureTrans captures all entries it receives for assertion in tests.
type captureTrans struct {
	mu      sync.Mutex
	entries []LogEntry
}

func (c *captureTrans) Send(_ context.Context, entries []LogEntry) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = append(c.entries, entries...)
	return nil
}

func (c *captureTrans) all() []LogEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]LogEntry, len(c.entries))
	copy(out, c.entries)
	return out
}

// newTestLogger creates a Logger wired to a captureTrans without starting the
// batcher goroutine, making tests deterministic.
func newTestLogger(t *testing.T) (*Logger, *captureTrans) {
	t.Helper()
	ct := &captureTrans{}
	cfg := Config{
		Service:       "test-svc",
		BatchSize:     100,
		FlushInterval: time.Minute,
	}.withDefaults()
	l := &Logger{cfg: cfg, trans: ct} // no batcher → console-style direct send
	return l, ct
}

func TestNew_missingAPIKey(t *testing.T) {
	_, err := New(Config{Service: "svc"})
	if err == nil {
		t.Fatal("expected error for missing APIKey")
	}
}

func TestNew_missingService(t *testing.T) {
	_, err := New(Config{APIKey: "key"})
	if err == nil {
		t.Fatal("expected error for missing Service")
	}
}

func TestNew_disableHTTP(t *testing.T) {
	l, err := New(Config{Service: "svc", DisableHTTP: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	l.Info("hello")
	l.Stop()
}

func TestNewDev(t *testing.T) {
	l := NewDev("dev-svc")
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
	l.Info("dev message")
}

func TestLogger_levels(t *testing.T) {
	l, ct := newTestLogger(t)

	l.Debug("dbg")
	l.Info("info")
	l.Warn("warn")
	l.Error("err", errors.New("oops"))

	entries := ct.all()
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}
	levels := []Level{LevelDebug, LevelInfo, LevelWarn, LevelError}
	for i, e := range entries {
		if e.Level != levels[i] {
			t.Errorf("entry %d: got level %q, want %q", i, e.Level, levels[i])
		}
		if e.Service != "test-svc" {
			t.Errorf("entry %d: got service %q, want %q", i, e.Service, "test-svc")
		}
	}
}

func TestLogger_errorEntryHasErrorInfo(t *testing.T) {
	l, ct := newTestLogger(t)
	l.Error("bad thing", errors.New("disk full"))
	entries := ct.all()
	if len(entries) == 0 {
		t.Fatal("no entries captured")
	}
	e := entries[0]
	if e.Error == nil {
		t.Fatal("expected ErrorInfo to be set")
	}
	if e.Error.Message != "disk full" {
		t.Errorf("got %q, want %q", e.Error.Message, "disk full")
	}
}

func TestLogger_child(t *testing.T) {
	l, ct := newTestLogger(t)
	child := l.Child(LogMetadata{"request_id": "abc-123"})
	child.Info("from child")

	entries := ct.all()
	if len(entries) == 0 {
		t.Fatal("no entries")
	}
	if entries[0].Metadata["request_id"] != "abc-123" {
		t.Errorf("child metadata not propagated: %v", entries[0].Metadata)
	}
}

func TestLogger_childInheritsMeta(t *testing.T) {
	l, ct := newTestLogger(t)
	l.meta = LogMetadata{"env": "prod"}
	child := l.Child(LogMetadata{"user": "alice"})
	child.Info("hi")

	entries := ct.all()
	if entries[0].Metadata["env"] != "prod" {
		t.Error("parent meta not inherited")
	}
	if entries[0].Metadata["user"] != "alice" {
		t.Error("child meta not present")
	}
}

func TestLogger_metaMerge(t *testing.T) {
	l, ct := newTestLogger(t)
	l.Info("msg", LogMetadata{"k": "v"})
	entries := ct.all()
	if entries[0].Metadata["k"] != "v" {
		t.Error("per-call metadata not attached")
	}
}

func TestLogger_flush(t *testing.T) {
	l, _ := newTestLogger(t)
	// Flush on a logger without a batcher should not panic.
	l.Flush()
	l.Stop()
}

func TestLogger_timestamp(t *testing.T) {
	before := time.Now().UTC().Add(-time.Second)
	l, ct := newTestLogger(t)
	l.Info("ts test")
	after := time.Now().UTC().Add(time.Second)

	e := ct.all()[0]
	if e.Timestamp.Before(before) || e.Timestamp.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", e.Timestamp, before, after)
	}
}
