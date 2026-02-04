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
func (h *captureHandler) WithGroup(_ string) slog.Handler     { return h }

func TestSlogAdapter_EmitsRecords(t *testing.T) {
	ch := &captureHandler{}
	base := slog.New(ch)
	logger := NewSlog(base)

	logger.Info("hello", slog.String("k", "v"))

	if len(ch.records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(ch.records))
	}
	if ch.records[0].Message != "hello" {
		t.Fatalf("unexpected message: %s", ch.records[0].Message)
	}
}
