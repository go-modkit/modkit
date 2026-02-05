package main

import (
	"log/slog"

	modkitlogging "github.com/aryeko/modkit/modkit/logging"
)

func logSeedDebug(logger modkitlogging.Logger, msg string) {
	if logger == nil {
		logger = modkitlogging.NewNopLogger()
	}
	logger.Debug(msg, slog.String("scope", "seed"))
}
