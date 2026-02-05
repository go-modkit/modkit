package config

import "testing"

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("MYSQL_DSN", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("JWT_ISSUER", "")
	t.Setenv("JWT_TTL", "")
	t.Setenv("AUTH_USERNAME", "")
	t.Setenv("AUTH_PASSWORD", "")

	cfg := Load()

	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	if cfg.JWTSecret != "dev-secret-change-me" {
		t.Fatalf("JWTSecret = %q", cfg.JWTSecret)
	}
}

func TestEnvOrDefault_TrimsSpace(t *testing.T) {
	t.Setenv("JWT_ISSUER", "   ")
	if got := envOrDefault("JWT_ISSUER", "hello-mysql"); got != "hello-mysql" {
		t.Fatalf("envOrDefault = %q", got)
	}
}
