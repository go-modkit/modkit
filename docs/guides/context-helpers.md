# Context Helpers

Go doesn't use decorators. The idiomatic pattern is to attach request-scoped data to `context.Context` using typed keys, then provide small helper functions for setting and retrieving values.

This guide shows the pattern modkit recommends for context helpers, which you can use anywhere in your middleware or handlers.

## Typed Context Keys

Use an unexported key type to avoid collisions with other packages.

```go
package auth

type userKey struct{}

var userKeyInstance = userKey{}
```

Keeping the key type and value unexported prevents accidental use from other packages and makes collisions impossible.

## Helper Functions

Wrap `context.WithValue` and `ctx.Value` in helpers so your handlers stay type-safe and readable.

```go
package auth

import "context"

type User struct {
    ID    string
    Email string
    Role  string
}

func WithUser(ctx context.Context, user *User) context.Context {
    return context.WithValue(ctx, userKeyInstance, user)
}

func UserFromContext(ctx context.Context) (*User, bool) {
    user, ok := ctx.Value(userKeyInstance).(*User)
    return user, ok
}
```

### Using in Middleware

```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user, err := authenticate(r)
        if err != nil {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }

        ctx := auth.WithUser(r.Context(), user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Using in Handlers

```go
func (c *UsersController) Me(w http.ResponseWriter, r *http.Request) {
    user, ok := auth.UserFromContext(r.Context())
    if !ok {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    json.NewEncoder(w).Encode(user)
}
```

## Best Practices

- Keep context keys unexported in the package that defines them.
- Prefer helper functions instead of calling `ctx.Value` directly.
- Return `nil`/`false` when the value is missing rather than panicking.
- Treat `context.Context` as request-scoped data only, not as a general dependency container.

Note: some existing modkit docs use exported context keys for brevity. In production code, unexported key types are the recommended practice to avoid collisions.

## Multiple Values

Use a separate key type and helpers for each value you need to store.

```go
package requestid

import "context"

type requestIDKey struct{}

var requestIDKeyInstance = requestIDKey{}

func WithRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKeyInstance, id)
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
    id, ok := ctx.Value(requestIDKeyInstance).(string)
    return id, ok
}
```

```go
package tenant

import "context"

type tenantKey struct{}

var tenantKeyInstance = tenantKey{}

func WithTenant(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, tenantKeyInstance, id)
}

func TenantFromContext(ctx context.Context) (string, bool) {
    id, ok := ctx.Value(tenantKeyInstance).(string)
    return id, ok
}
```

This keeps each value isolated and avoids type assertions scattered throughout your handlers.
