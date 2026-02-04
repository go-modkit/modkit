# Phase 05 — Docs + CI Completeness

## Assumptions (Initial State)
- Phase 04 is complete and committed.
- Example app and core packages compile and pass tests.

## Requirements
- Complete docs per `modkit_mvp_design_doc.md` Section 10.2:
  - README: what modkit is, quickstart, architecture overview, NestJS inspiration note.
  - Guides: `docs/guides/getting-started.md`, `docs/guides/modules.md`, `docs/guides/testing.md`.
- Ensure `docs/design/mvp.md` exists (from Phase 00).
- CI workflow runs `go test ./...` for library and example (or same command if monorepo).
- Optional stretch (not required for phase completion): add `golangci-lint` workflow step.

## Design
- Source of truth: `modkit_mvp_design_doc.md` Sections 4.1, 10.2, 10.3.
- Avoid duplicating architecture text; link to `docs/design/mvp.md` where possible.

## Validation
Run:
- `go test ./...`
- If lint was added: `golangci-lint run`.

Expected:
- All checks pass.

## Commit
- One commit after validation, e.g. `docs: complete guides and ci`.
