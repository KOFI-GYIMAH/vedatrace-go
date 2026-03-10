package vedatrace

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeTrans records every batch it receives.
type fakeTrans struct {
	mu      sync.Mutex
	batches [][]LogEntry
	err     error
}

func (f *fakeTrans) Send(_ context.Context, entries []LogEntry) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := make([]LogEntry, len(entries))
	copy(cp, entries)
	f.batches = append(f.batches, cp)
	return f.err
}

func (f *fakeTrans) total() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, b := range f.batches {
		n += len(b)
	}
	return n
}

func entry(msg string) LogEntry {
	return LogEntry{Level: LevelInfo, Message: msg, Service: "svc", Timestamp: time.Now()}
}

func TestBatcher_batchSizeTrigger(t *testing.T) {
	ft := &fakeTrans{}
	cfg := Config{
		Service:       "svc",
		BatchSize:     3,
		FlushInterval: time.Minute, // won't fire
		MaxRetries:    0,
	}.withDefaults()
	cfg.BatchSize = 3
	cfg.FlushInterval = time.Minute

	b := newBatcher(cfg, ft)
	b.Enqueue(entry("a"))
	b.Enqueue(entry("b"))
	b.Enqueue(entry("c")) // triggers flush
	b.Drain()

	if ft.total() < 3 {
		t.Errorf("expected >=3 delivered entries, got %d", ft.total())
	}
}

func TestBatcher_flushIntervalTrigger(t *testing.T) {
	ft := &fakeTrans{}
	cfg := Config{
		Service:       "svc",
		BatchSize:     1000,
		FlushInterval: 20 * time.Millisecond,
		MaxRetries:    0,
	}.withDefaults()
	cfg.BatchSize = 1000
	cfg.FlushInterval = 20 * time.Millisecond

	b := newBatcher(cfg, ft)
	b.Enqueue(entry("x"))

	// wait for the ticker to fire
	time.Sleep(60 * time.Millisecond)
	b.Drain()

	if ft.total() < 1 {
		t.Error("expected at least 1 entry flushed by interval")
	}
}

func TestBatcher_onSuccessCallback(t *testing.T) {
	ft := &fakeTrans{}
	var called atomic.Bool
	cfg := Config{
		Service:       "svc",
		BatchSize:     1,
		FlushInterval: time.Minute,
		MaxRetries:    0,
		OnSuccess: func(batch []LogEntry) {
			called.Store(true)
		},
	}.withDefaults()
	cfg.BatchSize = 1
	cfg.FlushInterval = time.Minute

	b := newBatcher(cfg, ft)
	b.Enqueue(entry("ok"))
	b.Drain()

	if !called.Load() {
		t.Error("OnSuccess was not called")
	}
}

func TestBatcher_onErrorCallback(t *testing.T) {
	ft := &fakeTrans{err: context.DeadlineExceeded}
	var errCalled atomic.Bool
	cfg := Config{
		Service:       "svc",
		BatchSize:     1,
		FlushInterval: time.Minute,
		MaxRetries:    0,
		OnError: func(err error, batch []LogEntry) {
			errCalled.Store(true)
		},
	}.withDefaults()
	cfg.BatchSize = 1
	cfg.FlushInterval = time.Minute

	b := newBatcher(cfg, ft)
	b.Enqueue(entry("fail"))
	b.Drain()

	if !errCalled.Load() {
		t.Error("OnError was not called")
	}
}
