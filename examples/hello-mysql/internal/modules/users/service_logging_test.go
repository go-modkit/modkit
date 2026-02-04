package users

import (
	"context"
	"log/slog"
	"testing"

	modkitlogging "github.com/aryeko/modkit/modkit/logging"
)

type captureLogger struct {
	debugMessages []string
	attrs         []slog.Attr
}

func (c *captureLogger) Debug(msg string, attrs ...slog.Attr) {
	c.debugMessages = append(c.debugMessages, msg)
	c.attrs = append(c.attrs, attrs...)
}
func (c *captureLogger) Info(string, ...slog.Attr)  {}
func (c *captureLogger) Error(string, ...slog.Attr) {}
func (c *captureLogger) With(attrs ...slog.Attr) modkitlogging.Logger {
	c.attrs = append(c.attrs, attrs...)
	return c
}

func TestService_EmitsDebugLogs(t *testing.T) {
	repo := &stubRepo{}
	logger := &captureLogger{}
	svc := NewService(repo, logger)

	_, _ = svc.GetUser(context.Background(), 1)

	if len(logger.debugMessages) == 0 {
		t.Fatalf("expected debug logs")
	}
	var scope string
	for _, attr := range logger.attrs {
		if attr.Key == "scope" {
			scope = attr.Value.String()
			break
		}
	}
	if scope != "users" {
		t.Fatalf("expected scope users, got %q", scope)
	}
}
