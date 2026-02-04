# C2: Extend Router Interface with Group and Use

**Status:** 🔴 Not started  
**Type:** Code change  
**Priority:** High

---

## Motivation

The current `Router` interface in `modkit/http/router.go` only exposes `Handle(method, pattern, handler)`. However, the documentation describes a richer interface including `Group()` for route grouping and `Use()` for middleware attachment.

Multiple guides depend on this API:
- `docs/guides/controllers.md` shows `r.Group("/users", ...)` patterns
- `docs/guides/middleware.md` shows `r.Use(authMiddleware)` patterns
- `docs/reference/api.md` documents the full interface

Without this, users following the guides will encounter compilation errors.

---

## Assumptions

1. The underlying chi router already supports `Group` and `Use` — we're exposing existing functionality
2. The `routerAdapter` wrapper can be extended to delegate to chi
3. No breaking changes to existing code — we're adding methods, not changing existing ones

---

## Requirements

### R1: Add Group method to Router interface

```go
Group(pattern string, fn func(Router))
```

- Creates a sub-router scoped to the pattern prefix
- The callback receives a `Router` that can register routes relative to the group
- Middleware applied via `Use` inside the group only affects routes in that group

### R2: Add Use method to Router interface

```go
Use(middlewares ...func(http.Handler) http.Handler)
```

- Attaches middleware to the router (or group)
- Middleware executes in order of registration
- Must work at both global level and within groups

### R3: Update routerAdapter to implement new methods

The `routerAdapter` struct must delegate to the underlying `chi.Router` methods.

---

## Files to Modify

| File | Change |
|------|--------|
| `modkit/http/router.go` | Add `Group` and `Use` to interface, implement in adapter |
| `modkit/http/router_test.go` | Add tests for new functionality |

---

## Implementation

### Step 1: Update Router interface

In `modkit/http/router.go`, change:

```go
type Router interface {
    Handle(method string, pattern string, handler http.Handler)
}
```

To:

```go
type Router interface {
    Handle(method string, pattern string, handler http.Handler)
    Group(pattern string, fn func(Router))
    Use(middlewares ...func(http.Handler) http.Handler)
}
```

### Step 2: Implement Group in routerAdapter

```go
func (r routerAdapter) Group(pattern string, fn func(Router)) {
    r.Router.Route(pattern, func(sub chi.Router) {
        fn(routerAdapter{Router: sub})
    })
}
```

### Step 3: Implement Use in routerAdapter

```go
func (r routerAdapter) Use(middlewares ...func(http.Handler) http.Handler) {
    r.Router.Use(middlewares...)
}
```

---

## Validation

### Unit Tests

Add to `modkit/http/router_test.go`:

```go
func TestRouterGroup(t *testing.T) {
    router := chi.NewRouter()
    r := AsRouter(router)
    
    called := false
    r.Group("/api", func(sub Router) {
        sub.Handle(http.MethodGet, "/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            called = true
        }))
    })
    
    req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)
    
    if !called {
        t.Error("handler not called for grouped route")
    }
}

func TestRouterUse(t *testing.T) {
    router := chi.NewRouter()
    r := AsRouter(router)
    
    middlewareCalled := false
    r.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
            middlewareCalled = true
            next.ServeHTTP(w, req)
        })
    })
    
    r.Handle(http.MethodGet, "/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
    
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)
    
    if !middlewareCalled {
        t.Error("middleware not called")
    }
}

func TestRouterGroupWithMiddleware(t *testing.T) {
    router := chi.NewRouter()
    r := AsRouter(router)
    
    groupMiddlewareCalled := false
    globalHandlerCalled := false
    groupHandlerCalled := false
    
    r.Handle(http.MethodGet, "/public", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        globalHandlerCalled = true
    }))
    
    r.Group("/protected", func(sub Router) {
        sub.Use(func(next http.Handler) http.Handler {
            return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
                groupMiddlewareCalled = true
                next.ServeHTTP(w, req)
            })
        })
        sub.Handle(http.MethodGet, "/resource", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            groupHandlerCalled = true
        }))
    })
    
    // Call public route - middleware should NOT be called
    req1 := httptest.NewRequest(http.MethodGet, "/public", nil)
    router.ServeHTTP(httptest.NewRecorder(), req1)
    
    if groupMiddlewareCalled {
        t.Error("group middleware should not affect routes outside group")
    }
    
    // Call protected route - middleware SHOULD be called
    req2 := httptest.NewRequest(http.MethodGet, "/protected/resource", nil)
    router.ServeHTTP(httptest.NewRecorder(), req2)
    
    if !groupMiddlewareCalled || !groupHandlerCalled {
        t.Error("group middleware or handler not called")
    }
}
```

### Integration Test

Verify the example app still works:

```bash
go test ./modkit/http/...
go test ./examples/hello-mysql/...
```

---

## Acceptance Criteria

- [ ] `Router` interface includes `Group(pattern string, fn func(Router))`
- [ ] `Router` interface includes `Use(middlewares ...func(http.Handler) http.Handler)`
- [ ] `routerAdapter` implements both methods by delegating to chi
- [ ] Routes registered inside `Group` are prefixed correctly
- [ ] Middleware applied via `Use` inside a group only affects that group
- [ ] All existing tests pass
- [ ] New unit tests cover `Group`, `Use`, and combination scenarios
- [ ] `make lint` passes
- [ ] `make test` passes

---

## References

- Current implementation: `modkit/http/router.go`
- Documented interface: `docs/reference/api.md` (lines 174-179)
- Guide usage: `docs/guides/controllers.md` (lines 126-134), `docs/guides/middleware.md` (lines 44-57)
- chi documentation: https://github.com/go-chi/chi
