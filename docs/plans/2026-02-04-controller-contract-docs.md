# Controller Registration Contract Documentation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Explicitly document the controller contract: controllers must implement `http.RouteRegistrar` and will be registered via `http.RegisterRoutes`.

**Architecture:** Add a short contract statement and minimal example in HTTP adapter docs and getting started guide.

**Tech Stack:** Markdown docs.

---

### Task 1: Update HTTP adapter doc

**Files:**
- Modify: `docs/design/http-adapter.md`

**Step 1: Add a contract section**

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

**Step 2: Commit**

```bash
git add docs/design/http-adapter.md
git commit -m "docs: clarify controller registration contract"
```

### Task 2: Update getting started guide

**Files:**
- Modify: `docs/guides/getting-started.md`

**Step 1: Add a minimal contract callout**

Suggested content:
```markdown
Controllers must implement `modkit/http.RouteRegistrar` and will be registered by `RegisterRoutes`.
```

**Step 2: Commit**

```bash
git add docs/guides/getting-started.md
git commit -m "docs: note controller RouteRegistrar contract"
```
