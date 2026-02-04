# D10: Context Helpers Guide

**Status:** 🔴 Not started  
**Type:** New guide  
**NestJS Equivalent:** Custom Decorators

---

## Goal

Document Go‑idiomatic context helper patterns as the modkit equivalent of Nest custom decorators.

## Why Different from NestJS

NestJS custom decorators like `@User()` or `@Roles()` use metadata reflection to extract request data. In Go, this is achieved with typed context keys and helper functions.

## Files to Create/Modify

- Create: `docs/guides/context-helpers.md`
- Modify: `README.md` (add link)

---

## Task 1: Create context helpers guide

**Files:**
- Create: `docs/guides/context-helpers.md`

### Step 1: Draft the guide

Include:

1. **Typed context keys** — unexported key types
2. **Helper functions** — `WithUser(ctx, user)` and `UserFromContext(ctx)`
3. **Best practices** — keep keys unexported, return zero values safely
4. **Multiple values** — request ID, tenant, etc.

Suggested structure:

```markdown
# Context Helpers

Go uses `context.Context` to pass request-scoped values. This guide shows how to create type-safe context helpers.

## Typed Context Keys

Always use unexported types for context keys to avoid collisions:

```go
// internal/auth/context.go
package auth

type contextKey string

const userKey contextKey = "user"
```

## Helper Functions

Provide exported functions to set and get values:

```go
func WithUser(ctx context.Context, user *User) context.Context {
    return context.WithValue(ctx, userKey, user)
}

func UserFromContext(ctx context.Context) *User {
    user, _ := ctx.Value(userKey).(*User)
    return user
}
```

## Using in Middleware

```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user, err := authenticate(r)
        if err != nil {
            http.Error(w, "Unauthorized", 401)
            return
        }
        ctx := auth.WithUser(r.Context(), user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Using in Handlers

```go
func (c *Controller) GetProfile(w http.ResponseWriter, r *http.Request) {
    user := auth.UserFromContext(r.Context())
    if user == nil {
        http.Error(w, "Unauthorized", 401)
        return
    }
    json.NewEncoder(w).Encode(user)
}
```

## Common Context Values

| Value | Package | Setter | Getter |
|-------|---------|--------|--------|
| User | `auth` | `WithUser` | `UserFromContext` |
| Request ID | `requestid` | `WithRequestID` | `RequestIDFromContext` |
| Tenant | `tenant` | `WithTenant` | `TenantFromContext` |

## Best Practices

1. **Keep keys unexported** — prevents external packages from accessing directly
2. **Return nil/zero safely** — check for nil in getters
3. **Don't store large objects** — context is copied frequently
4. **Use for request-scoped data only** — not for dependency injection
```

### Step 2: Commit

```bash
git add docs/guides/context-helpers.md
git commit -m "docs: add context helpers guide"
```

---

## Task 2: Link guide from README

**Files:**
- Modify: `README.md`

### Step 1: Add context helpers guide to the Guides list

### Step 2: Commit

```bash
git add README.md
git commit -m "docs: link context helpers guide"
```
