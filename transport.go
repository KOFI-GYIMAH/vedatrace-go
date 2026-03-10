package vedatrace

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// transport is the internal interface for delivering log batches.
type transport interface {
	Send(ctx context.Context, entries []LogEntry) error
}

// httpTransport delivers log batches to the VedaTrace ingest API over HTTP.
type httpTransport struct {
	cfg    Config
	client *http.Client
}

func newHTTPTransport(cfg Config) *httpTransport {
	return &httpTransport{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *httpTransport) Send(ctx context.Context, entries []LogEntry) error {
	payload := IngestPayload{Logs: entries}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("vedatrace: marshal payload: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= t.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * t.cfg.RetryDelay):
			}
		}

		req, err := http.
			NewRequestWithContext(
				ctx, http.MethodPost, t.cfg.Endpoint, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("vedatrace: build request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+t.cfg.APIKey)

		resp, err := t.client.Do(req)
		if err != nil {
			lastErr = fmt.
				Errorf("vedatrace: http request (attempt %d): %w", attempt+1, err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		lastErr = fmt.Errorf("vedatrace: unexpected status %d (attempt %d)", resp.StatusCode, attempt+1)
	}
	return lastErr
}
