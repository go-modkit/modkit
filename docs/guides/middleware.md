# Middleware

Middleware in modkit uses Go's standard `http.Handler` pattern. There are no framework-specific abstractions—just plain functions that wrap handlers.

## What is Middleware?

Middleware is a function that takes an `http.Handler` and returns a new `http.Handler`:

```go
func MyMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Before the handler
        // ...
        
        next.ServeHTTP(w, r)  // Call the next handler
        
        // After the handler
        // ...
    })
}
```

## Applying Middleware

### Global Middleware

Apply to all routes when creating the router:

```go
router := mkhttp.NewRouter()

// Add global middleware
router.Use(loggingMiddleware)
router.Use(recoveryMiddleware)
router.Use(corsMiddleware)

// Register routes
mkhttp.RegisterRoutes(mkhttp.AsRouter(router), app.Controllers)
```

### Route Group Middleware

Apply to specific route groups within a controller:

```go
func (c *UsersController) RegisterRoutes(r mkhttp.Router) {
    // Public routes
    r.Handle(http.MethodGet, "/users", http.HandlerFunc(c.List))
    
    // Protected routes
    r.Group("/users", func(r mkhttp.Router) {
        r.Use(authMiddleware)
        r.Handle(http.MethodPost, "/", http.HandlerFunc(c.Create))
        r.Handle(http.MethodDelete, "/{id}", http.HandlerFunc(c.Delete))
    })
}
```

### Per-Route Middleware

Wrap individual handlers:

```go
r.Handle(http.MethodPost, "/admin/users", 
    adminOnly(http.HandlerFunc(c.CreateAdmin)))
```

## Common Middleware Patterns

### Logging

```go
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // Wrap response writer to capture status
            ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
            
            next.ServeHTTP(ww, r)
            
            logger.Info("request",
                "method", r.Method,
                "path", r.URL.Path,
                "status", ww.status,
                "duration", time.Since(start),
            )
        })
    }
}

type responseWriter struct {
    http.ResponseWriter
    status int
}

func (w *responseWriter) WriteHeader(code int) {
    w.status = code
    w.ResponseWriter.WriteHeader(code)
}
```

### Recovery (Panic Handler)

```go
func RecoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic: %v\n%s", err, debug.Stack())
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

### Request ID

```go
type contextKey string

const RequestIDKey contextKey = "request_id"

func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := r.Header.Get("X-Request-ID")
        if id == "" {
            id = uuid.New().String()
        }
        
        ctx := context.WithValue(r.Context(), RequestIDKey, id)
        w.Header().Set("X-Request-ID", id)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Usage in handlers
func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
    requestID := r.Context().Value(RequestIDKey).(string)
    // ...
}
```

### CORS

```go
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")
            
            for _, allowed := range allowedOrigins {
                if origin == allowed || allowed == "*" {
                    w.Header().Set("Access-Control-Allow-Origin", origin)
                    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
                    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
                    break
                }
            }
            
            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### Authentication

```go
func AuthMiddleware(validateToken func(string) (*User, error)) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractBearerToken(r)
            if token == "" {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }
            
            user, err := validateToken(token)
            if err != nil {
                http.Error(w, "invalid token", http.StatusUnauthorized)
                return
            }
            
            ctx := context.WithValue(r.Context(), UserKey, user)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func extractBearerToken(r *http.Request) string {
    auth := r.Header.Get("Authorization")
    if !strings.HasPrefix(auth, "Bearer ") {
        return ""
    }
    return strings.TrimPrefix(auth, "Bearer ")
}
```

### Rate Limiting

```go
func RateLimitMiddleware(rps int) func(http.Handler) http.Handler {
    limiter := rate.NewLimiter(rate.Limit(rps), rps)
    
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

### Timeout

```go
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.TimeoutHandler(next, timeout, "request timeout")
    }
}
```

## Middleware Order

Middleware executes in the order it's added. The first middleware wraps all subsequent ones:

```go
router.Use(requestID)    // 1st: runs first (outermost)
router.Use(logging)      // 2nd: runs second
router.Use(recovery)     // 3rd: runs third
router.Use(auth)         // 4th: runs last (innermost)
```

Request flow:
```text
Request → requestID → logging → recovery → auth → Handler
Response ← requestID ← logging ← recovery ← auth ← Handler
```

**Recommended order:**
1. Request ID (for tracing)
2. Logging (to log all requests)
3. Recovery (to catch panics)
4. CORS (for cross-origin requests)
5. Rate limiting
6. Authentication
7. Authorization

## Middleware as Providers

For middleware that needs dependencies, register as a provider:

```go
func (m *AppModule) Definition() module.ModuleDef {
    return module.ModuleDef{
        Name: "app",
        Providers: []module.ProviderDef{
            {
                Token: "middleware.auth",
                Build: func(r module.Resolver) (any, error) {
                    userSvc, _ := module.Get[UserService](r, TokenUserService)
                    return AuthMiddleware(userSvc.ValidateToken), nil
                },
            },
        },
    }
}
```

Then retrieve and apply in your startup code:

```go
app, _ := kernel.Bootstrap(&AppModule{})

authMW, _ := module.Get[func(http.Handler) http.Handler](app, "middleware.auth")
router.Use(authMW)
```

## Tips

- Keep middleware focused on a single responsibility
- Use closures to inject dependencies
- Handle errors consistently (don't mix `http.Error` and JSON responses)
- Add context values for cross-cutting data (request ID, user, etc.)
- Test middleware in isolation using `httptest`
- Consider using existing middleware libraries (chi, gorilla) for common patterns

## See example

- [Middleware package](../../examples/hello-mysql/internal/middleware/)
- [CORS middleware](../../examples/hello-mysql/internal/middleware/cors.go)
- [Rate limit middleware](../../examples/hello-mysql/internal/middleware/rate_limit.go)
- [Timing middleware](../../examples/hello-mysql/internal/middleware/timing.go)
- [Route group + middleware order wiring](../../examples/hello-mysql/internal/httpserver/server.go)
