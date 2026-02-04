package main

import (
	"log"

	_ "github.com/aryeko/modkit/examples/hello-mysql/docs"
	"github.com/aryeko/modkit/examples/hello-mysql/internal/httpserver"
	"github.com/aryeko/modkit/examples/hello-mysql/internal/modules/app"
	"github.com/aryeko/modkit/examples/hello-mysql/internal/platform/config"
	modkithttp "github.com/aryeko/modkit/modkit/http"
)

// @title hello-mysql API
// @version 0.1
// @description Example modkit service with MySQL.
// @BasePath /
func main() {
	cfg := config.Load()
	handler, err := httpserver.BuildHandler(app.Options{HTTPAddr: cfg.HTTPAddr, MySQLDSN: cfg.MySQLDSN})
	if err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}

	if err := modkithttp.Serve(cfg.HTTPAddr, handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
