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
func (c *captureLogger) With(...slog.Attr) modkitlogging.Logger { return c }

func TestLogStartup_EmitsMessage(t *testing.T) {
	logger := &captureLogger{}
	logStartup(logger, ":8080")

	if len(logger.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(logger.messages))
	}
	if logger.messages[0] != "server starting" {
		t.Fatalf("unexpected message: %s", logger.messages[0])
	}

	var addr string
	for _, attr := range logger.attrs {
		if attr.Key == "addr" {
			addr = attr.Value.String()
			break
		}
	}
	if addr != ":8080" {
		t.Fatalf("expected addr :8080, got %q", addr)
	}
}
