package vedatrace

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestHTTPTransport_success(t *testing.T) {
	srv := httptest.
		NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-API-Key") != "test-key" {
				t.Errorf("bad X-API-Key header: %s", r.Header.Get("X-API-Key"))
			}
			w.WriteHeader(http.StatusOK)
		}))
	defer srv.Close()

	cfg := Config{
		APIKey:     "test-key",
		Service:    "svc",
		Endpoint:   srv.URL,
		MaxRetries: 0,
		RetryDelay: time.Millisecond,
	}.withDefaults()
	cfg.Endpoint = srv.URL

	tr := newHTTPTransport(cfg)
	err := tr.Send(context.Background(), []LogEntry{{Level: LevelInfo, Message: "hi", Service: "svc", Timestamp: time.Now()}})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestHTTPTransport_retryOn5xx(t *testing.T) {
	var calls int32
	srv := httptest.
		NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			n := atomic.AddInt32(&calls, 1)
			if n < 3 {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}))
	defer srv.Close()

	cfg := Config{
		APIKey:     "k",
		Service:    "svc",
		Endpoint:   srv.URL,
		MaxRetries: 3,
		RetryDelay: time.Millisecond,
	}.withDefaults()
	cfg.Endpoint = srv.URL

	tr := newHTTPTransport(cfg)
	err := tr.Send(context.Background(), []LogEntry{{Level: LevelInfo, Message: "hi", Service: "svc", Timestamp: time.Now()}})
	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Errorf("expected 3 calls, got %d", atomic.LoadInt32(&calls))
	}
}

func TestHTTPTransport_exhaustRetries(t *testing.T) {
	srv := httptest.
		NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
	defer srv.Close()

	cfg := Config{
		APIKey:     "k",
		Service:    "svc",
		Endpoint:   srv.URL,
		MaxRetries: 2,
		RetryDelay: time.Millisecond,
	}.withDefaults()
	cfg.Endpoint = srv.URL

	tr := newHTTPTransport(cfg)
	err := tr.Send(context.Background(), []LogEntry{{Level: LevelInfo, Message: "hi", Service: "svc", Timestamp: time.Now()}})
	if err == nil {
		t.Fatal("expected error after exhausted retries")
	}
}
