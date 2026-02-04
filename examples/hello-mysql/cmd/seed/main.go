package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aryeko/modkit/examples/hello-mysql/internal/platform/config"
	"github.com/aryeko/modkit/examples/hello-mysql/internal/platform/logging"
	"github.com/aryeko/modkit/examples/hello-mysql/internal/platform/mysql"
	"github.com/aryeko/modkit/examples/hello-mysql/internal/seed"
)

func main() {
	cfg := config.Load()
	logger := logging.New()
	logSeedDebug(logger, "starting seed")
	db, err := mysql.Open(cfg.MySQLDSN)
	if err != nil {
		logger.Error("open db failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	if err := seed.Seed(context.Background(), db); err != nil {
		logger.Error("seed failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logSeedComplete(logger)
}
