# Providers Docs Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Document provider lifecycle (lazy singleton construction, visibility, cycle errors) in the guides.

**Architecture:** Documentation-only changes in design and guide docs.

**Tech Stack:** Markdown docs.

---

### Task 1: Add provider lifecycle note to design doc

**Files:**
- Modify: `docs/design/mvp.md`

**Step 1: Add a short subsection under provider semantics**

Suggested content:
```markdown
**Provider lifecycle**
Providers are singletons built lazily on first `Get`. Cycles result in `ProviderCycleError`. Build errors surface as `ProviderBuildError` with module/token context.
```

**Step 2: Commit**

```bash
git add docs/design/mvp.md
git commit -m "docs: describe provider lifecycle and errors"
```

### Task 2: Add provider lifecycle note to modules guide

**Files:**
- Modify: `docs/guides/modules.md`

**Step 1: Add a short note near Providers**

Suggested content:
```markdown
Providers are lazy singletons; they are constructed on first `Get` and cached. Cycles are detected and returned as errors.
```

**Step 2: Commit**

```bash
git add docs/guides/modules.md
git commit -m "docs: add provider lifecycle note"
```
