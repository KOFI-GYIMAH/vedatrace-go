package nethttp_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	vedatrace "github.com/KOFI-GYIMAH/vedatrace-go"
	vnethttp "github.com/KOFI-GYIMAH/vedatrace-go/middleware/nethttp"
)

func TestMiddleware_logsRequest(t *testing.T) {
	logger := vedatrace.NewDev("test-svc")

	handler := vnethttp.Middleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestMiddleware_captures404(t *testing.T) {
	logger := vedatrace.NewDev("test-svc")

	handler := vnethttp.Middleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestMiddleware_panicRecovery(t *testing.T) {
	logger := vedatrace.NewDev("test-svc")

	handler := vnethttp.Middleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic to be re-panicked")
		}
	}()

	handler.ServeHTTP(rr, req)
}
