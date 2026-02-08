# Epic: Examples & Documentation Enhancement

**Status:** Implemented (v1)
**Last Reviewed:** 2026-02-08

## Overview

Enhance modkit examples to demonstrate all documented patterns and close the gap between documentation and working code. The `hello-mysql` example will be extended with authentication, validation, middleware patterns, and lifecycle management.

**Goal**: Users should be able to find working example code for every pattern described in the documentation.

**Success Criteria**:
- All documented patterns have corresponding example code
- Examples are tested and pass CI
- Documentation references point to actual runnable code

---

## Stories

### Story 1: Authentication Module

**Description**: Add a complete authentication module to `hello-mysql` demonstrating JWT-based auth, protected routes, and user context extraction.

**Acceptance Criteria**:
- [x] JWT middleware extracts and validates tokens
- [x] Protected routes require valid authentication
- [x] User info is available in request context
- [x] Mix of public and authenticated endpoints works
- [x] Token generation endpoint exists
- [x] Tests cover auth success and failure cases

**Documentation Reference**: `docs/guides/authentication.md`

#### Tasks
- [x] Create `internal/modules/auth/` module structure
- [x] Implement JWT middleware with token validation
- [x] Add typed context helpers for user extraction
- [x] Create `/auth/login` endpoint for token generation
- [x] Protect existing `/users` CRUD routes (except GET list)
- [x] Add auth module tests (unit + integration)
- [x] Update README with auth usage examples

---

### Story 2: Request Validation

**Description**: Add structured request validation to `hello-mysql` demonstrating validation patterns with proper error responses.

**Acceptance Criteria**:
- [x] Request body validation on POST/PUT endpoints
- [x] Field-level validation error responses
- [x] RFC 7807 Problem Details for validation errors
- [x] Query parameter validation examples
- [x] Path parameter validation examples

**Documentation Reference**: `docs/guides/validation.md`

#### Tasks
- [x] Add validation helper package in `internal/validation/`
- [x] Implement field-level validation for CreateUserRequest
- [x] Implement field-level validation for UpdateUserRequest
- [x] Add validation error response type (RFC 7807 extension)
- [x] Add query parameter validation example (list pagination)
- [x] Add validation tests
- [x] Document validation patterns in example README

---

### Story 3: Advanced Middleware Patterns

**Description**: Add examples of common middleware patterns: CORS, rate limiting, and custom middleware as providers.

**Acceptance Criteria**:
- [x] CORS middleware configured and working
- [x] Rate limiting middleware with configurable limits
- [x] Custom middleware registered as provider
- [x] Route group with scoped middleware
- [x] Middleware ordering is demonstrated

**Documentation Reference**: `docs/guides/middleware.md`

#### Tasks
- [x] Add CORS middleware configuration
- [x] Implement rate limiting middleware using `golang.org/x/time/rate`
- [x] Create timing/metrics middleware example
- [x] Add route group example (`/api/v1/` prefix)
- [x] Register middleware as providers for dependency injection
- [x] Add middleware tests
- [x] Document middleware patterns in example README

---

### Story 4: Lifecycle & Cleanup Patterns

**Description**: Add examples demonstrating proper resource cleanup and graceful shutdown patterns.

**Acceptance Criteria**:
- [x] Database connection cleanup on shutdown
- [x] Graceful shutdown with in-flight request handling
- [x] Context-based cancellation patterns
- [x] Resource cleanup order is correct

**Documentation Reference**: `docs/guides/lifecycle.md`

#### Tasks
- [x] Add cleanup interface and implementation to database module
- [x] Implement graceful shutdown hook in main
- [x] Add context cancellation example for long-running operations
- [x] Document cleanup order (LIFO)
- [x] Add lifecycle tests
- [x] Update README with lifecycle patterns

---

### Story 5: Route Groups & API Versioning

**Description**: Demonstrate route grouping patterns for API versioning and scoped middleware.

**Acceptance Criteria**:
- [x] API versioning with `/api/v1/` prefix
- [x] Route groups with shared middleware
- [ ] Nested route groups example
- [ ] Controller uses groups effectively

**Documentation Reference**: `docs/guides/controllers.md`

#### Tasks
- [x] Refactor routes to use `/api/v1/` prefix
- [ ] Add route group with auth middleware
- [ ] Add nested group example (e.g., `/api/v1/admin/`)
- [ ] Update controller to demonstrate grouping
- [x] Add routing tests for groups
- [x] Document routing patterns

---

### Story 6: Test Coverage Improvements

**Description**: Improve test coverage for core modkit library and ensure examples have comprehensive tests.

**Acceptance Criteria**:
- [x] `buildVisibility()` has direct unit tests
- [x] Provider/Controller build errors are tested
- [x] HTTP middleware edge cases are tested
- [x] Example tests cover new features
- [x] CI coverage meets threshold

**Documentation Reference**: `docs/guides/testing.md`

#### Tasks
- [x] Add unit tests for `buildVisibility()` in kernel
- [x] Add tests for ProviderBuildError scenarios
- [x] Add tests for ControllerBuildError scenarios
- [x] Add HTTP middleware error recovery tests
- [x] Add router edge case tests (conflicts, invalid methods)
- [x] Ensure all new example features have tests
- [x] Update testing guide with new patterns

---

### Story 7: Documentation Synchronization

**Description**: Ensure all documentation guides reference working example code and patterns are consistent.

**Acceptance Criteria**:
- [x] All guides reference example code where applicable
- [x] Code snippets in docs match actual example code
- [x] No orphaned documentation (patterns without examples)
- [x] README links are correct

#### Tasks
- [x] Audit `docs/guides/authentication.md` vs example code
- [x] Audit `docs/guides/validation.md` vs example code
- [x] Audit `docs/guides/middleware.md` vs example code
- [x] Audit `docs/guides/lifecycle.md` vs example code
- [x] Update guides with correct code references
- [x] Add "See example" links to each guide
- [x] Update main README with feature matrix

---

## Priority Order

1. **Story 1: Authentication** - Most requested missing feature
2. **Story 2: Validation** - Critical for real-world usage
3. **Story 3: Middleware** - Completes HTTP patterns
4. **Story 4: Lifecycle** - Important for production use
5. **Story 5: Route Groups** - Improves organization
6. **Story 6: Testing** - Quality assurance
7. **Story 7: Documentation** - Final polish

## Estimated Scope

- **Total Stories**: 7
- **Total Tasks**: ~42
- **Affected Files**: examples/hello-mysql/, modkit/ (tests), docs/guides/

## Dependencies

- Stories 2-5 depend on existing `hello-mysql` structure
- Story 6 depends on Stories 1-5 being complete
- Story 7 depends on all other stories

## Labels for Issues

- `epic` - This epic
- `story` - Each story
- `task` - Each task
- `enhancement` - New features
- `documentation` - Doc updates
- `examples` - Example improvements
- `testing` - Test improvements
