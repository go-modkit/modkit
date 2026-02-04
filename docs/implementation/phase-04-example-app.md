# Phase 04 — Example App (hello-mysql)

## Assumptions (Initial State)
- Phase 03 is complete and committed.
- `modkit/module`, `modkit/kernel`, and `modkit/http` are stable and usable.
- MySQL is available locally via Docker (or testcontainers) for integration tests.

## Requirements
- Implement example app in `examples/hello-mysql` as described in `modkit_mvp_design_doc.md` Section 4.2 and Section 8.2.
- Provide modules: AppModule, DatabaseModule, UsersModule, plus one additional module that consumes an exported service.
- Include SQL schema, sqlc config, and migrations.
- Provide a working endpoint group (GET/POST) plus a `/health` endpoint.
- Add Makefile targets: `make run`, `make test`.
- Default HTTP listen address for the example app is `:8080` (used by smoke tests).

## Design
- Source of truth: `modkit_mvp_design_doc.md` Sections 4.2, 6, 8.2, 10.4.
- Keep module boundaries clean and follow export/import visibility rules.
- Demonstrate re-export or cross-module service usage as specified.

## Validation
Run:
- `go test ./examples/hello-mysql/...`
- `make test` (from `examples/hello-mysql`)
- Manual smoke (required):
  - `make run` (server listens on `http://localhost:8080`)
  - `curl http://localhost:8080/health` → `200` and body `{\"status\":\"ok\"}` (or equivalent JSON with status ok)
  - `curl` one CRUD endpoint (GET or POST) → `200` or `201` with non-empty JSON body

Expected:
- All tests and smoke checks pass.

## Commit
- One commit after validation, e.g. `feat: add hello-mysql example app`.
