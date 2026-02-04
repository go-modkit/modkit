# D4: Controllers

**Status:** 🔴 Not started  
**Type:** Documentation improvement  
**NestJS Equivalent:** Controllers

---

## Goal

Explicitly document the controller contract: controllers must implement `http.RouteRegistrar` and will be registered via `http.RegisterRoutes`.

## Files to Modify

- `docs/design/http-adapter.md`
- `docs/guides/getting-started.md`

---

## Task 1: Update HTTP adapter doc

**Files:**
- Modify: `docs/design/http-adapter.md`

### Step 1: Add a contract section

Suggested content:

```markdown
**Controller contract**

Controllers must implement:

```go
type RouteRegistrar interface {
    RegisterRoutes(router Router)
}
```

The HTTP adapter will type-assert each controller to `RouteRegistrar` and return an error if any controller does not implement it.
```

### Step 2: Commit

```bash
git add docs/design/http-adapter.md
git commit -m "docs: clarify controller registration contract"
```

---

## Task 2: Update getting started guide

**Files:**
- Modify: `docs/guides/getting-started.md`

### Step 1: Add a minimal contract callout

Suggested content:

```markdown
Controllers must implement `modkit/http.RouteRegistrar` and will be registered by `RegisterRoutes`.
```

### Step 2: Commit

```bash
git add docs/guides/getting-started.md
git commit -m "docs: note controller RouteRegistrar contract"
```
