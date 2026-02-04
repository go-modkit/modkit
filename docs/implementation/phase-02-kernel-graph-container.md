# Phase 02 — Kernel Graph + Container

## Assumptions (Initial State)
- Phase 01 is complete and committed.
- `modkit/module` public API is stable per MVP design.

## Requirements
- Implement public package `modkit/kernel` per `modkit_mvp_design_doc.md` Section 5.2.
- Implement graph flattening, cycle detection, duplicate module name checks, duplicate provider token checks.
- Implement visibility enforcement via module-scoped resolver.
- Implement `kernel.Bootstrap` and `kernel.App`.
- Add full kernel unit test suite as per `modkit_mvp_design_doc.md` Section 8.1.

## Design
- Source of truth: `modkit_mvp_design_doc.md` Sections 3.2, 3.3, 5.2, 8.1.
- Graph order must be deterministic, imports-first.
- Errors include module/provider token context.

## Validation
Run:
- `go test ./modkit/kernel/...`
- `go test ./...`

Expected:
- All kernel tests pass.

## Commit
- One commit after validation, e.g. `feat: add kernel bootstrap and graph`.
