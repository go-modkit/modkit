package main

import (
	"log/slog"
	"testing"

	modkitlogging "github.com/aryeko/modkit/modkit/logging"
)

type debugCaptureLogger struct {
	debugMessages []string
}

func (c *debugCaptureLogger) Debug(msg string, _ ...slog.Attr) {
	c.debugMessages = append(c.debugMessages, msg)
}
func (c *debugCaptureLogger) Info(string, ...slog.Attr)              {}
func (c *debugCaptureLogger) Error(string, ...slog.Attr)             {}
func (c *debugCaptureLogger) With(...slog.Attr) modkitlogging.Logger { return c }

func TestLogSeedDebug(t *testing.T) {
	logger := &debugCaptureLogger{}
	logSeedDebug(logger, "starting seed")

	if len(logger.debugMessages) != 1 {
		t.Fatalf("expected debug log")
	}
}
