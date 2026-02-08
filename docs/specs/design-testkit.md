# Design Spec: modkit TestKit (P0)

**Status:** Ready for implementation
**Date:** 2026-02-08
**Author:** Sisyphus (AI Agent)
**Related PRD:** `docs/specs/prd-modkit-core.md` (Phase 2 `TestKit`)

## 1. Overview

This spec defines a first-class `modkit/testkit` package to make module-level tests faster to write, easier to read, and safer to maintain.

The target outcome is a test harness that keeps modkit's core constraints (explicit wiring, visibility enforcement, deterministic bootstrap, no reflection magic) while reducing boilerplate for:

- Bootstrapping modules in tests
- Overriding providers with mocks/fakes
- Resolving typed dependencies and controllers
- Ensuring cleanup is always executed

## 2. Motivation and Current Pain

The repository already uses strong testing patterns, but repeated setup boilerplate indicates missing framework support.

### 2.1 Repeated Inline Test Module Builders

- Kernel tests maintain local helper module structs/functions to build small graphs (`mod(...)`) in `modkit/kernel/mod_helper_test.go`.
- Example/config tests define local `rootModule` types just to bootstrap imports in `examples/hello-mysql/internal/modules/config/module_test.go`.

Implication: developers repeatedly rebuild the same scaffolding instead of focusing on behavior under test.

### 2.2 Repeated Bootstrap + Resolve Ceremony

Many tests follow the same sequence: construct module graph, `kernel.Bootstrap`, then `module.Get[T]` and error assertions.

Implication: high ceremony and inconsistent helper style across packages.

### 2.3 Mock/Override Flows Are Manual and Inconsistent

Testing guide recommends replacing dependencies via test-specific modules (`docs/guides/testing.md`), but there is no canonical mechanism for token-level override.

Implication: mock injection for deep graphs is ad-hoc and can be expensive to maintain.

### 2.4 Cleanup Discipline Is Not Centralized

modkit already tracks cleanup hooks and `io.Closer` lifecycle (`modkit/kernel/bootstrap.go`, `modkit/kernel/container.go`), but tests must remember to call cleanup explicitly when needed.

Implication: resource cleanup behavior is available but not ergonomically enforced at harness level.

## 3. Goals

1. Provide a canonical test harness API for modkit module tests.
2. Support provider override/mocking by token with explicit behavior.
3. Preserve modkit visibility and graph validation semantics in tests.
4. Auto-manage cleanup through `t.Cleanup` by default.
5. Keep API minimal and Go-idiomatic (no hidden reflection or global state).

## 4. Non-Goals

1. No replacement of Go's `testing` package.
2. No mocking framework inside modkit (testify/mock, gomock remain optional user choices).
3. No runtime mutation of provider instances after graph bootstrap.
4. No bypass of visibility/export rules by default.
5. No CLI generation features in this phase.

## 5. Design Principles

1. **Explicit over implicit:** overrides are declared with concrete tokens.
2. **Deterministic behavior:** identical inputs produce identical test graph behavior.
3. **Boundary safety:** maintain import/export visibility constraints.
4. **Small surface area:** optimize for common 80% use cases first.
5. **Typed ergonomics:** provide typed helpers for retrieval/assertion.

## 6. Proposed API (v1 Scope)

### 6.1 Package

New package:

`modkit/testkit`

### 6.2 Core Types

```go
package testkit

type TB interface {
    Helper()
    Fatalf(format string, args ...any)
    Cleanup(func())
}

type Harness struct {
    // unexported: app, tb, config
}

type Option interface {
    apply(*config)
}

type Override struct {
    Token module.Token
    Build func(module.Resolver) (any, error)
    Cleanup func(context.Context) error
}
```

### 6.3 Constructors and Helpers

```go
func New(tb TB, root module.Module, opts ...Option) *Harness
func NewE(tb TB, root module.Module, opts ...Option) (*Harness, error)

func WithOverrides(overrides ...Override) Option
func OverrideValue(token module.Token, value any) Override
func OverrideBuild(token module.Token, build func(module.Resolver) (any, error)) Override

func WithoutAutoClose() Option

func (h *Harness) App() *kernel.App
func (h *Harness) Close() error
func (h *Harness) CloseContext(ctx context.Context) error

func Get[T any](tb TB, h *Harness, token module.Token) T
func GetE[T any](h *Harness, token module.Token) (T, error)
func Controller[T any](tb TB, h *Harness, moduleName, controllerName string) T
func ControllerE[T any](h *Harness, moduleName, controllerName string) (T, error)
```

Behavior:

- `NewE` is the error-returning constructor.
- `New` is a convenience wrapper over `NewE` that fails test on constructor error.
- `New`/`NewE` register `h.Close()` via `tb.Cleanup` unless `WithoutAutoClose` is set.
- `Close` executes provider cleanup hooks first (LIFO), then `io.Closer` shutdown (`App.Close`) and returns aggregated errors.
- `Close`/`CloseContext` are idempotent and safe when called manually before `tb.Cleanup` fires.
- `GetE` and `ControllerE` return typed errors.
- `Get[T]` and `Controller[T]` are convenience wrappers that fail test with clear context on mismatch.

## 7. Override Model

### 7.1 Supported Override Forms

1. **Value override** (`OverrideValue`) for static fakes/mocks.
2. **Build override** (`OverrideBuild`) for dynamic dependencies.

### 7.2 Resolution Rules

1. Overrides are keyed by token.
2. Duplicate override tokens are rejected with `DuplicateOverrideTokenError`.
3. Overriding a token not present in the provider graph is rejected with `OverrideTokenNotFoundError`.
4. Override application is deterministic: validate all overrides first, then apply in input order.
5. Overridden providers still resolve through modkit container semantics (lazy singleton).
6. Override replaces both provider `Build` and provider `Cleanup`; nil override cleanup means no cleanup hook for the overridden provider.

### 7.3 Owner-Module and Resolver Semantics

1. An override inherits the original provider owner module scope.
2. During override `Build`, resolver visibility is enforced exactly as for the original provider owner module.
3. Overrides cannot change token ownership or make a token visible to modules that could not previously resolve it.

### 7.4 Visibility Rules

By default, overrides do **not** bypass visibility. If module A cannot see token X in production graph, tests for A should keep that behavior.

Override admission requires root-scope visibility: if `visibility[graph.Root][token]` is false, validation fails with `OverrideTokenNotVisibleFromRootError`.

Rationale: this preserves architecture guarantees and avoids false confidence.

## 8. Kernel Integration Strategy

To avoid brittle graph rewriting in `testkit`, introduce kernel-level bootstrap options that are explicit and reusable.

### 8.1 New Kernel API

```go
// Existing API remains unchanged.
func Bootstrap(root module.Module) (*App, error)

// New API for advanced/test scenarios.
func BootstrapWithOptions(root module.Module, opts ...BootstrapOption) (*App, error)

type BootstrapOption interface {
    apply(*bootstrapConfig)
}

func WithProviderOverrides(overrides ...ProviderOverride) BootstrapOption

type ProviderOverride struct {
    Token module.Token
    Build func(module.Resolver) (any, error)
    Cleanup func(context.Context) error
}
```

`testkit.New` uses `BootstrapWithOptions` under the hood.

Why this approach:

- Keeps override logic near graph/container creation.
- Avoids complex recursive module-definition rewriting.
- Keeps production API explicit and backward-compatible.

### 8.2 Option Application Contract

1. `Bootstrap(root)` is behavior-identical to `BootstrapWithOptions(root)` when no options are supplied.
2. `BootstrapWithOptions` flow is deterministic:
   - Build graph and validate graph invariants.
   - Build visibility map.
   - Validate all options against graph/visibility.
   - Apply option mutations to provider registry.
   - Build container and instantiate controllers.
3. Invalid options fail fast before controller construction.
4. `WithProviderOverrides` is test-focused but remains explicit and side-effect free.
5. Option collision rule is strict: if multiple options target the same token mutation in one bootstrap call, validation fails with `BootstrapOptionConflictError`.
6. v1 scope fence: only `WithProviderOverrides` is supported as a mutation option.

## 9. Errors

Add typed errors in `modkit/kernel` for bootstrap-option validation:

- `OverrideTokenNotFoundError{Token}`
- `OverrideTokenNotVisibleFromRootError{Root, Token}`
- `DuplicateOverrideTokenError{Token}`
- `BootstrapOptionConflictError{Token, Options}`

Add typed errors in `modkit/testkit` for harness/runtime helpers:

- `ControllerNotFoundError{Module, Name}`
- `TypeAssertionError{Target, Actual, Context}`
- `HarnessCloseError{HookErr error, CloseErr error}`

All errors should wrap root cause where possible. `HarnessCloseError` implements `Unwrap() []error` so `errors.Is/As` can match both hook and closer failures.

## 10. Testing Strategy for TestKit

### 10.1 Unit Tests

1. Harness bootstraps valid roots.
2. Harness fails tests on bootstrap errors.
3. `OverrideValue` and `OverrideBuild` replace expected provider behavior.
4. Missing-token override returns `OverrideTokenNotFoundError`.
5. Duplicate override token returns `DuplicateOverrideTokenError`.
6. `GetE[T]` and `ControllerE[T]` return typed assertion/not-found errors.
7. `Get[T]` and `Controller[T]` wrappers fail clearly on type mismatch.
8. `Close` and `CloseContext` are idempotent.
9. `Close` aggregates cleanup-hook and closer failures deterministically.

### 10.2 Integration Tests

1. Override in multi-module graph with re-export path.
2. Visibility remains enforced under override.
3. Override build executes with original owner-module visibility constraints.
4. Hidden-token override attempts fail with `OverrideTokenNotVisibleFromRootError` and do not alter visibility map.
5. Cleanup hooks and closers still execute in expected order.
6. `WithoutAutoClose` disables cleanup registration.
7. Override cleanup replacement behavior is verified (original cleanup replaced by override cleanup).

### 10.3 Parity and Compatibility Tests

1. `Bootstrap(root)` and `BootstrapWithOptions(root)` produce equivalent app behavior with no options.
2. Existing kernel tests continue passing without edits beyond additive coverage.
3. Override-free testkit harness behaves equivalently to manual bootstrap patterns.

### 10.4 Race and Stability

Run race-enabled tests and ensure no data races in harness lifecycle.

## 11. Documentation Deliverables

1. Update `docs/guides/testing.md` with a new "Using TestKit" section.
2. Add explicit guidance: when to use TestKit overrides vs test-specific modules.
3. Add API references to `docs/reference/api.md`.
4. Add one focused example in `examples/hello-simple` tests.
5. Add one realistic override example in `examples/hello-mysql` tests.

## 12. Rollout Plan

### Story P0.1 - Kernel Bootstrap Options (Foundation)

- Add `BootstrapWithOptions` and override plumbing.
- Keep `Bootstrap` backward-compatible and parity-tested against the no-options path.

### Story P0.2 - TestKit Harness Core

- Add `testkit.New`, `Harness`, auto-close behavior.
- Add typed retrieval/controller helpers.

### Story P0.3 - Override API and Validation

- Add override constructors and validation errors.
- Add strict duplicate and unknown-token validation.
- Add multi-module override coverage with owner-module visibility assertions.

### Story P0.4 - Docs + Examples

- Update guides and add runnable examples.

## 13. Acceptance Criteria

This P0 is complete when all are true:

1. `modkit/testkit` package exists with harness + typed helpers.
2. Token-level provider overrides are supported and tested.
3. Visibility and graph validation guarantees are preserved in tests.
4. Overrides enforce strict validation (duplicate tokens and unknown tokens fail fast).
5. `Bootstrap` and `BootstrapWithOptions` have parity coverage for no-options behavior.
6. Cleanup is auto-registered via `t.Cleanup` by default and manual close remains idempotent.
7. Docs include a canonical TestKit workflow and override-vs-test-module guidance.
   - Required evidence: one section in `docs/guides/testing.md` with two explicit decision examples:
     - "Use test module replacement when changing module wiring semantics"
     - "Use TestKit override when isolating dependency behavior without changing graph shape"
8. CI passes with `make fmt && make lint && make vuln && make test && make test-coverage`.

## 14. Risks and Mitigations

1. **Risk:** Overrides accidentally hide visibility regressions.
   - **Mitigation:** keep visibility enforcement enabled by default; explicit opt-out not in P0.

2. **Risk:** Kernel option surface grows too much.
   - **Mitigation:** keep only provider override option in v1.

3. **Risk:** Test harness becomes de-facto production API.
   - **Mitigation:** package under `testkit`, document test-only intent.

4. **Risk:** Override semantics ambiguous for duplicate tokens.
   - **Mitigation:** strict duplicate rejection + deterministic validation/apply order.

5. **Risk:** Option path changes bootstrap behavior unexpectedly.
   - **Mitigation:** enforce no-options parity tests for `BootstrapWithOptions`.

## 15. P0 Decisions (Locked)

1. TestKit does not expose module-scoped resolvers in P0; only `Get[T]`/`GetE[T]` and `Controller[T]`/`ControllerE[T]` are included.
2. No visibility-bypass option exists in P0.

## 16. External References

1. Fx testing helpers (`fxtest`) for harness ergonomics and lifecycle helpers:
   - https://pkg.go.dev/go.uber.org/fx/fxtest
2. Fx docs section "Testing Fx Applications":
   - https://pkg.go.dev/go.uber.org/fx
3. Wire community discussion on mocking pain points:
   - https://github.com/google/wire/issues/43
