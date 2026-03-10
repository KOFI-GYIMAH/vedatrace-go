package main

import (
	"errors"
	"log"

	vedatrace "github.com/KOFI-GYIMAH/vedatrace-go"
)

func main() {
	logger, err := vedatrace.New(vedatrace.Config{
		APIKey:  "vt_your_api_key_here",
		Service: "my-service",
		OnError: func(err error, batch []vedatrace.LogEntry) {
			log.Printf("vedatrace delivery error: %v (%d entries dropped)\n", err, len(batch))
		},
	})
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Stop()

	logger.Debug("application starting", vedatrace.LogMetadata{"version": "1.0.0"})
	logger.Info("user logged in", vedatrace.LogMetadata{"user_id": "u-42"})
	logger.Warn("rate limit approaching", vedatrace.LogMetadata{"usage_pct": 85})
	logger.Error("database query failed", errors.New("connection timeout"), vedatrace.LogMetadata{"query": "SELECT *"})

	// Child logger with shared context.
	requestLog := logger.Child(vedatrace.LogMetadata{"request_id": "req-abc"})
	requestLog.Info("processing request")
	requestLog.Info("request complete", vedatrace.LogMetadata{"status": 200})
}
