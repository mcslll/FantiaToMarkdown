package logger

import (
	"testing"

	"golang.org/x/exp/slog"
)

func TestSetupLogger(t *testing.T) {
	// SetupLogger 返回的是 *slog.Logger
	l := SetupLogger(slog.LevelDebug)
	if l == nil {
		t.Fatal("SetupLogger returned nil logger")
	}

	l.Info("Test logging")
}
