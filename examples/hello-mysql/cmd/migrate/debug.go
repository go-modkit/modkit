package main

import (
	"log/slog"

	modkitlogging "github.com/go-modkit/modkit/modkit/logging"
)

func logMigrateDebug(logger modkitlogging.Logger, msg string) {
	if logger == nil {
		logger = modkitlogging.NewNopLogger()
	}
	logger.Debug(msg, slog.String("scope", "migrate"))
}
