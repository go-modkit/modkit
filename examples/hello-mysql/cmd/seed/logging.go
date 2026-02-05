package main

import (
	"log/slog"

	modkitlogging "github.com/go-modkit/modkit/modkit/logging"
)

func logSeedComplete(logger modkitlogging.Logger) {
	if logger == nil {
		logger = modkitlogging.NewNopLogger()
	}
	logger = logger.With(slog.String("scope", "seed"))
	logger.Info("seed complete")
}
