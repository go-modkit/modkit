# Complete modkit CLI Implementation Plan

## TL;DR

> **Quick Summary**: Complete `modkit-cli` by replacing manual provider/controller registration with safe AST-based auto-registration, then close CI, release, and README gaps so generated projects compile without manual edits.
>
> **Deliverables**:
> - Fully wired `new provider` and `new controller` flows with idempotent AST insertion
> - Expanded unit tests for AST + command integration + malformed module shapes
> - CI `cli-smoke` job that proves generated code compiles/tests
> - Release pipeline updates to publish CLI binaries/checksums
> - README CLI install + quickstart guidance
>
> **Estimated Effort**: Large
> **Parallel Execution**: YES - 3 waves
> **Critical Path**: Task 1 -> Task 2 -> Task 3 -> Task 5 -> Task 6

---

## Context

### Original Request
Create a complete implementation plan first, then implement after user review.

### Interview Summary
**Key Discussions**:
- User requested exhaustive search and analysis mode before planning.
- User asked for confidence that background agents are actually running.
- User requested a complete implementation plan for full CLI remaining work.

**Research Findings**:
- Command scaffolding exists but provider/controller still print manual TODOs.
- Provider AST insertion helper exists and is currently not integrated.
- Controller AST insertion helper is missing.
- CI does not include CLI smoke generation/compile checks.
- Release workflow does not publish CLI binaries.
- Root README lacks CLI install/quickstart for `cmd/modkit`.

### Metis/Consultation Notes
Direct Metis/Oracle specialist sessions were invoked but returned no agent payload in this environment. Gap handling therefore uses direct repository evidence + official documentation references and explicit guardrails below.

### Manual High-Accuracy Review (Fallback)
- Momus high-accuracy review was requested but unavailable (no payload returned); manual strict review applied instead.
- Blocking gaps found and resolved in-plan:
  - Added explicit regression protection for `new app` and `new module` behavior continuity (Task 5 acceptance criteria).
  - Clarified duplicate-provider behavior target to deterministic idempotent outcome (no write, no duplicate).
  - Reinforced partial-failure requirement: generated file success + registration failure must always return actionable remediation.

---

## Work Objectives

### Core Objective
Deliver an end-to-end production-ready `modkit` CLI experience where `new app/module/provider/controller` produces compilable code without manual registration edits, and ensure CI/release/docs align with that behavior.

### Concrete Deliverables
- `internal/cli/ast` supports provider + controller insertion with idempotency.
- `internal/cli/cmd/new_provider.go` and `internal/cli/cmd/new_controller.go` integrate AST registration and remove manual TODO output.
- Unit tests cover insertion, duplicate prevention, malformed module shapes, and naming stability.
- `.github/workflows/ci.yml` adds `cli-smoke`.
- `.github/workflows/release.yml` + `.goreleaser.yml` publish CLI artifacts.
- `README.md` includes CLI install and quickstart.

### Definition of Done
- [x] `modkit new provider` and `modkit new controller` both auto-register in `Definition()`.
- [x] Re-running commands does not duplicate registrations.
- [x] CI enforces generator smoke compile path.
- [x] Tagged release path produces CLI artifacts + checksums.
- [x] README contains install and minimal scaffold workflow.

### Must Have
- Explicit typed errors for unsupported module shapes.
- Atomic write behavior for any `module.go` rewrite.
- No silent success when registration fails.

### Must NOT Have (Guardrails)
- No reflection/runtime auto-discovery.
- No weakening of existing quality gates (`lint`, `vuln`, `test-coverage`).
- No brittle string-only rewrites of `module.go` where AST safety is required.
- No manual TODO registration instructions in successful provider/controller paths.

---

## Verification Strategy (MANDATORY)

> **UNIVERSAL RULE: ZERO HUMAN INTERVENTION**
>
> All verification is agent-executed via commands/tools.

### Test Decision
- **Infrastructure exists**: YES (Go tests + CI workflow)
- **Automated tests**: YES (Tests-after default applied)
- **Framework**: `go test`

### Agent-Executed QA Scenarios (applies to all tasks)
- Use Bash for CLI/API/file verification.
- Use exact commands and deterministic temp directories.
- Include at least one negative scenario per feature.
- Capture outputs to `.sisyphus/evidence/task-{N}-{scenario}.txt`.

---

## Execution Strategy

### Mandatory Workspace Isolation (Git Worktree)

Before any implementation task, execution MUST happen in a new isolated worktree using `/superpowers/using-git-worktrees`.

Required preflight:
1. Announce skill usage and create isolated worktree for this plan branch.
2. Verify chosen worktree directory is gitignored if project-local (`.worktrees/` or `worktrees/`).
3. Run project setup in worktree (`go mod download` where applicable).
4. Run baseline tests in worktree before coding starts.
5. Proceed only after clean/known baseline is confirmed.

### Parallel Execution Waves

Wave 1 (Foundation):
- Task 1 (AST provider hardening)
- Task 4 (error/UX contract definitions)

Wave 2 (Feature integration):
- Task 2 (wire provider command)
- Task 3 (controller AST + wire command)
- Task 5 (unit test expansion)

Wave 3 (Delivery alignment):
- Task 6 (CI smoke)
- Task 7 (release artifacts)
- Task 8 (README)
- Task 9 (final verification + integration guard)

Critical Path: 1 -> 2 -> 3 -> 5 -> 6 -> 9

### Dependency Matrix

| Task | Depends On | Blocks | Can Parallelize With |
|---|---|---|---|
| 0 | None | 1,2,3,4,5,6,7,8,9 | None |
| 1 | None | 2,3,5 | 4 |
| 2 | 1 | 5,9 | 3 |
| 3 | 1 | 5,9 | 2 |
| 4 | None | 2,3 | 1 |
| 5 | 1,2,3 | 6,9 | None |
| 6 | 5 | 9 | 7,8 |
| 7 | 5 | 9 | 6,8 |
| 8 | 5 | 9 | 6,7 |
| 9 | 6,7,8 | None | None |

### Priority and Risk Order (Polish)

| Priority | Task | Reason | Risk if delayed |
|---|---|---|---|
| P0 | 1,2,3 | Core product promise: auto-registration without manual edits | CLI still incomplete and non-PRD-compliant |
| P0 | 5 | Locks behavior and prevents regressions while integrating AST changes | Hidden breakages ship into CI/release/docs work |
| P1 | 6 | Enforces behavior in CI for every PR | Regressions can reappear unnoticed |
| P1 | 7 | Enables actual distributable CLI artifacts | Feature complete but not deliverable |
| P1 | 8 | User-facing install/usage alignment | Adoption friction and support overhead |
| P2 | 4 | UX message consistency hardening | Functional but less actionable errors |
| P2 | 9 | Final global confidence gate | Late discovery of integration drift |

### Tight Sequencing (Fastest Safe Path)

1. **Stabilize AST core first**: complete Task 1 and immediately run focused AST tests.
2. **Integrate command flows in one wave**: execute Tasks 2 and 3, then run command-level tests.
3. **Lock behavior before delivery plumbing**: complete Task 5 before touching CI/release/docs.
4. **Ship safety rails**: add Task 6 (`cli-smoke`) before Task 7 (artifact publishing).
5. **Finish outward-facing docs**: complete Task 8 only after command signatures and UX messages are stable.
6. **Run final gate once**: execute Task 9 as single final pass to avoid redundant expensive runs.

### Merge and Rollback Guardrails

- Merge order should follow P0 -> P1 -> P2; do not merge release workflow changes before `cli-smoke` is green.
- If AST integration causes command instability, rollback only Task 2/3 callsites while preserving Task 1 testable utility improvements.
- If release workflow fails while core CLI passes, keep feature branch releasable by gating GoReleaser step on semantic-release output and preserving current release behavior as fallback.

---

## TODOs

- [x] 0. Create isolated git worktree and verify clean baseline (MANDATORY)

  **What to do**:
  - Run `/superpowers/using-git-worktrees` to create a dedicated worktree for this implementation.
  - Ensure worktree path is ignored if project-local.
  - Run baseline setup and tests in new worktree before any code changes.

  **Must NOT do**:
  - Do not implement in the current dirty/main workspace.
  - Do not start feature tasks before baseline verification is complete.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `using-git-worktrees`, `git-master`
    - `using-git-worktrees`: enforce safe isolation workflow.
    - `git-master`: ensure branch/worktree hygiene.

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Sequential preflight
  - **Blocks**: 1,2,3,4,5,6,7,8,9
  - **Blocked By**: None

  **References**:
  - `/superpowers/using-git-worktrees` - required isolation protocol.
  - `AGENTS.md` - repository workflow and verification expectations.

  **Acceptance Criteria**:
  - [ ] New worktree path exists and active branch is dedicated to this plan.
  - [ ] Baseline setup completed in that worktree.
  - [ ] Baseline tests executed and result recorded before Task 1 starts.

- [x] 1. Harden provider AST insertion utility for idempotent safe rewrites

  **What to do**:
  - Refactor `AddProvider` to detect existing `Token` before append.
  - Add typed errors for parse/unsupported shape/missing `Definition`/missing `Providers`.
  - Preserve atomic rewrite (`CreateTemp` + `Rename`) and file mode.

  **Must NOT do**:
  - Do not fallback to fragile regex/string append in `module.go`.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: AST mutation + robust error taxonomy.
  - **Skills**: `git-master`
    - `git-master`: maintain atomic, reviewable, isolated deltas.

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 4)
  - **Blocks**: 2, 3, 5
  - **Blocked By**: None

  **References**:
  - `internal/cli/ast/modify.go` - current provider insertion implementation.
  - `internal/cli/ast/modify_test.go` - current coverage and gaps.
  - `docs/specs/design-cli-tooling.md` - idempotency and typed error requirements.
  - `https://pkg.go.dev/github.com/dave/dst/decorator` - parse/print/restore APIs and constraints.

  **Acceptance Criteria**:
  - [ ] Duplicate provider token is detected and command result is idempotent (no additional write, no duplicate append).
  - [ ] Unsupported module shape returns actionable typed error.
  - [ ] Write path remains atomic and preserves file mode.

  **Agent-Executed QA Scenarios**:
  ```text
  Scenario: Provider insertion on empty Providers list
    Tool: Bash (go test)
    Preconditions: Test fixture module.go with empty Providers
    Steps:
      1. Run: go test ./internal/cli/ast -run TestAddProvider
      2. Assert: target test reports PASS
      3. Assert: resulting file contains one Token entry
    Expected Result: Single provider registration added
    Evidence: .sisyphus/evidence/task-1-provider-insert.txt

  Scenario: Duplicate token does not duplicate registration
    Tool: Bash (go test)
    Preconditions: Fixture already contains users.auth
    Steps:
      1. Run duplicate insertion test
      2. Assert: Providers length unchanged
      3. Assert: no second Token: "users.auth"
    Expected Result: Idempotent behavior
    Evidence: .sisyphus/evidence/task-1-provider-duplicate.txt
  ```

- [x] 2. Integrate provider auto-registration into `new provider` command flow

  **What to do**:
  - Call AST registration utility from `createNewProvider` after file generation.
  - Replace manual TODO output with deterministic success/error status output.
  - On registration failure after file creation, return explicit next-step guidance including module path and exact failure.

  **Must NOT do**:
  - Do not report success when registration fails.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: `git-master`

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Task 3)
  - **Blocks**: 5, 9
  - **Blocked By**: 1

  **References**:
  - `internal/cli/cmd/new_provider.go` - current generation flow and TODO output.
  - `internal/cli/cmd/new_provider_test.go` - command behavior patterns.
  - `docs/specs/design-cli-tooling.md:81` - provider behavior requirements.

  **Acceptance Criteria**:
  - [ ] Success output includes created file and registration status.
  - [ ] Manual registration TODO lines are removed.
  - [ ] Failure output includes target `module.go` and operation details.

  **Agent-Executed QA Scenarios**:
  ```text
  Scenario: Provider command auto-registers in module Definition
    Tool: Bash
    Preconditions: Temp app + module exists
    Steps:
      1. Run: modkit new provider auth --module users
      2. Assert: internal/modules/users/auth.go exists
      3. Assert: internal/modules/users/module.go contains Token: "users.auth"
    Expected Result: File + registration both present
    Evidence: .sisyphus/evidence/task-2-provider-command-success.txt

  Scenario: Malformed module.go returns actionable error
    Tool: Bash
    Preconditions: Corrupt users/module.go fixture
    Steps:
      1. Run provider command
      2. Assert: non-zero exit
      3. Assert: stderr includes module path + failed AST operation
    Expected Result: Actionable failure with follow-up guidance
    Evidence: .sisyphus/evidence/task-2-provider-command-fail.txt
  ```

- [x] 3. Implement controller AST insertion and integrate `new controller`

  **What to do**:
  - Add controller insertion utility mirroring provider safety/idempotency behavior.
  - Integrate utility into `createNewController`.
  - Remove manual TODO output in controller flow.

  **Must NOT do**:
  - Do not introduce divergent insertion semantics from provider behavior.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: `git-master`

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Task 2)
  - **Blocks**: 5, 9
  - **Blocked By**: 1

  **References**:
  - `internal/cli/cmd/new_controller.go` - command integration point.
  - `internal/cli/cmd/new_controller_test.go` - behavior and error patterns.
  - `examples/hello-simple/main.go` - canonical `Controllers` shape pattern.
  - `docs/specs/design-cli-tooling.md:91` - controller requirements.

  **Acceptance Criteria**:
  - [ ] Controller is inserted into `Definition().Controllers` safely.
  - [ ] Duplicate controller entries are prevented.
  - [ ] Unsupported shape produces typed actionable error.

  **Agent-Executed QA Scenarios**:
  ```text
  Scenario: Controller command auto-registers controller def
    Tool: Bash
    Preconditions: Temp module with valid Definition().Controllers list
    Steps:
      1. Run: modkit new controller auth --module users
      2. Assert: auth_controller.go exists
      3. Assert: module.go contains Name: "AuthController"
    Expected Result: Controller file + registration complete
    Evidence: .sisyphus/evidence/task-3-controller-success.txt

  Scenario: Duplicate controller registration prevented
    Tool: Bash
    Preconditions: Existing AuthController entry in module.go
    Steps:
      1. Re-run controller registration path
      2. Assert: no duplicate Name entries added
    Expected Result: Idempotent controller insertion
    Evidence: .sisyphus/evidence/task-3-controller-duplicate.txt
  ```

- [x] 4. Define and normalize UX/error contracts for partial failure paths

  **What to do**:
  - Standardize success/error messages across `new provider` and `new controller`.
  - Ensure partial failure text includes exact remediation (what was created, what failed, what command/line to apply manually if needed).

  **Must NOT do**:
  - No ambiguous errors like "failed" without file + operation context.

  **Recommended Agent Profile**:
  - **Category**: `writing`
  - **Skills**: `git-master`

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1
  - **Blocks**: 2, 3
  - **Blocked By**: None

  **References**:
  - `docs/specs/design-cli-tooling.md:114` - output and UX requirements.
  - `internal/cli/cmd/new_provider.go`
  - `internal/cli/cmd/new_controller.go`

  **Acceptance Criteria**:
  - [ ] Output includes created file + registration status.
  - [ ] Error output always includes target file path and failed operation.

  **Agent-Executed QA Scenarios**:
  ```text
  Scenario: Partial failure message includes remediation
    Tool: Bash
    Preconditions: Inject AST failure fixture
    Steps:
      1. Execute provider/controller command
      2. Assert: stderr includes created file path + module.go path + clear follow-up action
    Expected Result: Actionable, deterministic error contract
    Evidence: .sisyphus/evidence/task-4-ux-errors.txt
  ```

- [x] 5. Expand automated tests for insertion, idempotency, malformed shapes, naming stability

  **What to do**:
  - Add/expand AST unit tests for provider + controller insertion and duplicate prevention.
  - Expand command tests to assert module registration mutation.
  - Add malformed module shape fixtures and assert typed errors.

  **Must NOT do**:
  - Do not rely solely on golden output strings without structure assertions.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: `git-master`

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 2 (after 1/2/3)
  - **Blocks**: 6, 7, 8, 9
  - **Blocked By**: 1, 2, 3

  **References**:
  - `internal/cli/ast/modify_test.go`
  - `internal/cli/cmd/new_provider_test.go`
  - `internal/cli/cmd/new_controller_test.go`
  - `internal/cli/cmd/naming.go`

  **Acceptance Criteria**:
  - [ ] Tests cover successful insertion and duplicate prevention for both providers/controllers.
  - [ ] Tests cover malformed `module.go` structures and typed errors.
  - [ ] Regression tests confirm `new app` and `new module` continue existing behavior and path-safety checks.
  - [ ] `go test ./internal/cli/...` passes.

  **Agent-Executed QA Scenarios**:
  ```text
  Scenario: Internal CLI test suite passes
    Tool: Bash
    Steps:
      1. Run: go test ./internal/cli/...
      2. Assert: exit code 0
      3. Assert: no failing tests for new_provider/new_controller/ast
    Expected Result: Full CLI unit coverage green
    Evidence: .sisyphus/evidence/task-5-cli-tests.txt
  ```

- [x] 6. Add CI `cli-smoke` workflow job for generator compile checks

  **What to do**:
  - Add `cli-smoke` job in `.github/workflows/ci.yml`.
  - Build CLI from `./cmd/modkit`, scaffold app/module/provider/controller in temp dir, run `go test ./...` in generated project.
  - Keep existing jobs unchanged.

  **Must NOT do**:
  - Do not remove/relax existing quality jobs.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: `git-master`

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: 9
  - **Blocked By**: 5

  **References**:
  - `.github/workflows/ci.yml`
  - `docs/specs/design-release-versioning-sdlc-cli.md:59`

  **Acceptance Criteria**:
  - [ ] CI includes `cli-smoke` with scaffold+compile validation.
  - [ ] Existing `quality` and `test` jobs remain present.

  **Agent-Executed QA Scenarios**:
  ```text
  Scenario: Local emulation of cli-smoke sequence passes
    Tool: Bash
    Steps:
      1. Build CLI binary from cmd/modkit
      2. Scaffold demo app/module/provider/controller in temp dir
      3. Run go test ./... in generated app
      4. Assert all commands exit 0
    Expected Result: Generated project compiles/tests without manual edits
    Evidence: .sisyphus/evidence/task-6-cli-smoke.txt
  ```

- [x] 7. Publish CLI artifacts in release flow using GoReleaser

  **What to do**:
  - Add `.goreleaser.yml` for `./cmd/modkit` target matrix and checksums.
  - Update `.github/workflows/release.yml` to conditionally run GoReleaser when semantic release emits a version.
  - Pin/setup Go and preserve `contents: write` permission requirements.

  **Must NOT do**:
  - Do not run artifact publication when semantic-release returns empty version.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: `git-master`

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: 9
  - **Blocked By**: 5

  **References**:
  - `.github/workflows/release.yml`
  - `docs/specs/design-release-versioning-sdlc-cli.md:95`
  - `https://goreleaser.com/ci/actions/`
  - `https://github.com/go-semantic-release/action`

  **Acceptance Criteria**:
  - [ ] Release workflow conditionally runs GoReleaser for non-empty semrel output.
  - [ ] Artifacts include required platform matrix and `checksums.txt`.

  **Agent-Executed QA Scenarios**:
  ```text
  Scenario: Workflow config validates release gating
    Tool: Bash
    Steps:
      1. Inspect release workflow conditions for semrel output check
      2. Assert goreleaser step guarded by non-empty version condition
      3. Assert GITHUB_TOKEN present in release env
    Expected Result: Safe conditional artifact publishing logic
    Evidence: .sisyphus/evidence/task-7-release-workflow.txt
  ```

- [x] 8. Update README with CLI install and quickstart workflow

  **What to do**:
  - Add `go install github.com/go-modkit/modkit/cmd/modkit@latest` path.
  - Add release-binary install option and minimal `new app/module/provider/controller` quickstart.

  **Must NOT do**:
  - Do not remove existing library install guidance; add CLI guidance clearly.

  **Recommended Agent Profile**:
  - **Category**: `writing`
  - **Skills**: `git-master`

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: 9
  - **Blocked By**: 5

  **References**:
  - `README.md`
  - `docs/specs/design-cli-tooling.md:140`

  **Acceptance Criteria**:
  - [ ] README includes CLI install and quickstart commands.
  - [ ] Instructions align with generated paths and current command names.

  **Agent-Executed QA Scenarios**:
  ```text
  Scenario: README command path consistency check
    Tool: Bash
    Steps:
      1. Compare README CLI commands against actual command tree and binary path
      2. Assert no mismatched command names/flags
    Expected Result: Docs match implementation
    Evidence: .sisyphus/evidence/task-8-readme-consistency.txt
  ```

- [x] 9. Final integrated verification and regression guard pass

  **What to do**:
  - Run full project verification commands required by repository policy.
  - Run CLI smoke flow end-to-end once more after all changes.
  - Reconcile checklist evidence.

  **Must NOT do**:
  - Do not claim completion without command outputs and evidence artifacts.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: `git-master`

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Final sequential
  - **Blocks**: None
  - **Blocked By**: 2,3,6,7,8

  **References**:
  - `AGENTS.md` command policy (`make fmt`, `make lint`, `make vuln`, `make test`, `make test-coverage`).
  - `.github/workflows/ci.yml` and `.github/workflows/release.yml`.

  **Acceptance Criteria**:
  - [ ] `make fmt` succeeds.
  - [ ] `make lint` succeeds.
  - [ ] `make vuln` succeeds.
  - [ ] `make test` succeeds.
  - [ ] `make test-coverage` succeeds.
  - [ ] CLI smoke flow passes with no manual edits.

  **Agent-Executed QA Scenarios**:
  ```text
  Scenario: Full verification gate pass
    Tool: Bash
    Steps:
      1. Run make fmt && make lint && make vuln && make test && make test-coverage
      2. Assert all exit codes are 0
      3. Run cli-smoke equivalent and assert green
    Expected Result: Release-ready confidence
    Evidence: .sisyphus/evidence/task-9-full-verification.txt
  ```

---

## Commit Strategy

| After Task | Message | Files | Verification |
|---|---|---|---|
| 1-3 | `feat(cli): auto-register providers and controllers via AST` | `internal/cli/ast/*`, `internal/cli/cmd/new_{provider,controller}.go` | `go test ./internal/cli/...` |
| 5 | `test(cli): cover idempotent insertion and malformed module shapes` | `internal/cli/**/*_test.go` | `go test ./internal/cli/...` |
| 6-7 | `ci(release): add cli smoke checks and publish CLI artifacts` | `.github/workflows/*`, `.goreleaser.yml` | workflow config + smoke script checks |
| 8 | `docs(readme): add CLI install and quickstart` | `README.md` | command consistency check |

---

## Success Criteria

### Verification Commands
```bash
go test ./internal/cli/...
make fmt && make lint && make vuln && make test && make test-coverage
```

### Final Checklist
- [x] `new app/module/provider/controller` generate compilable code without manual registration edits.
- [x] Provider/controller insertion is deterministic and idempotent.
- [x] CI enforces CLI smoke checks.
- [x] Release publishes CLI artifacts/checksums for configured targets.
- [x] README includes CLI install and usage.
