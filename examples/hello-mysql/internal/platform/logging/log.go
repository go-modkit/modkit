package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	modkitlogging "github.com/aryeko/modkit/modkit/logging"
	"golang.org/x/term"
)

type config struct {
	format     string
	level      slog.Level
	color      string
	timeFormat string
	style      string
}

func New() modkitlogging.Logger {
	logger := newLogger(os.Stdout)
	return modkitlogging.NewSlog(logger)
}

func newLogger(w io.Writer) *slog.Logger {
	cfg := parseConfig()
	handler := buildHandler(cfg, w)
	return slog.New(handler)
}

func parseConfig() config {
	cfg := config{
		format:     envDefault("LOG_FORMAT", "text"),
		level:      slog.LevelInfo,
		color:      envDefault("LOG_COLOR", "auto"),
		timeFormat: envDefault("LOG_TIME", "local"),
		style:      envDefault("LOG_STYLE", "pretty"),
	}

	levelValue := strings.ToLower(envDefault("LOG_LEVEL", "info"))
	switch levelValue {
	case "debug":
		cfg.level = slog.LevelDebug
	case "info":
		cfg.level = slog.LevelInfo
	case "warn", "warning":
		cfg.level = slog.LevelWarn
	case "error":
		cfg.level = slog.LevelError
	}

	if cfg.format != "json" && cfg.format != "text" {
		cfg.format = "text"
	}
	if cfg.color != "on" && cfg.color != "off" && cfg.color != "auto" {
		cfg.color = "auto"
	}
	if cfg.timeFormat != "local" && cfg.timeFormat != "utc" && cfg.timeFormat != "none" {
		cfg.timeFormat = "local"
	}
	if cfg.style != "pretty" && cfg.style != "plain" && cfg.style != "multiline" {
		cfg.style = "pretty"
	}

	return cfg
}

func buildHandler(cfg config, w io.Writer) slog.Handler {
	opts := &slog.HandlerOptions{
		Level: cfg.level,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey {
				if cfg.timeFormat == "none" {
					return slog.Attr{}
				}
				if cfg.timeFormat == "utc" {
					if t, ok := attr.Value.Any().(time.Time); ok {
						attr.Value = slog.TimeValue(t.UTC())
					}
				}
			}
			if attr.Key == slog.LevelKey && cfg.format == "text" && cfg.style == "plain" && colorEnabled(cfg.color, w) {
				level := attr.Value.Any().(slog.Level)
				attr.Value = slog.StringValue(colorizeLevel(level))
			}
			return attr
		},
	}

	if cfg.format == "json" {
		return slog.NewJSONHandler(w, opts)
	}
	if cfg.style == "plain" {
		return slog.NewTextHandler(w, opts)
	}
	return newPrettyHandler(w, cfg)
}

func envDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return strings.ToLower(value)
}

func colorEnabled(mode string, w io.Writer) bool {
	switch mode {
	case "on":
		return true
	case "off":
		return false
	default:
		_, ok := w.(*os.File)
		return ok && isTerminal(w.(*os.File))
	}
}

func isTerminal(file *os.File) bool {
	if file == nil {
		return false
	}
	return term.IsTerminal(int(file.Fd()))
}

func colorizeLevel(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return "\x1b[34mDEBUG\x1b[0m"
	case slog.LevelInfo:
		return "\x1b[32mINFO\x1b[0m"
	case slog.LevelWarn:
		return "\x1b[33mWARN\x1b[0m"
	case slog.LevelError:
		return "\x1b[31mERROR\x1b[0m"
	default:
		return level.String()
	}
}

func padLevel(level string, width int) string {
	if len(level) >= width {
		return level
	}
	return level + strings.Repeat(" ", width-len(level))
}

type prettyHandler struct {
	w      io.Writer
	cfg    config
	attrs  []slog.Attr
	groups []string
	mutex  chan struct{}
}

func newPrettyHandler(w io.Writer, cfg config) slog.Handler {
	return &prettyHandler{
		w:     w,
		cfg:   cfg,
		mutex: make(chan struct{}, 1),
	}
}

func (h *prettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.cfg.level
}

func (h *prettyHandler) Handle(_ context.Context, record slog.Record) error {
	h.lock()
	defer h.unlock()

	builder := strings.Builder{}
	timeStr := h.formatTime(record.Time)
	levelStr := h.formatLevel(record.Level)
	scope := h.extractScope(record)

	if timeStr != "" {
		builder.WriteString(timeStr)
		builder.WriteString(" ")
	}
	if levelStr != "" {
		builder.WriteString(levelStr)
		builder.WriteString("  ")
	}
	if scope != "" {
		builder.WriteString("[")
		builder.WriteString(scope)
		builder.WriteString("] ")
	}
	builder.WriteString(record.Message)

	fields := h.collectAttrs(record)
	if len(fields) > 0 {
		switch h.cfg.style {
		case "multiline":
			builder.WriteString("\n  ")
			builder.WriteString(strings.Join(fields, " "))
		default:
			builder.WriteString("  ")
			builder.WriteString(strings.Join(fields, " "))
		}
	}
	builder.WriteString("\n")

	_, err := io.WriteString(h.w, builder.String())
	return err
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := h.clone()
	clone.attrs = append(clone.attrs, attrs...)
	return clone
}

func (h *prettyHandler) WithGroup(name string) slog.Handler {
	clone := h.clone()
	clone.groups = append(clone.groups, name)
	return clone
}

func (h *prettyHandler) clone() *prettyHandler {
	return &prettyHandler{
		w:      h.w,
		cfg:    h.cfg,
		attrs:  append([]slog.Attr{}, h.attrs...),
		groups: append([]string{}, h.groups...),
		mutex:  h.mutex,
	}
}

func (h *prettyHandler) collectAttrs(record slog.Record) []string {
	fields := make([]string, 0, record.NumAttrs()+len(h.attrs))
	for _, attr := range h.attrs {
		if attr.Key == "scope" {
			continue
		}
		fields = append(fields, h.formatAttr(attr))
	}
	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key == "scope" {
			return true
		}
		fields = append(fields, h.formatAttr(attr))
		return true
	})
	return fields
}

func (h *prettyHandler) formatAttr(attr slog.Attr) string {
	if len(h.groups) > 0 {
		attr.Key = strings.Join(h.groups, ".") + "." + attr.Key
	}
	return attr.Key + "=" + attr.Value.String()
}

func (h *prettyHandler) formatTime(t time.Time) string {
	if h.cfg.timeFormat == "none" || t.IsZero() {
		return ""
	}
	if h.cfg.timeFormat == "utc" {
		t = t.UTC()
	}
	return t.Format("15:04:05")
}

func (h *prettyHandler) formatLevel(level slog.Level) string {
	raw := strings.ToUpper(level.String())
	padded := padLevel(raw, 5)
	if h.cfg.format == "text" && colorEnabled(h.cfg.color, h.w) {
		colored := colorizeLevel(level)
		if len(raw) < 5 {
			return colored + strings.Repeat(" ", 5-len(raw))
		}
		return colored
	}
	return padded
}

func (h *prettyHandler) extractScope(record slog.Record) string {
	for _, attr := range h.attrs {
		if attr.Key == "scope" {
			return attr.Value.String()
		}
	}
	var scope string
	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key == "scope" {
			scope = attr.Value.String()
			return false
		}
		return true
	})
	return scope
}

func (h *prettyHandler) lock() {
	h.mutex <- struct{}{}
}

func (h *prettyHandler) unlock() {
	<-h.mutex
}
