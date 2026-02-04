package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestPrettyHandler_AlignsLevelColumn(t *testing.T) {
	clearEnv()
	t.Setenv("LOG_FORMAT", "text")
	t.Setenv("LOG_STYLE", "pretty")
	t.Setenv("LOG_COLOR", "off")
	t.Setenv("LOG_LEVEL", "debug")

	var buf bytes.Buffer
	logger := newLogger(&buf).With(slog.String("scope", "test"))
	logger.Debug("debug line")
	logger.Info("info line")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected two lines, got %d", len(lines))
	}
	idxDebug := strings.Index(lines[0], "[")
	idxInfo := strings.Index(lines[1], "[")
	if idxDebug == -1 || idxInfo == -1 {
		t.Fatalf("expected scope brackets, got: %q / %q", lines[0], lines[1])
	}
	if idxDebug != idxInfo {
		t.Fatalf("expected aligned scope column, got %d and %d", idxDebug, idxInfo)
	}
}
