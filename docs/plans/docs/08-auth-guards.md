# D8: Auth & Guards Guide

**Status:** 🔴 Not started  
**Type:** New guide  
**NestJS Equivalent:** Guards

---

## Goal

Document Go‑idiomatic auth/authorization patterns (middleware + context) as the modkit equivalent of Nest guards.

## Why Different from NestJS

NestJS guards are framework hooks that run before handlers and return boolean/throw. In Go, auth is implemented as middleware that:
- Validates credentials
- Sets user info in context
- Returns 401/403 or calls next handler

## Files to Create/Modify

- Create: `docs/guides/auth-guards.md`
- Modify: `README.md` (add link)

---

## Task 1: Create auth/guards guide

**Files:**
- Create: `docs/guides/auth-guards.md`

### Step 1: Draft the guide

Include:

1. **Auth middleware example** — validates token, sets user in context
2. **Handler example** — reads user from context
3. **Role-based authorization** — middleware that checks roles
4. **Recommendation** — keep auth in middleware, not core

Suggested structure:

```markdown
# Authentication & Authorization

modkit uses middleware for auth instead of framework-level guards.

## Authentication Middleware

```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        user, err := validateToken(token)
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        ctx := context.WithValue(r.Context(), userKey, user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Reading User in Handlers

```go
func (c *Controller) GetProfile(w http.ResponseWriter, r *http.Request) {
    user := UserFromContext(r.Context())
    if user == nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    // use user
}
```

## Role-Based Authorization

```go
func RequireRole(role string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := UserFromContext(r.Context())
            if user == nil || !user.HasRole(role) {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// Usage
router.Route("/admin", func(r chi.Router) {
    r.Use(AuthMiddleware)
    r.Use(RequireRole("admin"))
    r.Get("/dashboard", adminDashboard)
})
```

## Context Helpers

See the [Context Helpers](context-helpers.md) guide for typed context key patterns.
```

### Step 2: Commit

```bash
git add docs/guides/auth-guards.md
git commit -m "docs: add auth/guards guide"
```

---

## Task 2: Link guide from README

**Files:**
- Modify: `README.md`

### Step 1: Add auth/guards guide to the Guides list

### Step 2: Commit

```bash
git add README.md
git commit -m "docs: link auth/guards guide"
```
