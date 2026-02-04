package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/aryeko/modkit/modkit/logging"
	"github.com/go-chi/chi/v5/middleware"
)

func RequestLogger(logger logging.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = logging.Nop()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			next.ServeHTTP(ww, r)
			duration := time.Since(start)

			logger.Info("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.Status()),
				slog.Duration("duration", duration),
			)
		})
	}
}
