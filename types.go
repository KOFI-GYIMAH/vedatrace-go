package vedatrace

import "time"

// Level represents the severity of a log entry.
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
	LevelFatal Level = "fatal"
)

// validLevels is used for input validation.
var validLevels = map[Level]struct{}{
	LevelDebug: {},
	LevelInfo:  {},
	LevelWarn:  {},
	LevelError: {},
	LevelFatal: {},
}

// IsValid reports whether l is a recognize log level.
func (l Level) IsValid() bool {
	_, ok := validLevels[l]
	return ok
}

// String implements fmt.Stringer.
func (l Level) String() string { return string(l) }

// ErrorInfo holds structured information about a captured error.
type ErrorInfo struct {
	Message string `json:"message"`
	// Type is the Go type name of the error, e.g. "*url.Error".
	Type string `json:"type,omitempty"`
	// Stack is an optional stack trace string.
	Stack string `json:"stack,omitempty"`
}

// LogMetadata is a free-form map of contextual key/value pairs attached to a log entry.
type LogMetadata map[string]any

// LogEntry is a single structured log record.
// This is the schema sent to the VedaTrace ingest API.
type LogEntry struct {
	Level     Level       `json:"level"`
	Message   string      `json:"message"`
	Service   string      `json:"service"`
	Timestamp time.Time   `json:"timestamp"`
	Metadata  LogMetadata `json:"metadata,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
}

// IngestPayload is the envelope sent in each HTTP POST to the ingest endpoint.
type IngestPayload struct {
	Logs []LogEntry `json:"logs"`
}
