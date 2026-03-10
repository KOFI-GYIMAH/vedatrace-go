package main

import (
	"errors"

	vedatrace "github.com/KOFI-GYIMAH/vedatrace-go"
)

func main() {
	// NewDev creates a console-only logger — no API key needed.
	logger := vedatrace.NewDev("my-service")

	logger.Debug("debug message", vedatrace.LogMetadata{"detail": "verbose info"})
	logger.Info("server started", vedatrace.LogMetadata{"port": 8080})
	logger.Warn("config missing, using defaults")
	logger.Error("failed to connect", errors.New("connection refused"), vedatrace.LogMetadata{"host": "db:5432"})
}
