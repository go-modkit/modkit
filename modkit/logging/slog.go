package logging

import "log/slog"

type slogAdapter struct {
	logger *slog.Logger
}

func NewSlogLogger(logger *slog.Logger) Logger {
	if logger == nil {
		return NewNopLogger()
	}
	return slogAdapter{logger: logger}
}

func (s slogAdapter) Debug(msg string, args ...any) {
	s.logger.Debug(msg, args...)
}

func (s slogAdapter) Info(msg string, args ...any) {
	s.logger.Info(msg, args...)
}

func (s slogAdapter) Warn(msg string, args ...any) {
	s.logger.Warn(msg, args...)
}

func (s slogAdapter) Error(msg string, args ...any) {
	s.logger.Error(msg, args...)
}

func (s slogAdapter) With(args ...any) Logger {
	return slogAdapter{logger: s.logger.With(args...)}
}
