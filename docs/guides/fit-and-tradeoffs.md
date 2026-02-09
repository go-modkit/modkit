# Fit and Trade-offs

Use this page to decide whether modkit is a good match for your service.

## Good Fit

Choose modkit when you want:

- explicit module boundaries (`imports`/`exports`),
- explicit DI tokens without reflection,
- deterministic bootstrap behavior with clear errors,
- a structured backend architecture for multi-feature teams.

## Probably Not a Fit

Consider alternatives when you need:

- automatic lifecycle orchestration across many service components,
- request-scoped or transient dependency scopes,
- compile-time DI generation over runtime graph building,
- minimal abstraction for a very small one-file service.

## Trade-offs

- **Pros**: explicit wiring, predictable behavior, easy boundary review.
- **Cons**: more upfront structure, token management discipline, less automation magic.

## Non-goals

modkit intentionally does not provide:

- decorator-based reflection wiring,
- ORM/data-access opinionation,
- hidden auto-discovery of providers/controllers.

## Decision Checklist

Use modkit if most answers are "yes":

- Do we have multiple feature teams or bounded contexts?
- Do we need import/export visibility enforcement between features?
- Do we value explicitness over framework magic?

If most answers are "no", manual DI or a lighter container may be better.
