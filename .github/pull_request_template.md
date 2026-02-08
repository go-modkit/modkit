## Type

<!-- Check ALL that apply -->
- [ ] `feat` — New feature
- [ ] `fix` — Bug fix
- [ ] `refactor` — Code restructure (no behavior change)
- [ ] `docs` — Documentation only
- [ ] `test` — Test coverage
- [ ] `chore` — Build, CI, tooling
- [ ] `perf` — Performance improvement

## Summary

<!-- One paragraph explaining WHAT this PR does and WHY -->

## Changes

<!-- Bullet list of specific changes made -->
- 

## Breaking Changes

<!-- If none, write "None" -->

## Validation

<!-- Commands run and their results -->
```bash
make fmt && make lint && make vuln && make test && make test-coverage
make cli-smoke-build && make cli-smoke-scaffold
```

## Checklist

<!-- All boxes should be checked before requesting review -->
- [ ] Code follows project style (`make fmt` passes)
- [ ] Linter passes (`make lint`)
- [ ] Vulnerability scan passes (`make vuln`)
- [ ] Tests pass (`make test`)
- [ ] Coverage tests pass (`make test-coverage`)
- [ ] CLI smoke checks pass (`make cli-smoke-build && make cli-smoke-scaffold`)
- [ ] Tests added/updated for new functionality
- [ ] Documentation updated (if applicable)
- [ ] Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/)

## Resolves

<!-- If implementing a GitHub issue, add "Resolves #<number>" for each issue -->
<!-- For sub-issues, add separate "Resolves #<sub-issue>" lines -->
<!-- If work is not tied to an issue, delete this section -->

Resolves #

## Notes

<!-- Optional: anything reviewers should know -->
