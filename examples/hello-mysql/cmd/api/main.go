package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-modkit/modkit/examples/hello-mysql/docs"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/httpserver"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/lifecycle"
	configmodule "github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/config"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/platform/logging"
	modkithttp "github.com/go-modkit/modkit/modkit/http"
	"github.com/go-modkit/modkit/modkit/module"
)

// @title hello-mysql API
// @version 0.1
// @description Example modkit service with MySQL.
// @BasePath /api/v1
func main() {
	boot, handler, err := httpserver.BuildAppHandler()
	if err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}

	addr, err := module.Get[string](boot, configmodule.TokenHTTPAddr)
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	logger := logging.New()
	logStartup(logger, addr)

	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer func() {
		signal.Stop(sigCh)
		close(sigCh)
	}()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	hooks := buildShutdownHooks(boot)
	if err := runServer(modkithttp.ShutdownTimeout, server, sigCh, errCh, hooks); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

type shutdownServer interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

type appLifecycle interface {
	CleanupHooks() []func(context.Context) error
	CloseContext(context.Context) error
}

func buildShutdownHooks(app appLifecycle) []lifecycle.CleanupHook {
	hooks := lifecycle.FromFuncs(app.CleanupHooks())
	return append([]lifecycle.CleanupHook{app.CloseContext}, hooks...)
}

func runServer(shutdownTimeout time.Duration, server shutdownServer, sigCh <-chan os.Signal, errCh <-chan error, hooks []lifecycle.CleanupHook) error {
	select {
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case <-sigCh:
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		shutdownErr := lifecycle.ShutdownServer(ctx, server, hooks)

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
