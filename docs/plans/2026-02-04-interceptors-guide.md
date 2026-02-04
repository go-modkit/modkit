# Interceptors Guide Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Document Go‑idiomatic request/response interception using middleware and handler wrappers.

**Architecture:** New guide doc plus link from README.

**Tech Stack:** Markdown docs.

---

### Task 1: Create interceptors guide

**Files:**
- Create: `docs/guides/interceptors.md`

**Step 1: Draft the guide**

Include:
- Explanation that middleware/wrappers are the Go equivalent.
- Example: timing/logging wrapper and response status capture.

**Step 2: Commit**

```bash
git add docs/guides/interceptors.md
git commit -m "docs: add interceptors guide"
```

### Task 2: Link guide from README

**Files:**
- Modify: `README.md`

**Step 1: Add interceptors guide to the Guides list**

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: link interceptors guide"
```
