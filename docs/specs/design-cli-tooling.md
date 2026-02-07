# Design Spec: modkit CLI Remaining Work

**Status:** Draft
**Date:** 2026-02-07
**Author:** Sisyphus (AI Agent)
**Related PRD:** `docs/specs/prd-modkit-core.md` (Phase 2 `modkit-cli`)
**Related Spec:** `docs/specs/design-release-versioning-sdlc-cli.md`

## 1. Overview

This document defines the remaining work required to complete the PRD `modkit-cli` item.

The repository already contains a working CLI skeleton and scaffolding commands. The remaining work is to make scaffolding fully usable without manual follow-up edits, make behavior idempotent and safe, and close SDLC/documentation gaps.

## 2. Current State (Implemented)

The following are already implemented and tested:

- CLI entrypoint and command tree:
  - `cmd/modkit/main.go`
  - `internal/cli/cmd/root.go`
  - `internal/cli/cmd/new.go`
- Generation commands:
  - `modkit new app`: `internal/cli/cmd/new_app.go`
  - `modkit new module`: `internal/cli/cmd/new_module.go`
  - `modkit new provider`: `internal/cli/cmd/new_provider.go`
  - `modkit new controller`: `internal/cli/cmd/new_controller.go`
- Embedded template infrastructure:
  - `internal/cli/templates/embed.go`
- AST helper for provider registration (currently unused by command flow):
  - `internal/cli/ast/modify.go`
- Unit tests for command behaviors:
  - `internal/cli/cmd/new_*_test.go`

## 3. Problem Statement (Remaining Gaps)

### 3.1. Manual Registration Still Required

`new provider` and `new controller` still print manual TODO instructions instead of updating `module.go` automatically. This means generated code is not fully wired by default.

### 3.2. AST Registration Is Incomplete

- Provider AST utility exists but is not integrated into command flow.
- No equivalent controller AST insertion utility is integrated.
- Failure and idempotency behavior for repeated runs is not fully specified.

### 3.3. Delivery and Docs Gaps

- README does not yet describe CLI install and usage paths.
- Release pipeline does not yet publish CLI binaries (tracked in related SDLC spec).

## 4. Goals

- Make CLI scaffolding produce compilable, wired code with no manual TODO edits.
- Keep generated code explicit and idiomatic (no hidden runtime magic).
- Make command behavior deterministic and safe under repeated invocation.
- Keep compatibility with the project conventions around module structure and token naming.

## 5. Non-Goals

- Interactive TUI mode.
- `modkit graph` command.
- Project migration/upgrade assistants.
- Any reflection-based code generation or auto-discovery.

## 6. Detailed Requirements

## 6.1. Command Behavior Requirements

### `modkit new app <name>`

- Must continue to scaffold a runnable app skeleton.
- Must keep existing path safety and name validation behavior.
- Must keep deterministic template output.

### `modkit new module <name>`

- Must continue to create `internal/modules/<name>/module.go`.
- Must fail with clear error if module already exists.

### `modkit new provider <name> --module <module>`

- Must create provider file.
- Must register provider automatically in target module `Definition().Providers`.
- Must emit deterministic token naming based on module package convention.
- Must fail with actionable errors when:
  - `module.go` missing,
  - `Definition()` missing,
  - `Providers` cannot be safely updated.

### `modkit new controller <name> --module <module>`

- Must create controller file.
- Must register controller automatically in target module `Definition().Controllers`.
- Must fail with actionable errors when AST update cannot be applied safely.

## 6.2. AST Update Requirements

- Use `dave/dst`-based parsing and rewriting for both providers and controllers.
- Preserve comments and formatting as much as practical.
- Insertions must be idempotent:
  - Re-running the same command must not duplicate registration entries.
- Provide specific typed error messages for unsupported module shapes.
- Never silently skip registration when generation succeeded.

## 6.3. Idempotency and Safety

- File creation remains fail-fast when destination exists.
- Registration insertion must detect existing matching token/controller name before writing.
- Partial failure handling:
  - If file generation succeeds and AST registration fails, command must report exact follow-up action.
  - Prefer atomic write pattern for modified `module.go`.

## 6.4. Output and UX

- Success output should list created files and registration status.
- Error output should include target file path and failed operation.
- Remove manual TODO output once auto-registration is reliable.

## 7. Testing Plan

## 7.1. Unit Tests

Add/update tests for:

- Provider registration insertion and duplicate prevention.
- Controller registration insertion and duplicate prevention.
- Error paths for malformed `module.go` structures.
- Stability of generated token/controller naming.

## 7.2. Integration/Smoke Tests

Introduce CLI smoke checks in CI (as defined in the SDLC release spec):

1. Build CLI binary from `./cmd/modkit`.
2. Generate app, module, provider, controller.
3. Verify generated project compiles/tests successfully.
4. Ensure no manual edits are required before compile.

## 8. Documentation and Release Alignment

## 8.1. README

Add CLI install and quickstart:

- `go install github.com/go-modkit/modkit/cmd/modkit@latest`
- Release binary download option
- Minimal scaffold workflow

## 8.2. Release Artifacts

Keep this document scoped to CLI product behavior. Artifact publishing details remain in:

- `docs/specs/design-release-versioning-sdlc-cli.md`

## 9. Success Criteria

The PRD `modkit-cli` item is complete when all are true:

1. `new app/module/provider/controller` workflows generate code that compiles without manual registration edits.
2. Registration insertion for provider/controller is deterministic and idempotent.
3. CLI smoke tests are enforced in CI.
4. CLI install/usage documentation is present in README.
5. Release flow publishes CLI artifacts for tagged releases.

## 10. Risks and Mitigations

- Risk: AST insertion fails on uncommon `module.go` formatting.
  - Mitigation: strict error messages and tested fallback guidance.
- Risk: duplicate registrations from repeated command runs.
  - Mitigation: explicit duplicate detection before write.
- Risk: generated code drifts from framework conventions.
  - Mitigation: snapshot tests against templates and token patterns.

## 11. Rollout Plan

1. Integrate provider/controller AST registration in command flow.
2. Add/expand tests for insertion, idempotency, and error paths.
3. Add CI smoke job for generator compile checks.
4. Update README with CLI install and usage.
5. Enable release artifact publication (per SDLC release spec).

## 12. Out of Scope Follow-Ups

- Interactive CLI/TUI.
- Graph and devtools commands.
- Project migration generators.
