# Release Process

modkit uses a CI-gated release process powered by Release Please and GoReleaser.

## Release Flow

The release process follows a three-stage pipeline:

1.  **Continuous Integration (`ci.yml`)**:
    *   Runs on every pull request and merge to `main`.
    *   Executes linting, vulnerability scans, coverage tests, and CLI smoke tests.
    *   Must pass successfully on `main` to trigger the next stage.

2.  **Release PR / Tagging (`release.yml`)**:
    *   Triggered by `workflow_run` after `ci` completes on `main`.
    *   Runs only when CI concluded successfully for a `push` event on `main`.
    *   Uses `secrets.RELEASE_APP_ID` and `secrets.RELEASE_APP_PRIVATE_KEY` to mint a scoped GitHub App token.
    *   Runs `googleapis/release-please-action@v4` with `release-please-config.json` and `.release-please-manifest.json`.
    *   Release Please creates/updates a release PR containing version/changelog updates (`CHANGELOG.md`).
    *   When that release PR is merged, Release Please tags the release commit (`vX.Y.Z`) and creates the GitHub Release.

3.  **Artifact Publication (`release.yml`)**:
    *   Runs only when Release Please reports `release_created == true`.
    *   Uses a scoped GitHub App token, checks out the exact released tag, and runs GoReleaser.
    *   GoReleaser uses GitHub-native release notes (`changelog.use: github`) as changelog source of truth for release artifacts.

## Security Rationale

### GitHub App Bypass Actor
We use a dedicated GitHub App for release operations instead of a standard `GITHUB_TOKEN` or a Personal Access Token (PAT).
- **Scoped Permissions**: The app is granted only the minimum necessary permissions (`contents: write`).
- **Bypass Rules**: It acts as a "bypass actor" in repository rulesets, allowing it to push tags to protected branches without requiring a full PR cycle for the tag itself.
- **Auditability**: Release actions are clearly attributed to the App identity in the audit log.

### Release Please Control
Version/changelog commits are handled by Release Please via release PRs. The release workflow is hard-gated by successful CI completion on `main`, and artifact publication is conditioned on Release Please output in the same workflow.

## Quality Gates

Before creating or updating a PR, run:

```bash
make fmt && make lint && make vuln && make test && make test-coverage
make cli-smoke-build && make cli-smoke-scaffold
```

Branch protection should require these workflow checks on pull requests:

- `Validate PR Title`
- `Quality (Lint & Vuln)`
- `test`
- `CLI Smoke Test`
- `Analyze` (CodeQL)

## Versioning Rules

- `fix:` -> patch
- `feat:` -> minor
- `feat!:` / `BREAKING CHANGE:` -> major (or minor during initial development when configured)
- docs/chore/test/ci-only commits do not release by default
