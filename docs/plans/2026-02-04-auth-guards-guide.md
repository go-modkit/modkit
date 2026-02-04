# Auth/Guards Guide Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Document Go‑idiomatic auth/authorization patterns (middleware + context) as the modkit equivalent of Nest guards.

**Architecture:** New guide doc plus link from README.

**Tech Stack:** Markdown docs.

---

### Task 1: Create auth/guards guide

**Files:**
- Create: `docs/guides/auth-guards.md`

**Step 1: Draft the guide**

Include:
- Auth middleware example that sets user in context.
- Handler example that reads from context.
- Recommendation to keep auth in middleware, not core.

**Step 2: Commit**

```bash
git add docs/guides/auth-guards.md
git commit -m "docs: add auth/guards guide"
```

### Task 2: Link guide from README

**Files:**
- Modify: `README.md`

**Step 1: Add auth/guards guide to the Guides list**

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: link auth/guards guide"
```
