# Release Process

modkit uses a CI-gated release workflow so semantic versioning and artifact publishing stay in one controlled pipeline.

## Release Flow

The release process follows a two-stage pipeline:

1.  **Continuous Integration (`ci.yml`)**:
    *   Runs on every pull request and merge to `main`.
    *   Executes linting, vulnerability scans, coverage tests, and CLI smoke tests.
    *   Must pass successfully on `main` to trigger the next stage.

2.  **Release (`release.yml`)**:
    *   Triggered via `workflow_run` when the `ci` workflow completes successfully on a `main` push.
    *   **Semantic job**: Uses `secrets.RELEASE_APP_ID` and `secrets.RELEASE_APP_PRIVATE_KEY` to create a scoped GitHub App token, checks out the repository with that token, and runs `go-semantic-release`.
    *   **Artifacts job**: Runs only if semantic release emits a version, creates a scoped GitHub App token, checks out the created release tag (`v<version>`), and runs GoReleaser.
    *   **Changelog Ownership**: GoReleaser uses GitHub-native release notes (`use: github`) as the source of truth for changelog content.

## Security Rationale

### GitHub App Bypass Actor
We use a dedicated GitHub App for release operations instead of a standard `GITHUB_TOKEN` or a Personal Access Token (PAT).
- **Scoped Permissions**: The app is granted only the minimum necessary permissions (`contents: write`).
- **Bypass Rules**: It acts as a "bypass actor" in repository rulesets, allowing it to push tags to protected branches without requiring a full PR cycle for the tag itself.
- **Auditability**: Release actions are clearly attributed to the App identity in the audit log.

### Single Workflow Control
Semantic tagging and artifact publishing run in one workflow so artifact publication is conditioned on a released semantic version output rather than on an independent tag-triggered workflow.

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
