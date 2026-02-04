# Middleware Guide Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a Go‑idiomatic middleware guide for modkit using `http.Handler` and `chi`.

**Architecture:** New guide doc plus link from README.

**Tech Stack:** Markdown docs.

---

### Task 1: Create middleware guide

**Files:**
- Create: `docs/guides/middleware.md`

**Step 1: Draft the guide**

Include:
- What middleware is in Go (`func(http.Handler) http.Handler`).
- Example using `modkit/http.NewRouter()` and `RequestLogger`.
- Guidance for ordering (recover → auth → logging).

**Step 2: Commit**

```bash
git add docs/guides/middleware.md
git commit -m "docs: add middleware guide"
```

### Task 2: Link guide from README

**Files:**
- Modify: `README.md`

**Step 1: Add middleware guide to the Guides list**

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: link middleware guide"
```
