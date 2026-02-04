# D2: Modules

**Status:** 🔴 Not started  
**Type:** Documentation improvement  
**NestJS Equivalent:** Modules

---

## Goal

Clarify module semantics:
1. Modules must be pointers to ensure stable identity across shared imports
2. `Definition()` must be deterministic and side-effect free

## Files to Modify

- `docs/design/mvp.md`
- `docs/guides/modules.md`

---

## Task 1: Add module identity section to design doc

**Files:**
- Modify: `docs/design/mvp.md`

### Step 1: Add a module identity section

Suggested content:

```markdown
**Module identity and pointers**

Modules must be passed as pointers to ensure stable identity across shared imports. If two different module instances share the same `Name`, bootstrap will fail with a duplicate module name error. Reuse the same module pointer when importing a shared module.
```

### Step 2: Add Definition() purity section

Suggested content:

```markdown
**Definition() must be deterministic**

`Definition()` can be called more than once during graph construction. It must be side-effect free and return consistent metadata for the lifetime of the module instance.
```

### Step 3: Commit

```bash
git add docs/design/mvp.md
git commit -m "docs: clarify module identity and Definition purity"
```

---

## Task 2: Update modules guide

**Files:**
- Modify: `docs/guides/modules.md`

### Step 1: Add a note near Imports

Suggested content:

```markdown
**Module identity:** Always pass module pointers and reuse the same instance when sharing imports. Duplicate module names across different instances are errors.
```

### Step 2: Add a note near ModuleDef description

Suggested content:

```markdown
**Note:** `Definition()` must be deterministic and side-effect free. The kernel may call it multiple times when building the module graph.
```

### Step 3: Commit

```bash
git add docs/guides/modules.md
git commit -m "docs: add module identity and Definition purity notes"
```
