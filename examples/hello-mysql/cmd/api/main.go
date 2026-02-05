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
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/app"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/auth"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/platform/config"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/platform/logging"
	modkithttp "github.com/go-modkit/modkit/modkit/http"
)

// @title hello-mysql API
// @version 0.1
// @description Example modkit service with MySQL.
// @BasePath /
func main() {
	cfg := config.Load()
	jwtTTL := parseJWTTTL(cfg.JWTTTL)

	boot, handler, err := httpserver.BuildAppHandler(buildAppOptions(cfg, jwtTTL))
	if err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}

	logger := logging.New()
	logStartup(logger, cfg.HTTPAddr)

	server := &http.Server{
		Addr:    cfg.HTTPAddr,
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

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	case <-sigCh:
		ctx, cancel := context.WithTimeout(context.Background(), modkithttp.ShutdownTimeout)
		defer cancel()

		hooks := lifecycle.FromFuncs(boot.CleanupHooks())
		shutdownErr := lifecycle.ShutdownServer(ctx, server, hooks)

		err := <-errCh
		if err == http.ErrServerClosed {
			err = nil
		}
		if shutdownErr != nil {
			log.Fatalf("shutdown failed: %v", shutdownErr)
		}
		if err != nil {
			log.Fatalf("server failed: %v", err)
		}
	}
}

func buildAppOptions(cfg config.Config, jwtTTL time.Duration) app.Options {
	return app.Options{
		HTTPAddr: cfg.HTTPAddr,
		MySQLDSN: cfg.MySQLDSN,
		Auth:     buildAuthConfig(cfg, jwtTTL),
	}
}

func buildAuthConfig(cfg config.Config, jwtTTL time.Duration) auth.Config {
	return auth.Config{
		Secret:   cfg.JWTSecret,
		Issuer:   cfg.JWTIssuer,
		TTL:      jwtTTL,
		Username: cfg.AuthUsername,
		Password: cfg.AuthPassword,
	}
}

func parseJWTTTL(raw string) time.Duration {
	ttl, err := time.ParseDuration(raw)
	if err != nil {
		log.Printf("invalid JWT_TTL %q, using 1h: %v", raw, err)
		return time.Hour
	}
	if ttl <= 0 {
		log.Printf("invalid JWT_TTL %q, using 1h: non-positive duration", raw)
		return time.Hour
	}
	return ttl
}
