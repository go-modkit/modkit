package main

import (
	"log"
	"time"

	_ "github.com/go-modkit/modkit/examples/hello-mysql/docs"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/httpserver"
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

	authCfg := auth.Config{
		Secret:   cfg.JWTSecret,
		Issuer:   cfg.JWTIssuer,
		TTL:      jwtTTL,
		Username: cfg.AuthUsername,
		Password: cfg.AuthPassword,
	}

	handler, err := httpserver.BuildHandler(app.Options{
		HTTPAddr: cfg.HTTPAddr,
		MySQLDSN: cfg.MySQLDSN,
		Auth:     authCfg,
	})
	if err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}

	logger := logging.New()
	logStartup(logger, cfg.HTTPAddr)

	if err := modkithttp.Serve(cfg.HTTPAddr, handler); err != nil {
		log.Fatalf("server failed: %v", err)
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
