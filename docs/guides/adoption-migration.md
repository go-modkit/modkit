# Adoption and Migration Guide

This guide is for teams adopting modkit in existing services without a full rewrite.

## Adoption Strategy: Start Small

Use one bounded module first, then expand.

Recommended order:

1. Introduce one feature module (for example, `users`) with explicit exports.
2. Keep existing router and handlers around unaffected routes.
3. Move one route group to a modkit controller.
4. Expand module-by-module.

## Incremental Integration Patterns

### Pattern A: Keep Existing `chi` Router

- Keep your current `chi.Mux` as the main router.
- Bootstrap modkit modules for new route groups.
- Register modkit controllers into your existing route tree.

### Pattern B: Keep Existing Data Layer

- Reuse existing DB/repository types.
- Wrap them as providers and export only needed tokens.
- Avoid rewriting storage code during initial migration.

### Pattern C: Introduce Only Module Boundaries First

- Start with `ModuleDef`, `Imports`, `Exports` discipline.
- Delay controller migration if HTTP stack migration is not needed yet.

## Migration Plan Template

Use this checklist for each module migration:

- Define module name and exported tokens.
- Add provider factories with explicit error handling.
- Add controller registration.
- Add tests for provider resolution and route behavior.
- Roll out behind feature flags if needed.

## Rollback and Exit Path

If a migration step underperforms:

- Revert only the migrated module integration commit.
- Keep unaffected modules/routes running.
- Reintroduce old handler wiring for that route group.

You can also exit modkit entirely by:

- replacing `kernel.Bootstrap(...)` composition with explicit manual wiring,
- keeping provider/controller implementations as regular Go structs.

## What Not to Do

- Do not migrate every route and module at once.
- Do not couple migration with large DB/schema redesigns.
- Do not bypass module exports with cross-module direct state access.

## Verify Each Increment

After each migration slice:

```bash
make test
```

For full repository gates:

```bash
make fmt && make lint && make vuln && make test && make test-coverage
make cli-smoke-build && make cli-smoke-scaffold
```
