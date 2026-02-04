# Phase 00 — Repo Bootstrap

## Assumptions (Initial State)
- GitHub repo exists: `aryeko/modkit`.
- Local repo cloned and `origin` is configured for push access (SSH or HTTPS).
- `modkit_mvp_design_doc.md` exists at repo root and is committed.

## Requirements
- Add a placeholder test package so `go test ./...` runs in an empty repo.
- Initialize Go module at `github.com/aryeko/modkit`.
- Add baseline repository files:
  - `README.md` (MVP summary + quickstart stub)
  - `LICENSE`
  - `CONTRIBUTING.md`
  - `CODE_OF_CONDUCT.md`
  - `SECURITY.md`
- Create `docs/design/mvp.md` as the canonical design doc by moving the contents of `modkit_mvp_design_doc.md`.
- Replace `modkit_mvp_design_doc.md` with a short pointer to `docs/design/mvp.md` (or remove it entirely).
- Add CI workflow: `go test ./...`.

## Design
- Follow `modkit_mvp_design_doc.md` Section 4 (Repo structure) as the target layout.
- This phase establishes scaffolding only; no public API implementations yet.

## Validation
Run:
- `go mod tidy`
- `go test ./...`

Expected:
- `go test` succeeds (no packages or only empty packages with passing tests).

## Commit
- One commit after validation, e.g. `chore: bootstrap repo structure`.
