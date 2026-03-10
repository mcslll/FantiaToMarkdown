package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/exp/slog"
)

// ANSI 颜色代码
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

// MultiHandler 将日志分发给多个 Handler
type MultiHandler struct {
	handlers []slog.Handler
}

func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			_ = h.Handle(ctx, r)
		}
	}
	return nil
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: newHandlers}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers: newHandlers}
}

// ColoredHandler 是一个支持彩色输出的自定义 slog Handler
type ColoredHandler struct {
	level    slog.Level
	attrs    []slog.Attr
	writer   io.Writer
	useColor bool
}

// NewColoredHandler 创建一个新的彩色处理器
func NewColoredHandler(level slog.Level, writer io.Writer, useColor bool) *ColoredHandler {
	return &ColoredHandler{
		level:    level,
		attrs:    make([]slog.Attr, 0),
		writer:   writer,
		useColor: useColor,
	}
}

// Enabled 检查是否应该记录给级别级别的日志
func (h *ColoredHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle 处理日志记录
func (h *ColoredHandler) Handle(_ context.Context, r slog.Record) error {
	// 获取颜色和级别字符串
	color, levelStr := h.getLevelColorAndString(r.Level)

	// 格式化时间
	timeStr := r.Time.Format("2006-01-02 15:04:05.000")

	// 获取调用者信息（文件名和行号）
	file, line := h.getCallerInfo()

	// 构建基本日志行
	var builder strings.Builder

	// 时间
	if h.useColor {
		builder.WriteString(Gray + timeStr + Reset)
	} else {
		builder.WriteString(timeStr)
	}
	builder.WriteString(" | ")

	// 级别
	if h.useColor {
		builder.WriteString(color + Bold + levelStr + Reset)
	} else {
		builder.WriteString(levelStr)
	}
	builder.WriteString(" | ")

	// 文件名和行号
	callerStr := file + ":" + fmt.Sprintf("%d", line)
	if h.useColor {
		builder.WriteString(Cyan + callerStr + Reset)
	} else {
		builder.WriteString(callerStr)
	}
	builder.WriteString(" | ")

	// 消息
	if h.useColor {
		builder.WriteString(color + r.Message + Reset)
	} else {
		builder.WriteString(r.Message)
	}

	// 处理属性
	if len(h.attrs) > 0 {
		builder.WriteString(" | ")
		for i, attr := range h.attrs {
			builder.WriteString(h.formatAttr(attr))
			if i < len(h.attrs)-1 {
				builder.WriteString(" ")
			}
		}
	}

	if r.NumAttrs() > 0 {
		builder.WriteString(" | ")
		first := true
		r.Attrs(func(attr slog.Attr) bool {
			if !first {
				builder.WriteString(" ")
			}
			builder.WriteString(h.formatAttr(attr))
			first = false
			return true
		})
	}

	if len(h.attrs) > 0 || r.NumAttrs() > 0 {
		builder.WriteString(" |")
	}

	// 输出
	fmt.Fprintln(h.writer, builder.String())
	return nil
}

// getCallerInfo 获取调用者的文件名和行号
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

// WithAttrs 返回一个带有给定属性的新处理器
func (h *ColoredHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &ColoredHandler{
		level:    h.level,
		attrs:    newAttrs,
		writer:   h.writer,
		useColor: h.useColor,
	}
}

// WithGroup 返回一个带有给定组名的新处理器
func (h *ColoredHandler) WithGroup(name string) slog.Handler {
	return h
}

// getLevelColorAndString 根据日志级别返回对应的颜色和字符串
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

// formatAttr 格式化属性
func (h *ColoredHandler) formatAttr(attr slog.Attr) string {
	if h.useColor {
		return Purple + attr.Key + Reset + "=" + Blue + fmt.Sprintf("%v", attr.Value) + Reset
	}
	return fmt.Sprintf("%s=%v", attr.Key, attr.Value)
}

// SetupLogger 创建配置好的 logger，同时支持控制台和文件输出
func SetupLogger(level slog.Level, logFilePath string) *slog.Logger {
	var handlers []slog.Handler

	// 1. 控制台彩色输出
	handlers = append(handlers, NewColoredHandler(level, os.Stdout, true))

	// 2. 文件输出 (与终端一致，但不带颜色)
	if logFilePath != "" {
		f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			handlers = append(handlers, NewColoredHandler(level, f, false))
		}
	}

	return slog.New(&MultiHandler{handlers: handlers})
}

// GetLevelColorAndString 辅助方法 (被 client 使用以打印错误)
func GetLevelColorAndString(level slog.Level) (string, string) {
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
