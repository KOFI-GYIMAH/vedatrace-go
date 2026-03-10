package vedatrace

import (
	"context"
	"sync"
	"time"
)

// batcher accumulates LogEntry values and delivers them to the underlying
// transport in batches, either when the buffer reaches BatchSize or when the
// flush ticker fires.
type batcher struct {
	mu    sync.Mutex
	buf   []LogEntry
	cfg   Config
	trans transport
	quit  chan struct{}
	done  chan struct{}
}

func newBatcher(cfg Config, t transport) *batcher {
	b := &batcher{
		cfg:   cfg,
		trans: t,
		quit:  make(chan struct{}),
		done:  make(chan struct{}),
	}
	go b.run()
	return b
}

// Enqueue adds an entry to the buffer. If the buffer reaches BatchSize the
// batch is flushed immediately.
func (b *batcher) Enqueue(e LogEntry) {
	b.mu.Lock()
	b.buf = append(b.buf, e)
	full := len(b.buf) >= b.cfg.BatchSize
	b.mu.Unlock()

	if full {
		b.flush()
	}
}

// Drain stops the background goroutine and flushes any remaining entries.
func (b *batcher) Drain() {
	close(b.quit)
	<-b.done
	b.flush()
}

// run is the background goroutine that flushes on the configured interval.
func (b *batcher) run() {
	defer close(b.done)
	ticker := time.NewTicker(b.cfg.FlushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			b.flush()
		case <-b.quit:
			return
		}
	}
}

// flush drains the buffer and sends to the transport, invoking callbacks.
func (b *batcher) flush() {
	b.mu.Lock()
	if len(b.buf) == 0 {
		b.mu.Unlock()
		return
	}
	batch := make([]LogEntry, len(b.buf))
	copy(batch, b.buf)
	b.buf = b.buf[:0]
	b.mu.Unlock()

	// Apply redaction per entry.
	redacted := make([]LogEntry, len(batch))
	for i, e := range batch {
		e.Metadata = redact(e.Metadata, b.cfg.RedactFields)
		redacted[i] = e
	}

	err := b.trans.Send(context.Background(), redacted)
	if err != nil {
		if b.cfg.OnError != nil {
			b.cfg.OnError(err, redacted)
		}
		return
	}
	if b.cfg.OnSuccess != nil {
		b.cfg.OnSuccess(redacted)
	}
}
