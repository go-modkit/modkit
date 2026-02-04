# C6: Add Graceful Shutdown to Serve

**Status:** 🔴 Not started  
**Type:** Code change  
**Priority:** Medium

---

## Motivation

Documentation states that `mkhttp.Serve` handles SIGINT/SIGTERM automatically for graceful shutdown:

> `mkhttp.Serve` handles SIGINT/SIGTERM automatically.
> — docs/faq.md

However, the actual implementation is simply:

```go
func Serve(addr string, handler http.Handler) error {
    return listenAndServe(addr, handler)
}
```

This is just a passthrough to `http.ListenAndServe` with no signal handling.

---

## Assumptions

1. Graceful shutdown should wait for in-flight requests to complete
2. A reasonable default timeout (e.g., 30 seconds) is acceptable
3. The function signature can remain the same (no config options needed for MVP)
4. SIGINT and SIGTERM should both trigger shutdown

---

## Requirements

### R1: Handle SIGINT and SIGTERM

The server should listen for these signals and initiate shutdown.

### R2: Graceful shutdown with timeout

Use `http.Server.Shutdown()` to allow in-flight requests to complete, with a timeout to prevent indefinite waiting.

### R3: Return appropriate errors

- Return `nil` on clean shutdown via signal
- Return error if server fails to start
- Return error if shutdown times out

### R4: Maintain testability

Keep the existing test hook (`listenAndServe` variable) or add new hooks for testing shutdown behavior.

---

## Files to Modify

| File | Change |
|------|--------|
| `modkit/http/server.go` | Implement graceful shutdown |
| `modkit/http/server_test.go` | Add tests for shutdown behavior |

---

## Implementation

### Step 1: Update server.go

Replace the simple passthrough with proper shutdown handling:

```go
package http

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)

// ShutdownTimeout is the maximum time to wait for in-flight requests during shutdown.
var ShutdownTimeout = 30 * time.Second

// Serve starts an HTTP server on the given address and blocks until shutdown.
// It handles SIGINT and SIGTERM for graceful shutdown, waiting for in-flight
// requests to complete before returning.
func Serve(addr string, handler http.Handler) error {
    server := &http.Server{
        Addr:    addr,
        Handler: handler,
    }

    // Channel to receive server errors
    errCh := make(chan error, 1)
    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            errCh <- err
        }
        close(errCh)
    }()

    // Wait for interrupt signal
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    select {
    case err := <-errCh:
        // Server failed to start
        return err
    case <-sigCh:
        // Received shutdown signal
    }

    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        return err
    }

    // Check if there was a server error during shutdown
    if err := <-errCh; err != nil {
        return err
    }

    return nil
}
```

### Step 2: Add ServeWithContext for advanced use cases (optional)

For users who need more control:

```go
// ServeWithContext starts an HTTP server that shuts down when the context is canceled.
func ServeWithContext(ctx context.Context, addr string, handler http.Handler) error {
    server := &http.Server{
        Addr:    addr,
        Handler: handler,
    }

    errCh := make(chan error, 1)
    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            errCh <- err
        }
        close(errCh)
    }()

    select {
    case err := <-errCh:
        return err
    case <-ctx.Done():
    }

    shutdownCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
    defer cancel()

    if err := server.Shutdown(shutdownCtx); err != nil {
        return err
    }

    if err := <-errCh; err != nil {
        return err
    }

    return nil
}
```

---

## Validation

### Unit Tests

Add to `modkit/http/server_test.go`:

```go
func TestServeGracefulShutdown(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping in short mode")
    }

    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    // Start server in goroutine
    errCh := make(chan error, 1)
    go func() {
        errCh <- Serve(":0", handler) // :0 picks random available port
    }()

    // Give server time to start
    time.Sleep(100 * time.Millisecond)

    // Send SIGINT to self
    p, _ := os.FindProcess(os.Getpid())
    p.Signal(syscall.SIGINT)

    // Should shut down cleanly
    select {
    case err := <-errCh:
        if err != nil {
            t.Errorf("expected nil error on clean shutdown, got: %v", err)
        }
    case <-time.After(5 * time.Second):
        t.Error("shutdown timed out")
    }
}

func TestServeWithContextCancellation(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    ctx, cancel := context.WithCancel(context.Background())

    errCh := make(chan error, 1)
    go func() {
        errCh <- ServeWithContext(ctx, ":0", handler)
    }()

    // Give server time to start
    time.Sleep(100 * time.Millisecond)

    // Cancel context
    cancel()

    // Should shut down cleanly
    select {
    case err := <-errCh:
        if err != nil {
            t.Errorf("expected nil error on context cancel, got: %v", err)
        }
    case <-time.After(5 * time.Second):
        t.Error("shutdown timed out")
    }
}
```

### Manual Testing

1. Start an example app
2. Send requests
3. Send SIGINT (Ctrl+C)
4. Verify clean shutdown message
5. Verify in-flight requests complete

---

## Acceptance Criteria

- [ ] `Serve` handles SIGINT signal
- [ ] `Serve` handles SIGTERM signal
- [ ] In-flight requests are allowed to complete (up to timeout)
- [ ] Returns `nil` on clean shutdown
- [ ] Returns error if server fails to start
- [ ] `ShutdownTimeout` is configurable via package variable
- [ ] All existing tests pass
- [ ] New tests verify shutdown behavior
- [ ] `make lint` passes
- [ ] `make test` passes

---

## References

- Current implementation: `modkit/http/server.go`
- Documentation claim: `docs/faq.md` (lines 186-197)
- Go http.Server.Shutdown: https://pkg.go.dev/net/http#Server.Shutdown
