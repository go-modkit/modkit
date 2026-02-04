# Context Helpers Guide Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Document Go‑idiomatic context helper patterns as the modkit equivalent of Nest custom decorators.

**Architecture:** New guide doc plus link from README.

**Tech Stack:** Markdown docs.

---

### Task 1: Create context helpers guide

**Files:**
- Create: `docs/guides/context-helpers.md`

**Step 1: Draft the guide**

Include:
- Typed context keys and helper functions.
- Example: `type contextKey string` and `WithUser(ctx, user)`.
- Guidance to keep context keys unexported.

**Step 2: Commit**

```bash
git add docs/guides/context-helpers.md
git commit -m "docs: add context helpers guide"
```

### Task 2: Link guide from README

**Files:**
- Modify: `README.md`

**Step 1: Add context helpers guide to the Guides list**

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: link context helpers guide"
```
