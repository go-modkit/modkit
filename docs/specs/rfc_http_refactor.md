# RFC: HTTP Adapter Refactor

## Status: Proposed
## Proposed Date: 2026-02-05

## Overview
Refactor the `modkit/http` package to provide a more flexible, extensible, and idiomatic Go API for serving HTTP requests.

## Proposed Changes

### 1. Server Options Pattern
Replace the global `Serve` function's dependency on global variables with a `Server` struct and functional options.

```go
type Server struct {
    addr              string
    handler           http.Handler
    readHeaderTimeout time.Duration
    shutdownTimeout   time.Duration
    logger            logging.Logger
    // ...
}

type Option func(*Server)

func WithReadHeaderTimeout(d time.Duration) Option { ... }
func WithShutdownTimeout(d time.Duration) Option { ... }
func WithLogger(l logging.Logger) Option { ... }

func NewServer(addr string, handler http.Handler, opts ...Option) *Server { ... }
func (s *Server) ListenAndServe() error { ... }
```

### 2. Standardized Middleware
Formally define `Middleware` and provide helpers for chaining.

```go
type Middleware func(http.Handler) http.Handler

// Chain combines multiple middlewares into a single middleware.
func Chain(ms ...Middleware) Middleware {
    return func(next http.Handler) http.Handler {
        for i := len(ms) - 1; i >= 0; i-- {
            next = ms[i](next)
        }
        return next
    }
}
```

### 3. Enhanced Router Interface
Update the `Router` interface to be more explicit about its capabilities.

### 4. Problem Details (RFC 7807)
Introduce a standard error encoding mechanism that defaults to Problem Details but can be overridden.

## Alternatives Considered
- **Keeping `Serve` as is:** Too limiting for production use cases (custom timeouts, logging).
- **Reflection-based routing:** Rejected to maintain the "No Reflection" core principle of `modkit`.

## Success Criteria
- Existing tests pass with minimal changes.
- New functionality is fully covered by unit tests.
- Examples updated to show off the new API.
