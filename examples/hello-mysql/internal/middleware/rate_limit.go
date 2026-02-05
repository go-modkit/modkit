package middleware

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/go-modkit/modkit/examples/hello-mysql/internal/httpapi"
	"golang.org/x/time/rate"
)

type RateLimitConfig struct {
	RequestsPerSecond float64
	Burst             int
}

func NewRateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Limit(cfg.RequestsPerSecond), cfg.Burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reservation := limiter.Reserve()
			if !reservation.OK() {
				writeRateLimitExceeded(w, r, time.Second)
				return
			}

			delay := reservation.Delay()
			if delay > 0 {
				reservation.Cancel()
				writeRateLimitExceeded(w, r, delay)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeRateLimitExceeded(w http.ResponseWriter, r *http.Request, delay time.Duration) {
	w.Header().Set("Retry-After", retryAfterValue(delay))
	httpapi.WriteProblem(w, r, http.StatusTooManyRequests, "rate limit exceeded")
}

func retryAfterValue(delay time.Duration) string {
	seconds := int64(math.Ceil(delay.Seconds()))
	if seconds < 1 {
		seconds = 1
	}
	return strconv.FormatInt(seconds, 10)
}
