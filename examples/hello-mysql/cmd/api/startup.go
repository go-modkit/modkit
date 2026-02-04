package main

import (
	"log/slog"

	modkitlogging "github.com/aryeko/modkit/modkit/logging"
)

func logStartup(logger modkitlogging.Logger, addr string) {
	if logger == nil {
		logger = modkitlogging.Nop()
	}
	logger.Info("server starting",
		slog.String("addr", addr),
		slog.String("swagger", "http://localhost"+addr+"/docs/index.html"),
	)
}
