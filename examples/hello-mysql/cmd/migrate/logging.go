package main

import (
	"log/slog"

	modkitlogging "github.com/aryeko/modkit/modkit/logging"
)

func logMigrateComplete(logger modkitlogging.Logger) {
	if logger == nil {
		logger = modkitlogging.Nop()
	}
	logger = logger.With(slog.String("scope", "migrate"))
	logger.Info("migrations complete")
}
