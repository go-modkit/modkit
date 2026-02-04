# D5: Middleware Guide

**Status:** 🔴 Not started  
**Type:** New guide  
**NestJS Equivalent:** Middleware

---

## Goal

Add a Go‑idiomatic middleware guide for modkit using `http.Handler` and `chi`.

## Files to Create/Modify

- Create: `docs/guides/middleware.md`
- Modify: `README.md` (add link)

---

## Task 1: Create middleware guide

**Files:**
- Create: `docs/guides/middleware.md`

### Step 1: Draft the guide

Include:

1. **What middleware is in Go** — `func(http.Handler) http.Handler`
2. **Example using modkit/http.NewRouter() and RequestLogger**
3. **Ordering guidance** — recover → auth → logging
4. **Chi-specific patterns** — `router.Use()`, route groups

Suggested structure:

```markdown
# Middleware

Middleware in Go wraps HTTP handlers to add cross-cutting behavior like logging, authentication, and error recovery.

## The Middleware Pattern

```go
func MyMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // before
        next.ServeHTTP(w, r)
        // after
    })
}
```

## Using Middleware with modkit

```go
router := mkhttp.NewRouter()
router.Use(mkhttp.RequestLogger(logger))
router.Use(RecoverMiddleware)
```

## Ordering

Apply middleware in this order:
1. Recovery (outermost)
2. Request ID
3. Logging
4. Authentication
5. Route handlers (innermost)

## Route-Specific Middleware

Use chi route groups for middleware that applies to specific routes:

```go
router.Route("/admin", func(r chi.Router) {
    r.Use(RequireAdmin)
    r.Get("/", adminHandler)
})
```
```

### Step 2: Commit

```bash
git add docs/guides/middleware.md
git commit -m "docs: add middleware guide"
```

---

## Task 2: Link guide from README

**Files:**
- Modify: `README.md`

### Step 1: Add middleware guide to the Guides list

Add `docs/guides/middleware.md` to the Guides section.

### Step 2: Commit

```bash
git add README.md
git commit -m "docs: link middleware guide"
```
