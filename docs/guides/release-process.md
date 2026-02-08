# Release Process

modkit uses a split, gated release architecture to ensure that only verified code is released and that release actors have minimal, scoped permissions.

## Release Flow

The release process follows a three-stage pipeline:

1.  **Continuous Integration (`ci.yml`)**:
    *   Runs on every pull request and merge to `main`.
    *   Executes linting, vulnerability scans, coverage tests, and CLI smoke tests.
    *   Must pass successfully on `main` to trigger the next stage.

2.  **Semantic Versioning (`release-semantic.yml`)**:
    *   Triggered via `workflow_run` when the CI workflow completes successfully on the `main` branch.
    *   **GitHub App Authentication**: Uses `vars.RELEASE_APP_ID` and `secrets.RELEASE_APP_PRIVATE_KEY` to generate a scoped installation token. This allows the workflow to bypass branch protection rules (specifically to push tags) using a dedicated service identity rather than a personal access token.
    *   **SHA Drift Guard**: Before proceeding, the workflow verifies that the current `origin/main` SHA matches the SHA that triggered the CI run. This prevents "drift" where a new commit is pushed to `main` before the previous release completes, ensuring we only tag exactly what was tested.
    *   **Version Detection**: Runs `go-semantic-release` to analyze Conventional Commits. If a new version is required, it creates and pushes a git tag (e.g., `v1.2.3`).

3.  **Artifact Publication (`release-artifacts.yml`)**:
    *   Triggered by the push of a version tag (`v*`).
    *   **GoReleaser**: Builds binaries for multiple platforms (`darwin`, `linux`, `windows` for both `amd64` and `arm64`).
    *   **Changelog Ownership**: GoReleaser is configured to use GitHub-native release notes (`use: github`) as the source of truth for the changelog.
    *   Publishes the release with binaries and `checksums.txt` to GitHub.

## Security Rationale

### GitHub App Bypass Actor
We use a dedicated GitHub App for release operations instead of a standard `GITHUB_TOKEN` or a Personal Access Token (PAT).
- **Scoped Permissions**: The app is granted only the minimum necessary permissions (`contents: write`).
- **Bypass Rules**: It acts as a "bypass actor" in repository rulesets, allowing it to push tags to protected branches without requiring a full PR cycle for the tag itself.
- **Auditability**: Release actions are clearly attributed to the App identity in the audit log.

### SHA Drift Guard
In a high-velocity environment, multiple commits might land on `main` in quick succession. The SHA drift guard ensures that the release workflow only tags the specific commit that passed CI. If `main` has moved, the workflow fails safely, and the next CI run (for the newer commit) will handle the release.

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
