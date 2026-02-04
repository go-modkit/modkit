# Repository Guidelines

## Project Structure & Module Organization
- Root design source: `modkit_mvp_design_doc.md` (MVP scope and architecture).
- Implementation phase docs: `docs/implementation/` (master index and per‑phase instructions).
- Future code will live under `modkit/` for library packages and `examples/` for consumer apps (see `modkit_mvp_design_doc.md`).

## Build, Test, and Development Commands
- `go test ./...` runs the full test suite (library + examples once implemented).
- `go test ./modkit/...` runs library package tests only.
- `go test ./examples/...` runs example app tests.

## Coding Style & Naming Conventions
- Go formatting is enforced with `gofmt` (tabs for indentation).
- Package paths use lowercase, short names (e.g., `modkit/module`, `modkit/kernel`).
- Exported types/functions use PascalCase; unexported use camelCase.

## Testing Guidelines
- Use Go’s standard testing package (`testing`).
- Name tests `TestXxx` and keep them near the package under test.
- Prefer table‑driven tests for validators and graph cases.
- Run focused tests with `go test ./modkit/kernel -run TestName`.

## Commit & Pull Request Guidelines
- Commit messages follow conventional prefixes seen in history (e.g., `chore: ...`, `docs: ...`, `feat: ...`).
- One logical change per commit; phase docs require a single commit per phase after validation.
- PRs should include a concise summary, validation commands run, and link to the relevant phase doc (if applicable).

## Architecture Overview
- `module` defines metadata (imports/providers/controllers/exports).
- `kernel` builds the module graph, enforces visibility, and bootstraps the app.
- `http` adapts controllers to routing without reflection.

## Agent-Specific Instructions
- Follow `docs/implementation/master.md` and the phase docs in order.
- Avoid duplicating design text; reference `modkit_mvp_design_doc.md` instead.
