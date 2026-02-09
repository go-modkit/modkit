# Design Spec: GitHub App Bypass + Release Please + GoReleaser Artifacts

## Status

- Type: design spec
- Scope: `go-modkit/modkit`
- Goal: standard CI-gated release flow with release PRs, tag-on-version-commit semantics, and automated artifact publishing

## Constraints

- Keep existing cache setup in workflows exactly as-is.
- Go version remains pinned to `1.25.7` in CI.
- Public repository: remove `CODECOV_TOKEN` secret usage and workflow references.
- Release must be gated on successful CI for `main`.
- Use GitHub App token for release automation under ruleset bypass.

## Desired Release Architecture

1. `ci` workflow runs on pull requests and pushes to `main`.
2. A single `release` workflow runs via `workflow_run` after `ci` success on `main` push and executes Release Please.
3. Release Please maintains a release PR with changelog/version updates and, on merge, creates tag `vX.Y.Z` on that version commit.
4. Artifact publishing runs only when Release Please reports `release_created == true`.
5. GoReleaser publishes artifacts from the created tag and uses GitHub release notes/changelog ownership.

This keeps responsibilities clear:

- Release Please: version/changelog PR + release tag creation.
- GoReleaser: GitHub Release content + artifacts + changelog generation.

## Organization-Level Secrets and Variables

Use org-level credentials (shared across repos):

- Secret: `RELEASE_APP_ID=<app id>`
- Secret: `RELEASE_APP_PRIVATE_KEY=<PEM>`

Workflow usage requirements:

- Read app credentials from `${{ secrets.RELEASE_APP_ID }}` and `${{ secrets.RELEASE_APP_PRIVATE_KEY }}`.
- Scope generated app token to target repo when creating token:
  - `owner: go-modkit`
  - `repositories: modkit`

## Ruleset / Bypass Requirements

- GitHub App `go-modkit-release` must be a ruleset bypass actor for `main`.
- Keep required checks in place for normal contributor flow.
- Bypass is only for release automation using app token.

## Workflow Specs

### A) CI Workflow (`.github/workflows/ci.yml`)

Keep existing jobs/caching unchanged. Apply only these updates:

1. Keep coverage generation output at `.coverage/coverage.out`.
2. Upload coverage with `codecov/codecov-action@v5` without token input.
3. Set strict failure behavior: `fail_ci_if_error: true`.
4. Use OIDC for Codecov auth in public repo:
   - Add `use_oidc: true` on Codecov step.
   - Ensure permissions include `id-token: write` and `contents: read`.
5. Remove any reference to `${{ secrets.CODECOV_TOKEN }}`.

### B) Release Workflow (`.github/workflows/release.yml`)

Trigger:

- `on.workflow_run.workflows: ["ci"]`
- `types: [completed]`
- `branches: [main]`

Release orchestration:

- Gate must require all of:
  - `github.event.workflow_run.conclusion == 'success'`
  - `github.event.workflow_run.event == 'push'`
  - `github.event.workflow_run.head_branch == 'main'`
- Run `googleapis/release-please-action@v4` using:
  - `config-file: release-please-config.json`
  - `manifest-file: .release-please-manifest.json`
- Use GitHub App token for action authentication.

Core steps:

1. Create GitHub App installation token via `actions/create-github-app-token@v2` using org var/secret and owner/repositories scoping.
2. Run Release Please action to create/update release PR and create release on merge.
3. Capture outputs (`release_created`, `tag_name`, `version`).
4. Run artifact publish job only when `release_created == true`.
5. In artifacts job, generate scoped GitHub App token and publish artifacts from `tag_name` via `goreleaser/goreleaser-action@v6`.

Required controls:

- Top-level workflow permissions remain least-privilege (`contents: read`), with job-level elevation to `contents: write` where required.
- `concurrency` single lane per release ref (`cancel-in-progress: false`).
- Keep current Go setup and cache blocks unchanged in the artifacts job.

### C) Release Please Configuration

Required files:

- `release-please-config.json`
- `.release-please-manifest.json`

Configuration requirements:

- Root package (`."`) uses `release-type: go`.
- Changelog path is `CHANGELOG.md`.
- Manifest tracks current released version baseline for root package.

## `.goreleaser.yml` Changelog Ownership

GoReleaser is the source of truth for release notes/changelog.

Required change:

- Enable changelog generation and set:

```yaml
changelog:
  use: github
```

Additional rules:

- Do not disable changelog in `.goreleaser.yml`.
- Keep existing build/archive/checksum settings unless needed for compatibility.

## Documentation Update

Update `docs/guides/release-process.md` to describe:

1. CI hard gate via branch protection and required checks before merge to `main`.
2. Release Please release PR lifecycle and tag creation semantics.
3. Artifact publication conditioned on Release Please outputs in the same workflow.
4. Org-level app secret convention.
5. GoReleaser-generated changelog ownership.

## Acceptance Criteria

1. PRs and pushes to `main` run `ci` with existing caching unchanged.
2. Codecov upload runs without `CODECOV_TOKEN` and fails CI on upload errors.
3. Release workflow runs only after successful `ci` completion on `main` push and uses Release Please with config/manifest files.
4. Release PR includes changelog/version updates and merged release commit is tagged `vX.Y.Z`.
5. Artifact publishing runs only when Release Please reports `release_created == true`, uses app-token auth, and publishes from emitted tag.
6. GoReleaser publishes assets and generates release notes/changelog using `changelog.use: github`.
