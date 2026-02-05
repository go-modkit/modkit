# Implementation Plan - Refactor Http Adapter

## Phase 1: Analysis & API Design
- [x] Task: Analyze current `modkit/http` implementation and identify extension points. 80aea26
- [~] Task: Design the new Middleware API and Server Options pattern.
    - [ ] Create a design document or prototype snippet in `docs/specs/rfc_http_refactor.md`.
- [ ] Task: Conductor - User Manual Verification 'Analysis & API Design' (Protocol in workflow.md)

## Phase 2: Core Refactoring (TDD)
- [ ] Task: Refactor `Server` to use Functional Options.
    - [ ] Create `server_options_test.go` to test configuration injection.
    - [ ] Implement `NewServer(opts ...Option)`.
- [ ] Task: Implement Global Middleware support.
    - [ ] Create `middleware_test.go` verifying execution order.
    - [ ] Implement middleware chaining in the request pipeline.
- [ ] Task: Refactor `RegisterRoutes` for enhanced flexibility.
    - [ ] Update tests to reflect new signature or usage pattern.
    - [ ] Implement changes in `router.go`.
- [ ] Task: Conductor - User Manual Verification 'Core Refactoring' (Protocol in workflow.md)

## Phase 3: Integration & Migration
- [ ] Task: Update `examples/hello-simple` to use the new `http` package API.
- [ ] Task: Update `examples/hello-mysql` to use the new `http` package API.
- [ ] Task: Update `docs/guides/controllers.md` and `docs/guides/middleware.md`.
- [ ] Task: Conductor - User Manual Verification 'Integration & Migration' (Protocol in workflow.md)
