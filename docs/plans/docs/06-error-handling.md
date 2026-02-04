# D6: Error Handling Guide

**Status:** 🔴 Not started  
**Type:** New guide  
**NestJS Equivalent:** Exception Filters

---

## Goal

Document Go‑idiomatic error handling patterns (handler‑level errors and middleware) as the modkit equivalent of Nest exception filters.

## Why Different from NestJS

NestJS exception filters catch thrown exceptions and transform them. In Go, errors are returned values, not exceptions. The idiomatic approach is:
- Return errors from handlers
- Handle errors at the call site or via middleware
- Use structured error responses (RFC 7807 Problem Details)

## Files to Create/Modify

- Create: `docs/guides/error-handling.md`
- Modify: `README.md` (add link)

---

## Task 1: Create error handling guide

**Files:**
- Create: `docs/guides/error-handling.md`

### Step 1: Draft the guide

Include:

1. **Handler-level error handling** — explicit `if err != nil` patterns
2. **Example using `httpapi.WriteProblem`** from hello-mysql
3. **Error middleware pattern** — centralized error mapping
4. **Structured errors** — RFC 7807 Problem Details

Suggested structure:

```markdown
# Error Handling

Go handles errors as returned values, not exceptions. modkit follows this pattern.

## Handler-Level Errors

```go
func (c *Controller) GetUser(w http.ResponseWriter, r *http.Request) {
    user, err := c.service.FindByID(r.Context(), id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            httpapi.WriteProblem(w, http.StatusNotFound, "User not found")
            return
        }
        httpapi.WriteProblem(w, http.StatusInternalServerError, "Internal error")
        return
    }
    json.NewEncoder(w).Encode(user)
}
```

## Structured Error Responses

Use RFC 7807 Problem Details for consistent error responses:

```go
type Problem struct {
    Type   string `json:"type,omitempty"`
    Title  string `json:"title"`
    Status int    `json:"status"`
    Detail string `json:"detail,omitempty"`
}
```

## Error Middleware (Optional)

For centralized error handling, wrap handlers:

```go
func ErrorHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                httpapi.WriteProblem(w, 500, "Internal server error")
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```
```

### Step 2: Commit

```bash
git add docs/guides/error-handling.md
git commit -m "docs: add error handling guide"
```

---

## Task 2: Link guide from README

**Files:**
- Modify: `README.md`

### Step 1: Add error handling guide to the Guides list

### Step 2: Commit

```bash
git add README.md
git commit -m "docs: link error handling guide"
```
