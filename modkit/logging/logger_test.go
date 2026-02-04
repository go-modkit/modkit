package logging

import (
	"context"
	"log/slog"
	"testing"
)

type captureHandler struct {
	records []slog.Record
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.records = append(h.records, r)
	return nil
}
func (h *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(_ string) slog.Handler      { return h }

func TestSlogAdapter_EmitsRecords(t *testing.T) {
	ch := &captureHandler{}
	base := slog.New(ch)
	logger := NewSlogLogger(base)

	logger.Info("hello", "k", "v")

	if len(ch.records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(ch.records))
	}
	if ch.records[0].Message != "hello" {
		t.Fatalf("unexpected message: %s", ch.records[0].Message)
	}

	var gotValue slog.Value
	ch.records[0].Attrs(func(attr slog.Attr) bool {
		if attr.Key == "k" {
			gotValue = attr.Value
			return false
		}
		return true
	})
	if gotValue.Kind() == slog.KindAny && gotValue.Any() == nil {
		t.Fatalf("expected attr k to be set")
	}
	if gotValue.String() != "v" {
		t.Fatalf("unexpected attr value: %s", gotValue.String())
	}
}

func TestSlogAdapter_WarnLevel(t *testing.T) {
	ch := &captureHandler{}
	base := slog.New(ch)
	logger := NewSlogLogger(base)

	logger.Warn("heads up")

	if len(ch.records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(ch.records))
	}
	if ch.records[0].Level != slog.LevelWarn {
		t.Fatalf("unexpected level: %s", ch.records[0].Level)
	}
}

func TestNopLogger(t *testing.T) {
	logger := NewNopLogger()

	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")
	logger.With("key", "value").Info("with")
}
