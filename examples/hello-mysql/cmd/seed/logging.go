package main

import (
	"log/slog"

	modkitlogging "github.com/aryeko/modkit/modkit/logging"
)

func logSeedComplete(logger modkitlogging.Logger) {
	if logger == nil {
		logger = modkitlogging.Nop()
	}
	logger = logger.With(slog.String("scope", "seed"))
	logger.Info("seed complete")
}
