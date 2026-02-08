# Design Spec: SDLC Flow for Versioning and CLI Releases

**Status:** Implemented (v1)
**Date:** 2026-02-07
**Author:** Sisyphus (AI Agent)
**Last Reviewed:** 2026-02-08
**Applies After:** N/A (landed on `main`)

## 1. Overview

This spec defines the software delivery lifecycle (SDLC) updates required for `modkit` now that the CLI (`cmd/modkit`) exists.

Current release flow creates semantic versions and release notes, but it does not publish CLI binaries. The updated flow must produce versioned Go module releases and downloadable CLI artifacts from the same release event, with CI gates that prevent shipping broken generators.

## 2. Goals

- Keep semantic versioning behavior driven by Conventional Commits.
- Publish `modkit` CLI artifacts for supported platforms on each semantic release.
- Add CI checks to validate CLI generation commands and generated project compilation.
- Keep release process deterministic and auditable.
- Preserve current project quality gates and PR workflow conventions.

## 3. Non-Goals

- No Homebrew tap automation in this phase.
- No package manager publishing (apt, rpm, npm, etc.) in this phase.
- No signing/provenance requirements in this phase (can be added later).

## 4. Current State

- `release.yml` runs semantic release on pushes to `main`.
- `ci.yml` runs lint, vulnerability scan, tests, and coverage.
- CLI commands exist but are not part of release artifact publishing.

## 5. Target SDLC Flow

### 5.1. Development and PR

1. Developer creates branch (`feat/*`, `fix/*`, `chore/*`).
2. Developer opens PR with semantic title and template checklist.
3. CI validates code quality and CLI smoke generation.
4. PR merges to `main` after required checks pass.

### 5.2. Release

1. `release.yml` is triggered by push to `main`.
2. Semantic release computes next version and creates git tag/release notes when applicable.
3. If a new version is created:
   - Build CLI for matrix targets.
   - Create archives and checksums.
   - Upload assets to the corresponding GitHub release.

### 5.3. Post-release

1. Release page contains changelog and artifacts.
2. README install section points to release binaries and `go install` path.

## 6. Required CI Changes

## 6.1. Maintain CLI Smoke-Test Job in `ci.yml`

`cli-smoke` is already present in `ci.yml` and must remain a required check on pull requests and pushes.

Validation sequence:

```bash
go build -o dist/modkit ./cmd/modkit

TMP_DIR="$(mktemp -d)"
cd "$TMP_DIR"

/path/to/modkit new app demo
cd demo

/path/to/modkit new module user-service
/path/to/modkit new provider auth --module user-service
/path/to/modkit new controller auth --module user-service

go test ./...
```

Requirements:

- Job must fail if any generator outputs uncompilable code.
- Job must run after setup-go and tool installation.
- Temporary directories are cleaned up during workflow execution.

## 6.2. Keep Existing Quality Gates

Do not remove or weaken existing jobs:

- `make lint`
- `make vuln`
- `make test-coverage`

## 7. Required Release Changes

## 7.1. Artifact Publishing Strategy

Use GoReleaser for cross-platform packaging and GitHub release asset upload.

Add `.goreleaser.yml` with:

- Project name: `modkit`
- Build target: `./cmd/modkit`
- OS/ARCH matrix:
  - `darwin/amd64`
  - `darwin/arm64`
  - `linux/amd64`
  - `linux/arm64`
  - optional `windows/amd64`
- Archive naming template including version, OS, and architecture.
- `checksums.txt` generation.
- Changelog source set to use git history (or release notes from semantic release, depending on chosen integration).

## 7.2. Update `release.yml`

After semantic release step:

- Detect whether `steps.semrel.outputs.version` is non-empty.
- If empty: skip artifact publishing.
- If non-empty:
  1. `actions/setup-go` with pinned version `1.25.7`.
  2. Install GoReleaser (official action or pinned binary).
  3. Run GoReleaser in release mode against the tag created by semantic release.

Pseudo-flow:

```yaml
- uses: go-semantic-release/action@v1
  id: semrel

- name: Setup Go
  if: steps.semrel.outputs.version != ''
  uses: actions/setup-go@v6
  with:
    go-version: "1.25.7"

- name: Run GoReleaser
  if: steps.semrel.outputs.version != ''
  uses: goreleaser/goreleaser-action@v6
  with:
    version: latest
    args: release --clean
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## 8. Versioning Policy

Retain semantic release mapping:

- `fix:` -> patch
- `feat:` -> minor
- `feat!` or `BREAKING CHANGE:` -> major
- docs/chore/test-only commits should not trigger version bump unless explicitly configured.

## 9. Documentation Updates Required

After workflow changes are merged:

1. Update `README.md` with CLI install options:
   - Download binary from GitHub Releases.
   - `go install github.com/go-modkit/modkit/cmd/modkit@latest`.
2. Add a short release process note under `docs/` describing:
   - How versions are cut.
   - Where artifacts are published.
   - What checks gate release.

## 10. Acceptance Criteria

This initiative is complete when all are true:

1. A PR that breaks generated code fails in CI `cli-smoke` job.
2. A semantic release from `main` publishes CLI binaries and checksums to GitHub Release assets.
3. Published assets include all required target platforms.
4. Existing lint/vuln/test/coverage gates continue to pass.
5. README includes CLI install instructions.

## 11. Rollout Plan

### Phase 1: CI Guardrails

- Confirm `cli-smoke` job remains enforced in `ci.yml`.
- Validate against current `main`.

### Phase 2: Release Artifacts

- Add `.goreleaser.yml`.
- Update `release.yml` to publish assets only when version is released.

### Phase 3: Docs

- Update README and release process docs.

## 12. Risks and Mitigations

- Risk: Release workflow runs without a semantic version and fails artifact steps.
  - Mitigation: Guard artifact steps with `if: steps.semrel.outputs.version != ''`.

- Risk: CLI smoke tests become flaky due to temp-dir assumptions.
  - Mitigation: Use deterministic paths and explicit cleanup in workflow.

- Risk: Added release tooling increases maintenance burden.
  - Mitigation: Keep `.goreleaser.yml` minimal and version pin action dependencies.

## 13. Implementation PR Checklist

The follow-up PR (after `feat/cli-tooling` merge) should include:

- [x] `cli-smoke` job in `.github/workflows/ci.yml` is preserved and remains required.
- [x] `.github/workflows/release.yml` updated for CLI artifact publishing.
- [x] `.goreleaser.yml` added and validated.
- [x] README install section updated.
- [ ] `make fmt && make lint && make vuln && make test && make test-coverage` run and recorded.
