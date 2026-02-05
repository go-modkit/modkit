package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware" //nolint:goimports // False positive
	"github.com/go-modkit/modkit/modkit/logging"
)

// RequestLogger returns an HTTP middleware that logs each request with method, path, status, and duration.
func RequestLogger(logger logging.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = logging.NewNopLogger()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			next.ServeHTTP(ww, r)
			duration := time.Since(start)

			logger.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration", duration,
			)
		})
	}
}
