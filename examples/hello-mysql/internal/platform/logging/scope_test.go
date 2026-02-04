package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestPrettyHandler_UsesScopeAttr(t *testing.T) {
	clearEnv()
	t.Setenv("LOG_FORMAT", "text")
	t.Setenv("LOG_STYLE", "pretty")
	t.Setenv("LOG_COLOR", "off")

	var buf bytes.Buffer
	logger := newLogger(&buf)
	logger.Info("hello", slog.String("scope", "api"))

	out := buf.String()
	if !strings.Contains(out, "[api]") {
		t.Fatalf("expected scope in output, got %s", out)
	}
	if strings.Contains(out, "scope=api") {
		t.Fatalf("expected scope removed from meta, got %s", out)
	}
}
