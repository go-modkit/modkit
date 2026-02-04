package logging

import "log/slog"

type slogAdapter struct {
	logger *slog.Logger
}

func NewSlog(logger *slog.Logger) Logger {
	if logger == nil {
		return Nop()
	}
	return slogAdapter{logger: logger}
}

func (s slogAdapter) Debug(msg string, attrs ...slog.Attr) {
	s.logger.Debug(msg, attrsToAny(attrs)...)
}

func (s slogAdapter) Info(msg string, attrs ...slog.Attr) {
	s.logger.Info(msg, attrsToAny(attrs)...)
}

func (s slogAdapter) Error(msg string, attrs ...slog.Attr) {
	s.logger.Error(msg, attrsToAny(attrs)...)
}

func (s slogAdapter) With(attrs ...slog.Attr) Logger {
	return slogAdapter{logger: s.logger.With(attrsToAny(attrs)...)}
}

func attrsToAny(attrs []slog.Attr) []any {
	if len(attrs) == 0 {
		return nil
	}
	converted := make([]any, len(attrs))
	for i, attr := range attrs {
		converted[i] = attr
	}
	return converted
}
