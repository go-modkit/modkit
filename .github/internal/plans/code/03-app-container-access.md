# C3: Standardize Container Access Pattern

**Status:** 🔴 Not started  
**Type:** Code/Docs alignment  
**Priority:** Medium

---

## Motivation

Documentation shows users accessing providers via `app.Container.Get("token")`, but `container` is an unexported field in the `App` struct. This causes confusion when users follow the guides.

The `App` struct already has `App.Get(token)` and `App.Resolver()` methods that provide the same functionality, but docs don't use them consistently.

**Options:**
1. **Export Container** — Change `container` to `Container` 
2. **Fix docs** — Update all docs to use `App.Get(token)` instead

This plan recommends **Option 2** (fix docs) because:
- `App.Get()` is cleaner API — single method vs nested access
- Keeps internal container implementation private
- No breaking changes if container internals change later

---

## Assumptions

1. `App.Get(token)` provides equivalent functionality to direct container access
2. All doc references to `app.Container.Get()` can be replaced with `app.Get()`
3. No external code depends on `Container` being exported (it never was)

---

## Requirements

### R1: Audit all documentation references

Find and update all instances of `app.Container.Get()` pattern.

### R2: Ensure App.Get() is documented

The `App.Get(token)` method should be documented in the API reference.

---

## Files to Modify

| File | Change |
|------|--------|
| `docs/guides/providers.md` | Update container access pattern |
| `docs/guides/middleware.md` | Update container access pattern |
| `docs/guides/comparison.md` | Update container access pattern |
| `docs/guides/authentication.md` | Update container access pattern |
| `docs/reference/api.md` | Ensure `App.Get()` is documented, remove `Container` field if shown |

---

## Implementation

### Step 1: Update docs/guides/providers.md

Change (around line 202):

```go
// Cleanup on shutdown
if db, err := app.Container.Get("db.connection"); err == nil {
    db.(*sql.DB).Close()
}
```

To:

```go
// Cleanup on shutdown
if db, err := app.Get("db.connection"); err == nil {
    db.(*sql.DB).Close()
}
```

### Step 2: Update docs/guides/middleware.md

Change (around line 289):

```go
authMW, _ := app.Container.Get("middleware.auth")
```

To:

```go
authMW, _ := app.Get("middleware.auth")
```

### Step 3: Update docs/guides/comparison.md

Change (around line 130):

```go
svc, _ := app.Container.Get("users.service")
```

To:

```go
svc, _ := app.Get("users.service")
```

### Step 4: Update docs/guides/authentication.md

Change (around line 225):

```go
authMW, _ := app.Container.Get("auth.middleware")
```

To:

```go
authMW, _ := app.Get("auth.middleware")
```

### Step 5: Update docs/reference/api.md

Ensure `App` struct documentation shows:

```go
type App struct {
    Controllers map[string]any
    Graph       *Graph
}

// Get resolves a token from the root module scope.
func (a *App) Get(token Token) (any, error)

// Resolver returns a root-scoped resolver that enforces module visibility.
func (a *App) Resolver() Resolver
```

Remove any reference to `Container` field if present.

---

## Validation

### Grep Verification

After changes, this should return no results:

```bash
grep -r "app\.Container" docs/
```

### Build Check

Ensure example apps compile (they may use the correct pattern already):

```bash
go build ./examples/...
```

---

## Acceptance Criteria

- [ ] No documentation references `app.Container.Get()`
- [ ] All container access uses `app.Get(token)` pattern
- [ ] `App.Get()` method is documented in API reference
- [ ] Example apps compile and work correctly
- [ ] `grep -r "app\.Container" docs/` returns empty

---

## References

- Current App implementation: `modkit/kernel/bootstrap.go` (lines 5-9, 49-57)
- Affected docs: `docs/guides/providers.md`, `docs/guides/middleware.md`, `docs/guides/comparison.md`, `docs/guides/authentication.md`
