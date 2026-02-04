package logging

import "log/slog"

type nopLogger struct{}

func Nop() Logger {
	return nopLogger{}
}

func (nopLogger) Debug(string, ...slog.Attr) {}
func (nopLogger) Info(string, ...slog.Attr)  {}
func (nopLogger) Error(string, ...slog.Attr) {}
func (nopLogger) With(...slog.Attr) Logger   { return nopLogger{} }
