package main

import (
	"log/slog"
	"testing"

	modkitlogging "github.com/go-modkit/modkit/modkit/logging"
)

type captureLogger struct {
	messages []string
	attrs    []slog.Attr
}

func (c *captureLogger) Debug(string, ...any) {}
func (c *captureLogger) Info(msg string, args ...any) {
	c.messages = append(c.messages, msg)
	c.attrs = append(c.attrs, attrsFromArgs(args)...)
}
func (c *captureLogger) Warn(string, ...any)  {}
func (c *captureLogger) Error(string, ...any) {}
func (c *captureLogger) With(args ...any) modkitlogging.Logger {
	c.attrs = append(c.attrs, attrsFromArgs(args)...)
	return c
}

func attrsFromArgs(args []any) []slog.Attr {
	if len(args) == 0 {
		return nil
	}
	attrs := make([]slog.Attr, 0, len(args))
	for _, arg := range args {
		attr, ok := arg.(slog.Attr)
		if ok {
			attrs = append(attrs, attr)
		}
	}
	return attrs
}

func TestLogMigrateComplete(t *testing.T) {
	logger := &captureLogger{}
	logMigrateComplete(logger)

	if len(logger.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(logger.messages))
	}
	if logger.messages[0] != "migrations complete" {
		t.Fatalf("unexpected message: %s", logger.messages[0])
	}
	var scope string
	for _, attr := range logger.attrs {
		if attr.Key == "scope" {
			scope = attr.Value.String()
			break
		}
	}
	if scope != "migrate" {
		t.Fatalf("expected scope migrate, got %q", scope)
	}
}
