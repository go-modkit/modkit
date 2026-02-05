package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-modkit/modkit/examples/hello-mysql/internal/lifecycle"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/platform/config"
)

type stubServer struct {
	shutdownCalled bool
	shutdownErr    error
	shutdownCh     chan struct{}
}

func (s *stubServer) ListenAndServe() error {
	return nil
}

func (s *stubServer) Shutdown(ctx context.Context) error {
	s.shutdownCalled = true
	if s.shutdownCh != nil {
		close(s.shutdownCh)
	}
	return s.shutdownErr
}

func TestRunServer_ShutdownPath(t *testing.T) {
	server := &stubServer{}
	sigCh := make(chan os.Signal, 1)
	errCh := make(chan error, 1)
	cleanupCalled := false
	hooks := []lifecycle.CleanupHook{
		func(ctx context.Context) error {
			cleanupCalled = true
			return nil
		},
	}

	server.shutdownCh = make(chan struct{})
	go func() {
		<-server.shutdownCh
		errCh <- http.ErrServerClosed
	}()
	sigCh <- os.Interrupt

	err := runServer(50*time.Millisecond, server, sigCh, errCh, hooks)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !server.shutdownCalled {
		t.Fatal("expected shutdown to be called")
	}
	if !cleanupCalled {
		t.Fatal("expected cleanup to be called")
	}
}

func TestRunServer_ReturnsListenError(t *testing.T) {
	server := &stubServer{}
	sigCh := make(chan os.Signal, 1)
	errCh := make(chan error, 1)
	errCh <- errors.New("listen failed")

	err := runServer(50*time.Millisecond, server, sigCh, errCh, nil)
	if err == nil || err.Error() != "listen failed" {
		t.Fatalf("expected listen error, got %v", err)
	}
	if server.shutdownCalled {
		t.Fatal("shutdown should not be called")
	}
}

func TestRunServer_ShutdownReturnsError(t *testing.T) {
	server := &stubServer{shutdownErr: errors.New("shutdown failed"), shutdownCh: make(chan struct{})}
	sigCh := make(chan os.Signal, 1)
	errCh := make(chan error, 1)
	go func() {
		<-server.shutdownCh
		errCh <- http.ErrServerClosed
	}()
	sigCh <- os.Interrupt

	err := runServer(50*time.Millisecond, server, sigCh, errCh, nil)
	if err == nil || err.Error() != "shutdown failed" {
		t.Fatalf("expected shutdown error, got %v", err)
	}
}

func TestParseJWTTTL_InvalidFallsBack(t *testing.T) {
	got := parseJWTTTL("nope")
	if got != time.Hour {
		t.Fatalf("expected 1h fallback, got %v", got)
	}
}

func TestParseJWTTTL_NonPositiveFallsBack(t *testing.T) {
	got := parseJWTTTL("0s")
	if got != time.Hour {
		t.Fatalf("expected 1h fallback, got %v", got)
	}
}

func TestParseJWTTTL_Valid(t *testing.T) {
	got := parseJWTTTL("2h")
	if got != 2*time.Hour {
		t.Fatalf("expected 2h, got %v", got)
	}
}

func TestBuildAuthConfig_MapsFields(t *testing.T) {
	cfg := config.Config{JWTSecret: "s", JWTIssuer: "i", AuthUsername: "u", AuthPassword: "p"}
	got := buildAuthConfig(cfg, 5*time.Minute)
	if got.Secret != "s" || got.Issuer != "i" || got.Username != "u" || got.Password != "p" || got.TTL != 5*time.Minute {
		t.Fatalf("unexpected auth config: %+v", got)
	}
}

func TestBuildAppOptions_MapsFields(t *testing.T) {
	cfg := config.Config{HTTPAddr: ":1234", MySQLDSN: "dsn", JWTSecret: "s", JWTIssuer: "i", AuthUsername: "u", AuthPassword: "p"}
	got := buildAppOptions(cfg, 10*time.Minute)
	if got.HTTPAddr != ":1234" || got.MySQLDSN != "dsn" {
		t.Fatalf("unexpected options: %+v", got)
	}
	if got.Auth.Secret != "s" || got.Auth.Issuer != "i" || got.Auth.Username != "u" || got.Auth.Password != "p" || got.Auth.TTL != 10*time.Minute {
		t.Fatalf("unexpected auth: %+v", got.Auth)
	}
}
