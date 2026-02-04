package logging

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestParseConfigDefaults(t *testing.T) {
	clearEnv()
	cfg := parseConfig()
	if cfg.format != "text" {
		t.Fatalf("expected format text, got %s", cfg.format)
	}
	if cfg.level != slog.LevelInfo {
		t.Fatalf("expected level info, got %v", cfg.level)
	}
	if cfg.color != "auto" {
		t.Fatalf("expected color auto, got %s", cfg.color)
	}
	if cfg.timeFormat != "local" {
		t.Fatalf("expected time local, got %s", cfg.timeFormat)
	}
	if cfg.style != "pretty" {
		t.Fatalf("expected style pretty, got %s", cfg.style)
	}
}

func TestParseConfigOverrides(t *testing.T) {
	clearEnv()
	t.Setenv("LOG_FORMAT", "json")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_COLOR", "on")
	t.Setenv("LOG_TIME", "utc")
	t.Setenv("LOG_STYLE", "plain")

	cfg := parseConfig()
	if cfg.format != "json" {
		t.Fatalf("expected format json, got %s", cfg.format)
	}
	if cfg.level != slog.LevelDebug {
		t.Fatalf("expected level debug, got %v", cfg.level)
	}
	if cfg.color != "on" {
		t.Fatalf("expected color on, got %s", cfg.color)
	}
	if cfg.timeFormat != "utc" {
		t.Fatalf("expected time utc, got %s", cfg.timeFormat)
	}
	if cfg.style != "plain" {
		t.Fatalf("expected style plain, got %s", cfg.style)
	}
}

func TestBuildHandler_TextColorOn(t *testing.T) {
	clearEnv()
	t.Setenv("LOG_FORMAT", "text")
	t.Setenv("LOG_COLOR", "on")

	var buf bytes.Buffer
	logger := newLogger(&buf)
	logger.Info("hello")

	out := buf.String()
	if !strings.Contains(out, "hello") {
		t.Fatalf("expected output to contain message, got %s", out)
	}
	if !strings.Contains(out, "\x1b[") && !strings.Contains(out, "\\x1b[") {
		t.Fatalf("expected color codes, got %s", out)
	}
}

func TestBuildHandler_TextPrettyFormat(t *testing.T) {
	clearEnv()
	t.Setenv("LOG_FORMAT", "text")
	t.Setenv("LOG_STYLE", "pretty")
	t.Setenv("LOG_COLOR", "off")

	var buf bytes.Buffer
	logger := newLogger(&buf)
	logger.Info("server starting", slog.String("addr", ":8080"))

	out := buf.String()
	if !strings.Contains(out, "INFO") {
		t.Fatalf("expected level INFO, got %s", out)
	}
	if !strings.Contains(out, "addr=:8080") {
		t.Fatalf("expected addr field, got %s", out)
	}
	if !strings.Contains(out, "  ") {
		t.Fatalf("expected spacing between columns, got %s", out)
	}
}

func TestBuildHandler_TextMultilineFormat(t *testing.T) {
	clearEnv()
	t.Setenv("LOG_FORMAT", "text")
	t.Setenv("LOG_STYLE", "multiline")
	t.Setenv("LOG_COLOR", "off")

	var buf bytes.Buffer
	logger := newLogger(&buf)
	logger.Info("server starting", slog.String("addr", ":8080"))

	out := buf.String()
	if !strings.Contains(out, "\n  ") {
		t.Fatalf("expected multiline output, got %s", out)
	}
	if !strings.Contains(out, "addr=:8080") {
		t.Fatalf("expected addr field, got %s", out)
	}
}

func TestBuildHandler_JSON(t *testing.T) {
	clearEnv()
	t.Setenv("LOG_FORMAT", "json")

	var buf bytes.Buffer
	logger := newLogger(&buf)
	logger.Info("hello")

	out := buf.String()
	if !strings.HasPrefix(strings.TrimSpace(out), "{") {
		t.Fatalf("expected json output, got %s", out)
	}
	if !strings.Contains(out, "\"msg\":\"hello\"") {
		t.Fatalf("expected msg field, got %s", out)
	}
}

func clearEnv() {
	_ = os.Unsetenv("LOG_FORMAT")
	_ = os.Unsetenv("LOG_LEVEL")
	_ = os.Unsetenv("LOG_COLOR")
	_ = os.Unsetenv("LOG_TIME")
	_ = os.Unsetenv("LOG_STYLE")
}
