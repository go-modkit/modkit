package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-modkit/modkit/examples/hello-mysql/internal/lifecycle"
)

type stubServer struct {
	shutdownCalled bool
	shutdownErr    error
	shutdownCh     chan struct{}
}

type stubApp struct {
	closeCalled   bool
	cleanupCalled bool
	order         []string
	closeErr      error
}

func (s *stubApp) CleanupHooks() []func(context.Context) error {
	return []func(context.Context) error{
		func(ctx context.Context) error {
			s.cleanupCalled = true
			s.order = append(s.order, "cleanup")
			return nil
		},
	}
}

func (s *stubApp) CloseContext(ctx context.Context) error {
	s.closeCalled = true
	s.order = append(s.order, "close")
	return s.closeErr
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
	app := &stubApp{}
	sigCh := make(chan os.Signal, 1)
	errCh := make(chan error, 1)
	hooks := buildShutdownHooks(app)

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
	if !app.cleanupCalled {
		t.Fatal("expected cleanup hook to be called")
	}
	if !app.closeCalled {
		t.Fatal("expected CloseContext to be called")
	}
}

func TestBuildShutdownHooks_AppCloseRunsLast(t *testing.T) {
	app := &stubApp{}

	hooks := buildShutdownHooks(app)
	if len(hooks) != 2 {
		t.Fatalf("expected 2 hooks, got %d", len(hooks))
	}

	if err := lifecycle.RunCleanup(context.Background(), hooks); err != nil {
		t.Fatalf("unexpected cleanup error: %v", err)
	}

	if !app.cleanupCalled {
		t.Fatal("expected cleanup hook to run")
	}
	if !app.closeCalled {
		t.Fatal("expected CloseContext to run")
	}
	if len(app.order) != 2 || app.order[0] != "cleanup" || app.order[1] != "close" {
		t.Fatalf("expected cleanup then close, got %v", app.order)
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

func TestRunServer_ShutdownReturnsCloseError(t *testing.T) {
	server := &stubServer{shutdownCh: make(chan struct{})}
	app := &stubApp{closeErr: errors.New("close failed")}
	hooks := buildShutdownHooks(app)
	sigCh := make(chan os.Signal, 1)
	errCh := make(chan error, 1)
	go func() {
		<-server.shutdownCh
		errCh <- http.ErrServerClosed
	}()
	sigCh <- os.Interrupt

	err := runServer(50*time.Millisecond, server, sigCh, errCh, hooks)
	if err == nil || err.Error() != "close failed" {
		t.Fatalf("expected close error, got %v", err)
	}
}

func TestLoadAppOptions_Defaults(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("MYSQL_DSN", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("JWT_ISSUER", "")
	t.Setenv("JWT_TTL", "")
	t.Setenv("AUTH_USERNAME", "")
	t.Setenv("AUTH_PASSWORD", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	t.Setenv("CORS_ALLOWED_METHODS", "")
	t.Setenv("CORS_ALLOWED_HEADERS", "")
	t.Setenv("RATE_LIMIT_PER_SECOND", "")
	t.Setenv("RATE_LIMIT_BURST", "")

	got, err := loadAppOptions()
	if err != nil {
		t.Fatalf("loadAppOptions failed: %v", err)
	}

	if got.HTTPAddr != ":8080" {
		t.Fatalf("unexpected HTTPAddr: %q", got.HTTPAddr)
	}
	if got.MySQLDSN == "" {
		t.Fatalf("expected default MySQLDSN")
	}
	if got.Auth.TTL != time.Hour {
		t.Fatalf("unexpected JWT TTL default: %v", got.Auth.TTL)
	}
	if got.RateLimitPerSecond != 5 || got.RateLimitBurst != 10 {
		t.Fatalf("unexpected rate limit defaults: %v/%d", got.RateLimitPerSecond, got.RateLimitBurst)
	}
}

func TestLoadAppOptions_WithOverrides(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":19090")
	t.Setenv("MYSQL_DSN", "user:pass@tcp(host:3307)/db")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("JWT_ISSUER", "issuer")
	t.Setenv("JWT_TTL", "2h")
	t.Setenv("AUTH_USERNAME", "alice")
	t.Setenv("AUTH_PASSWORD", "pw")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.example,https://b.example")
	t.Setenv("CORS_ALLOWED_METHODS", "GET,POST")
	t.Setenv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization")
	t.Setenv("RATE_LIMIT_PER_SECOND", "7.5")
	t.Setenv("RATE_LIMIT_BURST", "25")

	got, err := loadAppOptions()
	if err != nil {
		t.Fatalf("loadAppOptions failed: %v", err)
	}

	if got.HTTPAddr != ":19090" || got.MySQLDSN != "user:pass@tcp(host:3307)/db" {
		t.Fatalf("unexpected core fields: %+v", got)
	}
	if got.Auth.TTL != 2*time.Hour || got.Auth.Secret != "secret" || got.Auth.Issuer != "issuer" {
		t.Fatalf("unexpected auth fields: %+v", got.Auth)
	}
	if len(got.CORSAllowedOrigins) != 2 || got.CORSAllowedOrigins[0] != "https://a.example" {
		t.Fatalf("unexpected CORS origins: %+v", got.CORSAllowedOrigins)
	}
	if got.RateLimitPerSecond != 7.5 || got.RateLimitBurst != 25 {
		t.Fatalf("unexpected rate limits: %v/%d", got.RateLimitPerSecond, got.RateLimitBurst)
	}
}

func TestLoadAppOptions_InvalidNonPositiveJWTTTL(t *testing.T) {
	t.Setenv("JWT_TTL", "0s")

	_, err := loadAppOptions()
	if err == nil {
		t.Fatalf("expected error for non-positive JWT_TTL")
	}
}
