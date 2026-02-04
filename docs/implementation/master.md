# modkit MVP Implementation — Master Index

This is the master index for the modkit MVP implementation. Each phase is a separately validated and committed unit of work, with its own assumptions, requirements, design notes, and validation steps. Do not start a phase unless its assumptions are true.

## Canonical Design Source
- `docs/design/mvp.md` is the source of truth for MVP scope, architecture, and validation criteria.
- If `modkit_mvp_design_doc.md` remains at repo root, it should be a short pointer to `docs/design/mvp.md` to avoid duplication.

## Phase Structure (applies to every phase)
Each phase doc contains:
- **Assumptions (Initial State)**: Preconditions that must be true before starting.
- **Requirements**: Deliverables for the phase.
- **Design**: Implementation intent and structure, referencing shared docs.
- **Validation**: Exact commands and expected outcomes.
- **Commit**: One commit per phase after validation.

## Phases (in order)
1. Phase 00 — Repo Bootstrap
   - Doc: `docs/implementation/phase-00-repo-bootstrap.md`
2. Phase 01 — module package
   - Doc: `docs/implementation/phase-01-module-package.md`
3. Phase 02 — kernel graph + container
   - Doc: `docs/implementation/phase-02-kernel-graph-container.md`
4. Phase 03 — HTTP adapter
   - Doc: `docs/implementation/phase-03-http-adapter.md`
5. Phase 04 — Example app (hello-mysql)
   - Doc: `docs/implementation/phase-04-example-app.md`
6. Phase 05 — Docs + CI completeness
   - Doc: `docs/implementation/phase-05-docs-ci.md`

## Cross-Phase Notes
- **Branching/Isolation:** Use a worktree for each implementation run if needed.
- **Validation Discipline:** Never claim completion without running the phase’s validation commands.
- **No Duplication:** Shared architecture and semantics are defined in `modkit_mvp_design_doc.md`. Phase docs refer to it instead of restating.
