package main

import (
	"log/slog"
	"testing"

	modkitlogging "github.com/aryeko/modkit/modkit/logging"
)

type captureLogger struct {
	messages []string
	attrs    []slog.Attr
}

func (c *captureLogger) Debug(string, ...slog.Attr) {}
func (c *captureLogger) Info(msg string, attrs ...slog.Attr) {
	c.messages = append(c.messages, msg)
	c.attrs = append(c.attrs, attrs...)
}
func (c *captureLogger) Error(string, ...slog.Attr) {}
func (c *captureLogger) With(attrs ...slog.Attr) modkitlogging.Logger {
	c.attrs = append(c.attrs, attrs...)
	return c
}

func TestLogSeedComplete(t *testing.T) {
	logger := &captureLogger{}
	logSeedComplete(logger)

	if len(logger.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(logger.messages))
	}
	if logger.messages[0] != "seed complete" {
		t.Fatalf("unexpected message: %s", logger.messages[0])
	}
	var scope string
	for _, attr := range logger.attrs {
		if attr.Key == "scope" {
			scope = attr.Value.String()
			break
		}
	}
	if scope != "seed" {
		t.Fatalf("expected scope seed, got %q", scope)
	}
}
