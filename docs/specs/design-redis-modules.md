# Design Spec: Redis Module Support (P1)

**Status:** Ready for implementation
**Date:** 2026-02-09
**Author:** Sisyphus (AI Agent)
**Related PRD:** `docs/specs/prd-modkit-core.md` (adoption follow-up)

## 1. Overview

This spec defines first-class Redis support for modkit as a composable set of modules, not a monolithic adapter.

Primary objectives:

1. provide a stable Redis client provider module,
2. layer optional capability modules (cache/session/rate-limit store) on top,
3. preserve modkit architecture constraints (explicit tokens, deterministic wiring, visibility boundaries).

## 2. Current State and Problem

The repository currently includes middleware examples (including rate limiting) but no Redis provider module contract in core or examples.

Adoption impact:

1. teams cannot see an official pattern for distributed cache/session integration,
2. production-readiness perception suffers without a canonical Redis path,
3. users build ad-hoc Redis modules with inconsistent tokens and lifecycle behavior.

## 3. Goals

1. Define a canonical Redis client module with explicit exports.
2. Keep domain modules independent from Redis client implementation details.
3. Support common OSS use cases first: cache, session storage, and distributed rate-limit store.
4. Provide typed errors and clean shutdown semantics.
5. Keep v1 small and explicit.

## 4. Non-Goals

1. No queue/worker framework in v1.
2. No Redis Cluster/Sentinel orchestration layer in v1.
3. No transparent fallback cache chain logic in core.
4. No hidden key naming magic.

## 5. Design Principles

1. **Client first:** one base module exports Redis client and config contract.
2. **Capability layering:** higher-level modules import base client module and export focused contracts.
3. **Token stability:** token naming is stable and shared through a contract package.
4. **Lifecycle correctness:** ping-on-build (optional) and close-on-cleanup are explicit.
5. **No silent behavior:** timeouts, key prefixes, and TTL policies are explicit config values.

## 6. Proposed Package and Module Shape

### 6.1 Contract Package

Add:

- `modkit/data/redismodule`

Proposed tokens:

```go
const (
    TokenClient module.Token = "redis.client"
    TokenPrefix module.Token = "redis.key_prefix"
)
```

Optional capability tokens:

```go
const (
    TokenCacheStore     module.Token = "redis.cache.store"
    TokenSessionStore   module.Token = "redis.session.store"
    TokenRateLimitStore module.Token = "redis.ratelimit.store"
)
```

### 6.2 Base Redis Module

Add:

- `modkit/data/redis`

Exports:

1. `redismodule.TokenClient`
2. `redismodule.TokenPrefix` (for key namespace hygiene)

`NewModule(opts Options) module.Module` pattern matches existing module constructors.

### 6.3 Capability Modules

Add optional modules (can live under core or examples first, same contract):

1. `modkit/data/rediscache`
2. `modkit/data/redissession`
3. `modkit/data/redisratelimit`

Each imports the base Redis module and exports exactly one focused token.

## 7. Configuration Contract

Required:

- `REDIS_ADDR`

Optional:

- `REDIS_PASSWORD`
- `REDIS_DB`
- `REDIS_DIAL_TIMEOUT`
- `REDIS_READ_TIMEOUT`
- `REDIS_WRITE_TIMEOUT`
- `REDIS_KEY_PREFIX`

Capability modules add their own config keys (for example session TTL, cache default TTL, limiter window).

## 8. API and Resolver Semantics

Base resolution:

```go
client, err := module.Get[*redis.Client](r, redismodule.TokenClient)
prefix, err := module.Get[string](r, redismodule.TokenPrefix)
```

Capability resolution examples:

```go
cacheStore, err := module.Get[CacheStore](r, redismodule.TokenCacheStore)
sessionStore, err := module.Get[SessionStore](r, redismodule.TokenSessionStore)
```

Behavior requirements:

1. singleton client per module instance,
2. cleanup hook closes client,
3. optional startup ping controlled by explicit config,
4. typed errors for connection/config failures.

## 9. Modular Alignment Check and Suggested Improvements

### 9.1 What Already Aligns

1. Existing middleware/docs already model explicit request pipeline and configuration.
2. Existing module conventions (`Options`, `NewModule`, explicit `Definition`) map directly to Redis modules.

### 9.2 Improvements Required

1. **Introduce contract package for Redis tokens** to prevent ad-hoc token names.
2. **Separate base client module from capability modules** to avoid over-exporting and maintain clear boundaries.
3. **Standardize key prefix policy** through exported config token to prevent cross-module key collisions.
4. **Document visibility recommendations** (app imports capability modules, feature modules consume exported contracts only).

No core kernel changes are required for this design.

## 10. Testing Strategy

### 10.1 Unit Tests

1. config parsing and defaults,
2. provider build error typing,
3. cleanup idempotency and error propagation,
4. key prefix normalization behavior.

### 10.2 Integration Tests

1. Redis testcontainers smoke test for base client module,
2. cache/session/rate-limit capability module roundtrips,
3. visibility tests for non-exported internals.

### 10.3 Compatibility Tests

1. Existing middleware tests continue passing unchanged.
2. New Redis-enabled example compiles and passes module tests.

## 11. Documentation Deliverables

1. New guide: `docs/guides/redis-patterns.md`.
2. Update `docs/guides/middleware.md` with Redis-backed rate limiting option.
3. Update `README.md` feature matrix with Redis cache/session/rate-limit examples.
4. Add explicit section for key namespace and TTL policy conventions.

## 12. Rollout Plan

### Story P1.R1 - Base Redis Client Module

1. Add contract tokens and base module.
2. Add example usage in one focused example app path.

### Story P1.R2 - Cache Capability Module

1. Add `rediscache` module with simple get/set/delete contract.
2. Add guide snippet showing service-level usage.

### Story P1.R3 - Session + Rate-Limit Capability Modules

1. Add session and rate-limit store modules.
2. Integrate Redis-backed rate limiter in example app (opt-in path).

### Story P1.R4 - Docs + CI

1. Add docs and README links.
2. Add smoke checks for Redis-enabled example flow.

## 13. Acceptance Criteria

This phase is complete when all are true:

1. A base Redis module exists and exports stable contract tokens.
2. At least one capability module (cache) is shipped and documented.
3. Session/rate-limit capability modules are specified and have runnable example coverage.
4. Lifecycle cleanup and error typing are tested.
5. Docs define clear key-prefix and TTL conventions.

## 14. Risks and Mitigations

1. **Risk:** Redis module becomes an oversized "kitchen sink".
   - **Mitigation:** keep base client module minimal and split capabilities into separate modules.
2. **Risk:** Key collisions across modules.
   - **Mitigation:** mandatory prefix policy token and docs.
3. **Risk:** Runtime outages due to hidden network assumptions.
   - **Mitigation:** explicit timeout config and optional startup ping behavior.

## 15. Future Enhancements (Out of Scope)

1. Redis Cluster/Sentinel support.
2. Stream/consumer-group module.
3. Distributed lock abstraction.
4. OpenTelemetry Redis instrumentation helper module.
