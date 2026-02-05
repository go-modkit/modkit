package http

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ShutdownTimeout controls how long the server will wait for in-flight requests
// to finish after receiving a shutdown signal.
var ShutdownTimeout = 30 * time.Second

var listenAndServe = func(server *http.Server) error {
	return server.ListenAndServe()
}

var shutdownServer = func(ctx context.Context, server *http.Server) error {
	return server.Shutdown(ctx)
}

// Serve starts an HTTP server on the given address using the provided handler.
// It handles SIGINT and SIGTERM for graceful shutdown.
func Serve(addr string, handler http.Handler) error {
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 15 * time.Second,
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer func() {
		signal.Stop(sigCh)
		close(sigCh)
	}()

	errCh := make(chan error, 1)
	go func() {
		errCh <- listenAndServe(server)
	}()

	select {
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case <-sigCh:
		ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
		defer cancel()

		shutdownErr := shutdownServer(ctx, server)
		err := <-errCh
		if err == http.ErrServerClosed {
			err = nil
		}
		if shutdownErr != nil {
			return shutdownErr
		}
		return err
	}
}
