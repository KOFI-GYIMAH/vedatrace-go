package vedatrace

import (
	"errors"
	"strings"
	"testing"
)

func TestCaptureError_nil(t *testing.T) {
	if got := CaptureError(nil); got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestCaptureError(t *testing.T) {
	err := errors.New("boom")
	info := CaptureError(err)
	if info == nil {
		t.Fatal("expected non-nil ErrorInfo")
	}
	if info.Message != "boom" {
		t.Errorf("message: got %q, want %q", info.Message, "boom")
	}
	if info.Type == "" {
		t.Error("expected non-empty Type")
	}
	if info.Stack != "" {
		t.Error("CaptureError should not include a stack trace")
	}
}

func TestCaptureErrorWithStack(t *testing.T) {
	err := errors.New("with stack")
	info := CaptureErrorWithStack(err)
	if info == nil {
		t.Fatal("expected non-nil ErrorInfo")
	}
	if !strings.Contains(info.Stack, "goroutine") {
		t.Errorf("stack trace looks wrong: %q", info.Stack)
	}
}
