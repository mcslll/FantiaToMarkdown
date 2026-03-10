package logger

import (
	"context"
	"fmt"
	"golang.org/x/exp/slog"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"
	Bold   = "\033[1m"
)

type ColoredHandler struct {
	level slog.Level
	attrs []slog.Attr
}

func NewColoredHandler(level slog.Level) *ColoredHandler {
	return &ColoredHandler{
		level: level,
		attrs: make([]slog.Attr, 0),
	}
}

func (h *ColoredHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *ColoredHandler) Handle(_ context.Context, r slog.Record) error {
	color, levelStr := h.getLevelColorAndString(r.Level)
	timeStr := r.Time.Format("2006-01-02 15:04:05.000")
	file, line := h.getCallerInfo()

	var builder strings.Builder
	builder.WriteString(Gray + timeStr + Reset)
	builder.WriteString(" | ")
	builder.WriteString(color + Bold + levelStr + Reset)
	builder.WriteString(" | ")
	builder.WriteString(Cyan + file + ":" + fmt.Sprintf("%d", line) + Reset)
	builder.WriteString(" | ")
	builder.WriteString(color + r.Message + Reset)

	if r.NumAttrs() > 0 || len(h.attrs) > 0 {
		builder.WriteString(" | ")
		for _, attr := range h.attrs {
			builder.WriteString(h.formatAttr(attr))
			builder.WriteString(" ")
		}
		r.Attrs(func(attr slog.Attr) bool {
			builder.WriteString(h.formatAttr(attr))
			builder.WriteString(" ")
			return true
		})
	}

	fmt.Println(builder.String())
	return nil
}

func (h *ColoredHandler) getCallerInfo() (string, int) {
	for skip := 4; skip < 10; skip++ {
		_, file, line, ok := runtime.Caller(skip)
		if !ok {
			return "unknown", 0
		}
		if !strings.Contains(file, "slog") && !strings.Contains(file, "log") {
			return filepath.Base(file), line
		}
	}
	return "unknown", 0
}

func (h *ColoredHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &ColoredHandler{
		level: h.level,
		attrs: newAttrs,
	}
}

func (h *ColoredHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *ColoredHandler) getLevelColorAndString(level slog.Level) (string, string) {
	switch level {
	case slog.LevelDebug:
		return Cyan, "DEBUG"
	case slog.LevelInfo:
		return Green, "INFO "
	case slog.LevelWarn:
		return Yellow, "WARN "
	case slog.LevelError:
		return Red, "ERROR"
	default:
		return White, "TRACE"
	}
}

func (h *ColoredHandler) formatAttr(attr slog.Attr) string {
	return Purple + attr.Key + Reset + "=" + Blue + fmt.Sprintf("%v", attr.Value) + Reset
}

func SetupLogger(level slog.Level) *slog.Logger {
	handler := NewColoredHandler(level)
	return slog.New(handler)
}
