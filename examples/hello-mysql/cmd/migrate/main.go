package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aryeko/modkit/examples/hello-mysql/internal/platform/config"
	"github.com/aryeko/modkit/examples/hello-mysql/internal/platform/logging"
	"github.com/aryeko/modkit/examples/hello-mysql/internal/platform/mysql"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()
	logger := logging.New()
	logMigrateDebug(logger, "starting migrations")

	db, err := mysql.Open(cfg.MySQLDSN)
	if err != nil {
		logger.Error("open db failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	if err := mysql.ApplyMigrations(ctx, db, "migrations"); err != nil {
		logger.Error("migrations failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logMigrateComplete(logger)
}
