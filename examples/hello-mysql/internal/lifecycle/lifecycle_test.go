package lifecycle

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestShutdown_InvokesCleanupHooksInLIFO(t *testing.T) {
	calls := make([]string, 0, 2)
	hooks := []CleanupHook{
		func(ctx context.Context) error {
			calls = append(calls, "first")
			return nil
		},
		func(ctx context.Context) error {
			calls = append(calls, "second")
			return nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := RunCleanup(ctx, hooks); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	if got, want := strings.Join(calls, ","), "second,first"; got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestRunCleanup_JoinsErrorsAndSkipsNil(t *testing.T) {
	calls := make([]string, 0, 2)
	errFirst := errors.New("first")
	errSecond := errors.New("second")
	hooks := []CleanupHook{
		nil,
		func(ctx context.Context) error {
			calls = append(calls, "first")
			return errFirst
		},
		func(ctx context.Context) error {
			calls = append(calls, "second")
			return errSecond
		},
	}

	if err := RunCleanup(context.Background(), hooks); err == nil {
		t.Fatal("expected error, got nil")
	} else if !errors.Is(err, errFirst) || !errors.Is(err, errSecond) {
		t.Fatalf("expected joined errors, got %v", err)
	}

	if got, want := strings.Join(calls, ","), "second,first"; got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestFromFuncs_WrapsFuncsAndSkipsNil(t *testing.T) {
	calls := make([]string, 0, 2)
	fnFirst := func(ctx context.Context) error {
		calls = append(calls, "first")
		return nil
	}
	fnSecond := func(ctx context.Context) error {
		calls = append(calls, "second")
		return nil
	}

	hooks := FromFuncs([]func(context.Context) error{fnFirst, nil, fnSecond})
	if len(hooks) != 3 {
		t.Fatalf("expected 3 hooks, got %d", len(hooks))
	}
	if hooks[0] == nil || hooks[2] == nil {
		t.Fatal("expected non-nil hooks for non-nil funcs")
	}
	if hooks[1] != nil {
		t.Fatal("expected nil hook for nil func")
	}

	if err := RunCleanup(context.Background(), hooks); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	if got, want := strings.Join(calls, ","), "second,first"; got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestShutdown_WaitsForInFlightRequest(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})
	cleanupCalled := make(chan struct{})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		w.WriteHeader(http.StatusOK)
		close(done)
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	server := &http.Server{Handler: handler}
	go func() {
		_ = server.Serve(ln)
	}()

	reqDone := make(chan struct{})
	go func() {
		_, _ = http.Get("http://" + ln.Addr().String())
		close(reqDone)
	}()

	<-started

	hooks := []CleanupHook{
		func(ctx context.Context) error {
			close(cleanupCalled)
			return nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	shutdownDone := make(chan error, 1)
	go func() {
		shutdownDone <- ShutdownServer(ctx, server, hooks)
	}()

	select {
	case <-cleanupCalled:
		t.Fatal("cleanup ran before in-flight request completed")
	default:
	}

	close(release)
	<-done
	<-reqDone

	select {
	case <-cleanupCalled:
	case <-time.After(time.Second):
		t.Fatal("cleanup did not run after in-flight request completed")
	}

	if err := <-shutdownDone; err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}
}

type stubServer struct {
	err    error
	called bool
}

func (s *stubServer) Shutdown(ctx context.Context) error {
	s.called = true
	return s.err
}

func TestShutdownServer_ReturnsShutdownErrorAndRunsCleanup(t *testing.T) {
	shutdownErr := errors.New("shutdown failed")
	server := &stubServer{err: shutdownErr}
	cleanupCalled := false
	hooks := []CleanupHook{
		func(ctx context.Context) error {
			cleanupCalled = true
			return nil
		},
	}

	if err := ShutdownServer(context.Background(), server, hooks); !errors.Is(err, shutdownErr) {
		t.Fatalf("expected shutdown error, got %v", err)
	}
	if !server.called {
		t.Fatal("expected shutdown to be called")
	}
	if !cleanupCalled {
		t.Fatal("expected cleanup to run even when shutdown fails")
	}
}

func TestShutdownServer_ReturnsCleanupErrorWhenShutdownOk(t *testing.T) {
	cleanupErr := errors.New("cleanup failed")
	server := &stubServer{}
	hooks := []CleanupHook{
		func(ctx context.Context) error {
			return cleanupErr
		},
	}

	if err := ShutdownServer(context.Background(), server, hooks); !errors.Is(err, cleanupErr) {
		t.Fatalf("expected cleanup error, got %v", err)
	}
}
