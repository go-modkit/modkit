package main

import (
	"log/slog"

	modkitlogging "github.com/go-modkit/modkit/modkit/logging"
)

func logStartup(logger modkitlogging.Logger, addr string) {
	if logger == nil {
		logger = modkitlogging.NewNopLogger()
	}
	logger = logger.With(slog.String("scope", "api"))
	logger.Info("server starting",
		slog.String("addr", addr),
		slog.String("swagger", "http://localhost"+addr+"/docs/index.html"),
	)
}
