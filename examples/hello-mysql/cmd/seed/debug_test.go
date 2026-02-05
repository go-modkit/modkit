package main

import (
	"testing"

	modkitlogging "github.com/go-modkit/modkit/modkit/logging"
)

type debugCaptureLogger struct {
	debugMessages []string
}

func (c *debugCaptureLogger) Debug(msg string, _ ...any) {
	c.debugMessages = append(c.debugMessages, msg)
}
func (c *debugCaptureLogger) Info(string, ...any)              {}
func (c *debugCaptureLogger) Warn(string, ...any)              {}
func (c *debugCaptureLogger) Error(string, ...any)             {}
func (c *debugCaptureLogger) With(...any) modkitlogging.Logger { return c }

func TestLogSeedDebug(t *testing.T) {
	logger := &debugCaptureLogger{}
	logSeedDebug(logger, "starting seed")

	if len(logger.debugMessages) != 1 {
		t.Fatalf("expected debug log")
	}
}
