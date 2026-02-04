# Validation Guide Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Document Go‑idiomatic validation and transformation patterns as the modkit equivalent of Nest pipes.

**Architecture:** New guide doc plus link from README.

**Tech Stack:** Markdown docs.

---

### Task 1: Create validation guide

**Files:**
- Create: `docs/guides/validation.md`

**Step 1: Draft the guide**

Include:
- JSON decode + validation flow.
- Example with `json.Decoder` + `DisallowUnknownFields`.
- Optional mention of a validator library (no dependency in core).

**Step 2: Commit**

```bash
git add docs/guides/validation.md
git commit -m "docs: add validation guide"
```

### Task 2: Link guide from README

**Files:**
- Modify: `README.md`

**Step 1: Add validation guide to the Guides list**

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: link validation guide"
```
