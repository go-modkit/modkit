package main

import (
	"log/slog"

	modkitlogging "github.com/go-modkit/modkit/modkit/logging"
)

func logMigrateComplete(logger modkitlogging.Logger) {
	if logger == nil {
		logger = modkitlogging.NewNopLogger()
	}
	logger = logger.With(slog.String("scope", "migrate"))
	logger.Info("migrations complete")
}
