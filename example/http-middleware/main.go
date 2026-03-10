package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	vedatrace "github.com/KOFI-GYIMAH/vedatrace-go"
)

func main() {
	logger, err := vedatrace.New(vedatrace.Config{
		APIKey:  "vt_your_api_key_here",
		Service: "http-example",
	})
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Stop()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from vedatrace-go!")
	})

	handler := loggingMiddleware(logger, mux)

	logger.Info("server listening", vedatrace.LogMetadata{"addr": ":8080"})
	if err := http.ListenAndServe(":8080", handler); err != nil {
		logger.Error("server error", err)
	}
}

// loggingMiddleware is an example of how to wire vedatrace into net/http
// without any middleware package.
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
