package main

import (
	"context"
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
