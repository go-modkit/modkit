# Error Handling Guide Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Document Go‑idiomatic error handling patterns (handler‑level errors and middleware) as the modkit equivalent of Nest exception filters.

**Architecture:** New guide doc plus link from README.

**Tech Stack:** Markdown docs.

---

### Task 1: Create error handling guide

**Files:**
- Create: `docs/guides/error-handling.md`

**Step 1: Draft the guide**

Include:
- Recommended handler‑level error handling.
- Example using `examples/hello-mysql/internal/httpapi.WriteProblem`.
- Optional middleware pattern for centralized error mapping.

**Step 2: Commit**

```bash
git add docs/guides/error-handling.md
git commit -m "docs: add error handling guide"
```

### Task 2: Link guide from README

**Files:**
- Modify: `README.md`

**Step 1: Add error handling guide to the Guides list**

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: link error handling guide"
```
