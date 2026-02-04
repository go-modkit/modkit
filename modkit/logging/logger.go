package logging

import "log/slog"

type Logger interface {
	Debug(msg string, attrs ...slog.Attr)
	Info(msg string, attrs ...slog.Attr)
	Error(msg string, attrs ...slog.Attr)
	With(attrs ...slog.Attr) Logger
}
