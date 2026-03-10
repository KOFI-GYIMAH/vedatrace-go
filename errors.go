package vedatrace

import (
	"fmt"
	"runtime"
)

// CaptureError builds an ErrorInfo from err without a stack trace.
func CaptureError(err error) *ErrorInfo {
	if err == nil {
		return nil
	}
	return &ErrorInfo{
		Message: err.Error(),
		Type:    fmt.Sprintf("%T", err),
	}
}

// CaptureErrorWithStack builds an ErrorInfo from err and includes the current
// goroutine stack trace.
func CaptureErrorWithStack(err error) *ErrorInfo {
	if err == nil {
		return nil
	}
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return &ErrorInfo{
		Message: err.Error(),
		Type:    fmt.Sprintf("%T", err),
		Stack:   string(buf[:n]),
	}
}
