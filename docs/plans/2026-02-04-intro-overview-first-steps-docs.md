# Intro/Overview/First Steps Docs Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Clarify modkit’s purpose and onboarding by adding a “Why modkit / no reflection” callout, a simple architecture flow, and a minimal bootstrap snippet.

**Architecture:** Documentation-only changes in `README.md` and `docs/guides/getting-started.md`.

**Tech Stack:** Markdown docs.

---

### Task 1: Add “Why modkit” and architecture flow to README

**Files:**
- Modify: `README.md`

**Step 1: Add a short “Why modkit” section**

Suggested content:
```markdown
## Why modkit?

modkit is a Go‑idiomatic alternative to decorator‑driven frameworks. It keeps wiring explicit, avoids reflection, and makes module boundaries and dependencies visible in code.
```

**Step 2: Add an architecture flow callout**

Suggested content:
```markdown
## Architecture Flow

Module definitions → kernel graph/visibility → provider container → controller instances → HTTP adapter
```

**Step 3: Commit**

```bash
git add README.md
git commit -m "docs: clarify modkit purpose and architecture flow"
```

### Task 2: Add a minimal bootstrap snippet to getting started

**Files:**
- Modify: `docs/guides/getting-started.md`

**Step 1: Add a short “Minimal main.go” snippet near the top**

Suggested content:
```go
func main() {
    appInstance, err := kernel.Bootstrap(&app.AppModule{})
    if err != nil {
        log.Fatal(err)
    }

    router := mkhttp.NewRouter()
    _ = mkhttp.RegisterRoutes(mkhttp.AsRouter(router), appInstance.Controllers)
    _ = mkhttp.Serve(":8080", router)
}
```

**Step 2: Commit**

```bash
git add docs/guides/getting-started.md
git commit -m "docs: add minimal bootstrap snippet"
```
