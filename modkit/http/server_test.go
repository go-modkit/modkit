package http

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"syscall"
	"testing"
	"time"
)

func TestServe_ReturnsErrorWhenServerFailsToStart(t *testing.T) {
	originalListen := listenAndServe
	originalShutdown := shutdownServer
	defer func() {
		listenAndServe = originalListen
		shutdownServer = originalShutdown
	}()

	var gotAddr string
	var gotHandler http.Handler
	listenAndServe = func(server *http.Server) error {
		gotAddr = server.Addr
		gotHandler = server.Handler
		return errors.New("boom")
	}
	shutdownServer = func(_ context.Context, _ *http.Server) error {
		return nil
	}

	router := NewRouter()
	err := Serve("127.0.0.1:12345", router)

	if gotAddr != "127.0.0.1:12345" {
		t.Fatalf("expected addr %q, got %q", "127.0.0.1:12345", gotAddr)
	}
	if gotHandler != router {
		t.Fatalf("expected handler to be router")
	}
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected error from listenAndServe, got %v", err)
	}
}

func TestServe_HandlesSignals_ReturnsNil(t *testing.T) {
	for _, tt := range []struct {
		name string
		sig  os.Signal
	}{
		{name: "SIGINT", sig: os.Interrupt},
		{name: "SIGTERM", sig: syscall.SIGTERM},
	} {
		t.Run(tt.name, func(t *testing.T) {
			originalListen := listenAndServe
			originalShutdown := shutdownServer
			defer func() {
				listenAndServe = originalListen
				shutdownServer = originalShutdown
			}()

			serveStarted := make(chan struct{})
			shutdownRequested := make(chan struct{})

			listenAndServe = func(_ *http.Server) error {
				close(serveStarted)
				<-shutdownRequested
				return http.ErrServerClosed
			}
			shutdownServer = func(_ context.Context, _ *http.Server) error {
				close(shutdownRequested)
				return nil
			}

			errCh := make(chan error, 1)
			go func() {
				errCh <- Serve("127.0.0.1:12345", NewRouter())
			}()

			<-serveStarted
			if tt.sig == syscall.SIGTERM && runtime.GOOS == "windows" {
				t.Skip("SIGTERM not supported on Windows")
			}
			proc, err := os.FindProcess(os.Getpid())
			if err != nil {
				t.Fatalf("failed to find process: %v", err)
			}
			if err := proc.Signal(tt.sig); err != nil {
				t.Fatalf("failed to send signal: %v", err)
			}

			if err := <-errCh; err != nil {
				t.Fatalf("expected nil on clean shutdown, got %v", err)
			}
		})
	}
}

func TestServe_ShutdownWaitsForInFlightRequest(t *testing.T) {
	originalListen := listenAndServe
	originalShutdown := shutdownServer
	originalTimeout := ShutdownTimeout
	defer func() {
		listenAndServe = originalListen
		shutdownServer = originalShutdown
		ShutdownTimeout = originalTimeout
	}()

	ShutdownTimeout = 2 * time.Second

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	addr := ln.Addr().String()

	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(requestStarted)
		<-releaseRequest
		w.WriteHeader(http.StatusOK)
	})

	listenAndServe = func(server *http.Server) error {
		return server.Serve(ln)
	}
	shutdownServer = func(ctx context.Context, server *http.Server) error {
		return server.Shutdown(ctx)
	}

	serveErrCh := make(chan error, 1)
	go func() {
		serveErrCh <- Serve(addr, handler)
	}()

	clientErrCh := make(chan error, 1)
	go func() {
		resp, err := http.Get("http://" + addr)
		if err != nil {
			clientErrCh <- err
			return
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		clientErrCh <- nil
	}()

	<-requestStarted

	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("failed to find process: %v", err)
	}
	if err := proc.Signal(os.Interrupt); err != nil {
		t.Fatalf("failed to send signal: %v", err)
	}

	select {
	case err := <-serveErrCh:
		t.Fatalf("expected Serve to wait for in-flight request, got %v", err)
	case <-time.After(200 * time.Millisecond):
	}

	close(releaseRequest)

	if err := <-serveErrCh; err != nil {
		t.Fatalf("expected nil on clean shutdown, got %v", err)
	}
	if err := <-clientErrCh; err != nil {
		t.Fatalf("request failed: %v", err)
	}
}
