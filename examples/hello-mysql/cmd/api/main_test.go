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
}

func (s *stubServer) ListenAndServe() error {
	return nil
}

func (s *stubServer) Shutdown(ctx context.Context) error {
	s.shutdownCalled = true
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

	sigCh <- os.Interrupt
	errCh <- http.ErrServerClosed

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
