package logger

import (
	"strings"
	"testing"

	"golang.org/x/exp/slog"
)

func TestSetupLogger(t *testing.T) {
	// SetupLogger 返回的是 *slog.Logger
	l := SetupLogger(slog.LevelDebug, "")
	if l == nil {
		t.Fatal("SetupLogger returned nil logger")
	}

	l.Info("Test logging")
}

func TestColoredHandlerAttributes(t *testing.T) {
	var buf strings.Builder
	h := NewColoredHandler(slog.LevelInfo, &buf, false) // No color for easier testing
	l := slog.New(h)

	// Case 1: Handler-level only (e.g. progress)
	l1 := l.With("post", "1/10")
	l1.Info("Task started")
	if !strings.Contains(buf.String(), "| Task started | post=1/10 |") {
		t.Errorf("Format error (handler only): %s", buf.String())
	}
	buf.Reset()

	// Case 2: Record-level only
	l.Info("Download error", "url", "http://err.com")
	if !strings.Contains(buf.String(), "| Download error | url=http://err.com |") {
		t.Errorf("Format error (record only): %s", buf.String())
	}
	buf.Reset()

	// Case 3: Both levels (Our main use case)
	l3 := l.With("post", "5/20")
	l3.Info("Saving file", "id", "999", "status", "ok")
	// Expected order: Message | Handler Attrs | Record Attrs |
	if !strings.Contains(buf.String(), "| Saving file | post=5/20 | id=999 status=ok |") {
		t.Errorf("Format error (combined): %s", buf.String())
	}
}
