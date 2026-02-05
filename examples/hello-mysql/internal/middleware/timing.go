package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	modkitlogging "github.com/go-modkit/modkit/modkit/logging"
)

func NewTiming(logger modkitlogging.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = modkitlogging.NewNopLogger()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			next.ServeHTTP(ww, r)

			logger.Info("request timing",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration", time.Since(start),
			)
		})
	}
}
