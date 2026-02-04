# modkit MVP Implementation Docs

This folder contains the master index and phase-specific implementation docs for the modkit MVP. Each phase doc is standalone for execution, but references shared context to avoid duplication.

## Master Index
- `docs/implementation/master.md`

## Phases
- `docs/implementation/phase-00-repo-bootstrap.md`
- `docs/implementation/phase-01-module-package.md`
- `docs/implementation/phase-02-kernel-graph-container.md`
- `docs/implementation/phase-03-http-adapter.md`
- `docs/implementation/phase-04-example-app.md`
- `docs/implementation/phase-05-docs-ci.md`

## Shared Context
The canonical MVP design is in `docs/design/mvp.md`. If `modkit_mvp_design_doc.md` exists at repo root, it should be a short pointer to the canonical doc. Phase docs reference the canonical design and/or prior phases for shared architecture.
