# D3: Providers

**Status:** 🔴 Not started  
**Type:** Documentation improvement  
**NestJS Equivalent:** Providers

---

## Goal

Document provider lifecycle: lazy singleton construction, visibility enforcement, and cycle/build error handling.

## Files to Modify

- `docs/design/mvp.md`
- `docs/guides/modules.md`

---

## Task 1: Add provider lifecycle note to design doc

**Files:**
- Modify: `docs/design/mvp.md`

### Step 1: Add a short subsection under provider semantics

Suggested content:

```markdown
**Provider lifecycle**

Providers are singletons built lazily on first `Get`. Cycles result in `ProviderCycleError`. Build errors surface as `ProviderBuildError` with module/token context.
```

### Step 2: Commit

```bash
git add docs/design/mvp.md
git commit -m "docs: describe provider lifecycle and errors"
```

---

## Task 2: Add provider lifecycle note to modules guide

**Files:**
- Modify: `docs/guides/modules.md`

### Step 1: Add a short note near Providers section

Suggested content:

```markdown
Providers are lazy singletons; they are constructed on first `Get` and cached. Cycles are detected and returned as errors.
```

### Step 2: Commit

```bash
git add docs/guides/modules.md
git commit -m "docs: add provider lifecycle note"
```
