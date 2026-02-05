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
	mwconfig "github.com/go-modkit/modkit/examples/hello-mysql/internal/config"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/httpserver"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/lifecycle"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/app"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/auth"
	platformconfig "github.com/go-modkit/modkit/examples/hello-mysql/internal/platform/config"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/platform/logging"
	modkithttp "github.com/go-modkit/modkit/modkit/http"
)

// @title hello-mysql API
// @version 0.1
// @description Example modkit service with MySQL.
// @BasePath /api/v1
func main() {
	cfg := platformconfig.Load()
	mwCfg := mwconfig.Load()
	jwtTTL := parseJWTTTL(cfg.JWTTTL)

	boot, handler, err := httpserver.BuildAppHandler(buildAppOptions(cfg, mwCfg, jwtTTL))
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

	hooks := buildShutdownHooks(boot)
	if err := runServer(modkithttp.ShutdownTimeout, server, sigCh, errCh, hooks); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func buildAppOptions(cfg platformconfig.Config, mwCfg mwconfig.Config, jwtTTL time.Duration) app.Options {
	return app.Options{
		HTTPAddr:           cfg.HTTPAddr,
		MySQLDSN:           cfg.MySQLDSN,
		Auth:               buildAuthConfig(cfg, jwtTTL),
		CORSAllowedOrigins: mwCfg.CORSAllowedOrigins,
		CORSAllowedMethods: mwCfg.CORSAllowedMethods,
		CORSAllowedHeaders: mwCfg.CORSAllowedHeaders,
		RateLimitPerSecond: mwCfg.RateLimitPerSecond,
		RateLimitBurst:     mwCfg.RateLimitBurst,
	}
}

func buildAuthConfig(cfg platformconfig.Config, jwtTTL time.Duration) auth.Config {
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
