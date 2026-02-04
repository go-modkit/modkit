# Interceptors

Interceptors in NestJS are used for request/response transformation, caching, and timing. In Go and modkit, these patterns are implemented as standard `http.Handler` middleware. This guide shows Go-idiomatic equivalents you can compose in the same way.

## Timing Middleware

Measure request duration with a wrapper:

```go
func TimingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            next.ServeHTTP(w, r)
            logger.Info("request completed",
                "method", r.Method,
                "path", r.URL.Path,
                "duration", time.Since(start),
            )
        })
    }
}
```

## Response Status Capture

Capture the response status by wrapping `http.ResponseWriter`:

```go
type statusWriter struct {
    http.ResponseWriter
    status int
}

func (w *statusWriter) WriteHeader(code int) {
    w.status = code
    w.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
            next.ServeHTTP(sw, r)
            logger.Info("request",
                "method", r.Method,
                "path", r.URL.Path,
                "status", sw.status,
            )
        })
    }
}
```

## Caching Middleware

Define a minimal cache interface and wrap `GET` responses:

```go
import (
    "net/http"
    "net/http/httptest"
    "time"
)

type Cache interface {
    Get(key string) ([]byte, bool)
    Set(key string, value []byte, ttl time.Duration)
}

func CacheMiddleware(cache Cache, ttl time.Duration) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if r.Method != http.MethodGet {
                next.ServeHTTP(w, r)
                return
            }

            key := r.URL.String()
            if cached, ok := cache.Get(key); ok {
                w.Write(cached)
                return
            }

            rec := httptest.NewRecorder()
            next.ServeHTTP(rec, r)

            cache.Set(key, rec.Body.Bytes(), ttl)
            for k, v := range rec.Header() {
                w.Header()[k] = v
            }
            w.WriteHeader(rec.Code)
            w.Write(rec.Body.Bytes())
        })
    }
}
```

## Response Transformation

Capture and transform a response before returning it:

```go
func WrapResponse(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        rec := httptest.NewRecorder()
        next.ServeHTTP(rec, r)

        body := rec.Body.Bytes()
        wrapped := map[string]any{"data": json.RawMessage(body)}

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(wrapped)
    })
}
```
