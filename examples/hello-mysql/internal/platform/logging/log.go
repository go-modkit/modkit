package logging

import (
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
			if attr.Key == slog.LevelKey && cfg.format == "text" && colorEnabled(cfg.color, w) {
				level := attr.Value.Any().(slog.Level)
				attr.Value = slog.StringValue(colorizeLevel(level))
			}
			return attr
		},
	}

	if cfg.format == "json" {
		return slog.NewJSONHandler(w, opts)
	}
	return slog.NewTextHandler(w, opts)
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
