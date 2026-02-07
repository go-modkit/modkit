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
