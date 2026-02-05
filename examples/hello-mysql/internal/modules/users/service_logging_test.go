package users

import (
	"context"
	"log/slog"
	"testing"

	modkitlogging "github.com/go-modkit/modkit/modkit/logging"
)

type captureLogger struct {
	debugMessages []string
	attrs         []slog.Attr
}

func (c *captureLogger) Debug(msg string, args ...any) {
	c.debugMessages = append(c.debugMessages, msg)
	c.attrs = append(c.attrs, attrsFromArgs(args)...)
}
func (c *captureLogger) Info(string, ...any)  {}
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
