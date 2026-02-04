# Phase 03 — HTTP Adapter

## Assumptions (Initial State)
- Phase 02 is complete and committed.
- `modkit/kernel` and `modkit/module` are usable for controller resolution.

## Requirements
- Implement `modkit/http` minimal router + server helpers per `modkit_mvp_design_doc.md` Section 5.3.
- Provide baseline middleware helpers (if included in MVP design) and docs on route registration pattern.

## Design
- Source of truth: `modkit_mvp_design_doc.md` Section 5.3 and Section 3.3.
- No reflection; routing is explicit via module route registration functions.
- `NewRouter()` returns a `chi.Router` with baseline middleware.
- Use `github.com/go-chi/chi/v5` as the router dependency.

## Validation
Run:
- `go test ./modkit/http/...`
- `go test ./...`

Expected:
- All tests pass.

## Commit
- One commit after validation, e.g. `feat: add http adapter`.
