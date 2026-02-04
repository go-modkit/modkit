# Definition Purity Documentation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Document that `Module.Definition()` must be deterministic and side-effect free, since it may be called multiple times during graph construction.

**Architecture:** Add explicit guidance in design and modules guide; no code changes required.

**Tech Stack:** Markdown docs.

---

### Task 1: Update design doc

**Files:**
- Modify: `docs/design/mvp.md`

**Step 1: Add a short section under module semantics**

Add a subsection like:
```markdown
**Definition() must be deterministic**
`Definition()` can be called more than once during graph construction. It must be side-effect free and return consistent metadata for the lifetime of the module instance.
```

**Step 2: Commit**

```bash
git add docs/design/mvp.md
git commit -m "docs: clarify module Definition purity"
```

### Task 2: Update modules guide

**Files:**
- Modify: `docs/guides/modules.md`

**Step 1: Add a note near the ModuleDef description**

Add:
```markdown
**Note:** `Definition()` must be deterministic and side-effect free. The kernel may call it multiple times when building the module graph.
```

**Step 2: Commit**

```bash
git add docs/guides/modules.md
git commit -m "docs: note Definition must be deterministic"
```
