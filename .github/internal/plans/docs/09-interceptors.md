# D9: Interceptors Guide

**Status:** 🔴 Not started  
**Type:** New guide  
**NestJS Equivalent:** Interceptors

---

## Goal

Document Go‑idiomatic request/response interception using middleware and handler wrappers.

## Why Different from NestJS

NestJS interceptors use RxJS observables to wrap handler execution for logging, caching, and response transformation. In Go, this is achieved with:
- Middleware for request/response wrapping
- Response writers for capturing output
- Handler wrappers for timing/logging

## Files to Create/Modify

- Create: `docs/guides/interceptors.md`
- Modify: `README.md` (add link)

---

## Task 1: Create interceptors guide

**Files:**
- Create: `docs/guides/interceptors.md`

### Step 1: Draft the guide

Include:

1. **Explanation** — middleware/wrappers are the Go equivalent
2. **Timing middleware** — measure request duration
3. **Response capture** — wrap ResponseWriter to capture status
4. **Caching middleware** — simple caching pattern

Suggested structure:

```markdown
# Interceptors

In Go, interceptor patterns are implemented using middleware and response wrappers.

## Timing Middleware

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

To capture the response status code:

```go
type statusWriter struct {
    http.ResponseWriter
    status int
}

func (w *statusWriter) WriteHeader(code int) {
    w.status = code
    w.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
        next.ServeHTTP(sw, r)
        log.Printf("%s %s -> %d", r.Method, r.URL.Path, sw.status)
    })
}
```

## Caching Middleware

```go
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

For transforming responses, capture and modify:

```go
func WrapResponse(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        rec := httptest.NewRecorder()
        next.ServeHTTP(rec, r)
        
        // Transform the response
        body := rec.Body.Bytes()
        wrapped := map[string]any{"data": json.RawMessage(body)}
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(wrapped)
    })
}
```
```

### Step 2: Commit

```bash
git add docs/guides/interceptors.md
git commit -m "docs: add interceptors guide"
```

---

## Task 2: Link guide from README

**Files:**
- Modify: `README.md`

### Step 1: Add interceptors guide to the Guides list

### Step 2: Commit

```bash
git add README.md
git commit -m "docs: link interceptors guide"
```
