package vedatrace

import "time"

const (
	defaultEndpoint      = "https://ingest.vedatrace.dev/v1/logs"
	defaultBatchSize     = 100
	defaultFlushInterval = 5 * time.Second
	defaultMaxRetries    = 3
	defaultRetryDelay    = 1 * time.Second
)

// Config holds all configuration for a VedaTrace logger instance.
type Config struct {
	// APIKey is required for HTTP transport. Obtain from the VedaTrace dashboard.
	APIKey string

	// Service is the name of the application or service emitting logs.
	// It is attached to every log entry.
	Service string

	// Endpoint is the VedaTrace ingest URL.
	// Defaults to https://ingest.vedatrace.dev/v1/logs
	Endpoint string

	// BatchSize is the maximum number of log entries collected before an
	// automatic flush is triggered. Defaults to 100.
	BatchSize int

	// FlushInterval is how often the batcher flushes even when BatchSize has
	// not been reached. Defaults to 5 seconds.
	FlushInterval time.Duration

	// MaxRetries is the number of additional attempts made after a failed HTTP
	// send. Defaults to 3.
	MaxRetries int

	// RetryDelay is the base delay between retry attempts. Each retry is
	// multiplied by the attempt number (linear back-off). Defaults to 1 second.
	RetryDelay time.Duration

	// RedactFields is a list of dot-notation field paths within LogMetadata
	// whose values will be replaced with "[REDACTED]" before transmission.
	// Example: []string{"password", "card.cvv"}
	RedactFields []string

	// OnSuccess is called after each batch is successfully delivered.
	// It receives a copy of the batch. The callback must be non-blocking.
	OnSuccess func(batch []LogEntry)

	// OnError is called when a batch fails all retry attempts.
	// The callback must be non-blocking.
	OnError func(err error, batch []LogEntry)

	// DisableHTTP forces console-only output regardless of whether an APIKey
	// is set. Useful in development and testing.
	DisableHTTP bool
}

// withDefaults returns a new Config where every zero-valued field has been
// filled with the package default.
func (c Config) withDefaults() Config {
	if c.Endpoint == "" {
		c.Endpoint = defaultEndpoint
	}
	if c.BatchSize <= 0 {
		c.BatchSize = defaultBatchSize
	}
	if c.FlushInterval <= 0 {
		c.FlushInterval = defaultFlushInterval
	}
	if c.MaxRetries <= 0 {
		c.MaxRetries = defaultMaxRetries
	}
	if c.RetryDelay <= 0 {
		c.RetryDelay = defaultRetryDelay
	}
	return c
}
