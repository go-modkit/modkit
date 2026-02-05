# Epic: Examples & Documentation Enhancement

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
- [ ] JWT middleware extracts and validates tokens
- [ ] Protected routes require valid authentication
- [ ] User info is available in request context
- [ ] Mix of public and authenticated endpoints works
- [ ] Token generation endpoint exists
- [ ] Tests cover auth success and failure cases

**Documentation Reference**: `docs/guides/authentication.md`

#### Tasks
- [ ] Create `internal/modules/auth/` module structure
- [ ] Implement JWT middleware with token validation
- [ ] Add typed context helpers for user extraction
- [ ] Create `/auth/login` endpoint for token generation
- [ ] Protect existing `/users` CRUD routes (except GET list)
- [ ] Add auth module tests (unit + integration)
- [ ] Update README with auth usage examples

---

### Story 2: Request Validation

**Description**: Add structured request validation to `hello-mysql` demonstrating validation patterns with proper error responses.

**Acceptance Criteria**:
- [ ] Request body validation on POST/PUT endpoints
- [ ] Field-level validation error responses
- [ ] RFC 7807 Problem Details for validation errors
- [ ] Query parameter validation examples
- [ ] Path parameter validation examples

**Documentation Reference**: `docs/guides/validation.md`

#### Tasks
- [ ] Add validation helper package in `internal/validation/`
- [ ] Implement field-level validation for CreateUserRequest
- [ ] Implement field-level validation for UpdateUserRequest
- [ ] Add validation error response type (RFC 7807 extension)
- [ ] Add query parameter validation example (list pagination)
- [ ] Add validation tests
- [ ] Document validation patterns in example README

---

### Story 3: Advanced Middleware Patterns

**Description**: Add examples of common middleware patterns: CORS, rate limiting, and custom middleware as providers.

**Acceptance Criteria**:
- [ ] CORS middleware configured and working
- [ ] Rate limiting middleware with configurable limits
- [ ] Custom middleware registered as provider
- [ ] Route group with scoped middleware
- [ ] Middleware ordering is demonstrated

**Documentation Reference**: `docs/guides/middleware.md`

#### Tasks
- [ ] Add CORS middleware configuration
- [ ] Implement rate limiting middleware using `golang.org/x/time/rate`
- [ ] Create timing/metrics middleware example
- [ ] Add route group example (`/api/v1/` prefix)
- [ ] Register middleware as providers for dependency injection
- [ ] Add middleware tests
- [ ] Document middleware patterns in example README

---

### Story 4: Lifecycle & Cleanup Patterns

**Description**: Add examples demonstrating proper resource cleanup and graceful shutdown patterns.

**Acceptance Criteria**:
- [ ] Database connection cleanup on shutdown
- [ ] Graceful shutdown with in-flight request handling
- [ ] Context-based cancellation patterns
- [ ] Resource cleanup order is correct

**Documentation Reference**: `docs/guides/lifecycle.md`

#### Tasks
- [ ] Add cleanup interface and implementation to database module
- [ ] Implement graceful shutdown hook in main
- [ ] Add context cancellation example for long-running operations
- [ ] Document cleanup order (LIFO)
- [ ] Add lifecycle tests
- [ ] Update README with lifecycle patterns

---

### Story 5: Route Groups & API Versioning

**Description**: Demonstrate route grouping patterns for API versioning and scoped middleware.

**Acceptance Criteria**:
- [ ] API versioning with `/api/v1/` prefix
- [ ] Route groups with shared middleware
- [ ] Nested route groups example
- [ ] Controller uses groups effectively

**Documentation Reference**: `docs/guides/controllers.md`

#### Tasks
- [ ] Refactor routes to use `/api/v1/` prefix
- [ ] Add route group with auth middleware
- [ ] Add nested group example (e.g., `/api/v1/admin/`)
- [ ] Update controller to demonstrate grouping
- [ ] Add routing tests for groups
- [ ] Document routing patterns

---

### Story 6: Test Coverage Improvements

**Description**: Improve test coverage for core modkit library and ensure examples have comprehensive tests.

**Acceptance Criteria**:
- [ ] `buildVisibility()` has direct unit tests
- [ ] Provider/Controller build errors are tested
- [ ] HTTP middleware edge cases are tested
- [ ] Example tests cover new features
- [ ] CI coverage meets threshold

**Documentation Reference**: `docs/guides/testing.md`

#### Tasks
- [ ] Add unit tests for `buildVisibility()` in kernel
- [ ] Add tests for ProviderBuildError scenarios
- [ ] Add tests for ControllerBuildError scenarios  
- [ ] Add HTTP middleware error recovery tests
- [ ] Add router edge case tests (conflicts, invalid methods)
- [ ] Ensure all new example features have tests
- [ ] Update testing guide with new patterns

---

### Story 7: Documentation Synchronization

**Description**: Ensure all documentation guides reference working example code and patterns are consistent.

**Acceptance Criteria**:
- [ ] All guides reference example code where applicable
- [ ] Code snippets in docs match actual example code
- [ ] No orphaned documentation (patterns without examples)
- [ ] README links are correct

#### Tasks
- [ ] Audit `docs/guides/authentication.md` vs example code
- [ ] Audit `docs/guides/validation.md` vs example code
- [ ] Audit `docs/guides/middleware.md` vs example code
- [ ] Audit `docs/guides/lifecycle.md` vs example code
- [ ] Update guides with correct code references
- [ ] Add "See example" links to each guide
- [ ] Update main README with feature matrix

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
