package vedatrace

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// ANSI colour codes used for console output.
const (
	colReset   = "\033[0m"
	colCyan    = "\033[36m"
	colGreen   = "\033[32m"
	colYellow  = "\033[33m"
	colRed     = "\033[31m"
	colMagenta = "\033[35m"
)

// consoleTransport writes log entries to stderr as human-readable lines.
// It is used when DisableHTTP is true or no APIKey is provided.
type consoleTransport struct{}

func (c *consoleTransport) Send(_ context.Context, entries []LogEntry) error {
	for _, e := range entries {
		fmt.Fprintf(os.Stderr, "%s%s%s %s [%s] %s",
			levelColor(e.Level),
			strings.ToUpper(e.Level.String()),
			colReset,
			e.Timestamp.Format(time.RFC3339),
			e.Service,
			e.Message,
		)
		if e.Error != nil {
			fmt.Fprintf(os.Stderr, " error=%q type=%s", e.Error.Message, e.Error.Type)
		}
		if len(e.Metadata) > 0 {
			b, _ := json.Marshal(e.Metadata)
			fmt.Fprintf(os.Stderr, " %s", b)
		}
		fmt.Fprintln(os.Stderr)
	}
	return nil
}

func levelColor(l Level) string {
	switch l {
	case LevelDebug:
		return colCyan
	case LevelInfo:
		return colGreen
	case LevelWarn:
		return colYellow
	case LevelError:
		return colRed
	case LevelFatal:
		return colMagenta
	default:
		return ""
	}
}
