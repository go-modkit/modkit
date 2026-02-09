# Maintainer Operations

This guide defines maintainer workflow expectations for OSS reliability and adoption.

## Triage and Response Targets

- New issue first response target: within 3 business days.
- New discussion first response target: within 5 business days.
- Security reports: acknowledge as soon as possible and follow `SECURITY.md` process.

## Issue Label Flow

Apply one primary type label:

- `bug`
- `enhancement`
- `documentation`
- `question`

Then apply one priority label:

- `priority:high`
- `priority:medium`
- `priority:low`

Then apply onboarding labels when relevant:

- `good first issue`
- `help wanted`

## Monthly OSS Adoption Review

Run once per month and record outcomes in a maintainer note.

### KPI Set

- Quickstart success signal: count of quickstart-related issues/regressions.
- Documentation friction signal: recurring confusion points in issues/discussions.
- Time to first maintainer response for issues.
- Time from issue opened to closed for `bug` and `documentation` labels.

### Review Cadence and Owners

- Cadence: monthly.
- Suggested owner rotation: one maintainer per month.
- Output: short note with top regressions and next actions.

## Release-Adjacent Operations

- Keep release process aligned with `docs/guides/release-process.md`.
- Do not bypass CI/release guardrails for convenience.
- Prefer forward fixes in follow-up releases over forceful history surgery.
