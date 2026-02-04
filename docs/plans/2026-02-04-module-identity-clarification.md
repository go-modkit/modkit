# Module Identity Clarification Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Clarify module identity rules: modules must be pointers, module names must be unique, and shared imports must reuse the same module instance.

**Architecture:** Expand docs to explain why pointer identity is required and how duplicate names are resolved.

**Tech Stack:** Markdown docs.

---

### Task 1: Update design doc

**Files:**
- Modify: `docs/design/mvp.md`

**Step 1: Add a module identity section**

Suggested content:
```markdown
**Module identity and pointers**
Modules must be passed as pointers to ensure stable identity across shared imports. If two different module instances share the same `Name`, bootstrap will fail with a duplicate module name error. Reuse the same module pointer when importing a shared module.
```

**Step 2: Commit**

```bash
git add docs/design/mvp.md
git commit -m "docs: clarify module identity rules"
```

### Task 2: Update modules guide

**Files:**
- Modify: `docs/guides/modules.md`

**Step 1: Add a short note near Imports**

Suggested content:
```markdown
**Module identity:** Always pass module pointers and reuse the same instance when sharing imports. Duplicate module names across different instances are errors.
```

**Step 2: Commit**

```bash
git add docs/guides/modules.md
git commit -m "docs: add module identity note"
```
