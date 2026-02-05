# Auth Example Design

**Goal:** Add a runnable, example-focused JWT authentication module to hello-mysql with a login endpoint, middleware validation, and typed context helpers.

**Architecture:** A dedicated `auth` module provides a login handler and a JWT middleware provider. Configuration is explicit via example config/env. The middleware validates tokens and stores user info in a typed context helper, which handlers can read. User write routes (`POST /users`, `PUT /users/{id}`, `DELETE /users/{id}`) are protected by the auth middleware, while the list route (`GET /users`) remains public.

**Tech Stack:** Go, chi router via modkit http adapter, standard library + minimal JWT dependency.

---

## Section 1 — Architecture Summary

We add `examples/hello-mysql/internal/modules/auth` with a deterministic module definition and provider scaffolding. The module exports two primary providers: a JWT validation middleware and a login handler/controller. Configuration is explicit and local to the example (`JWT_SECRET`, `JWT_ISSUER`, `JWT_TTL`, `AUTH_USERNAME`, `AUTH_PASSWORD`). The login endpoint verifies demo credentials (no DB, no hashing) and returns a signed HS256 JWT with a minimal subject/email claim. The middleware validates the `Authorization: Bearer <token>` header, verifies signature + expiry, and stores authenticated user info in the request context via typed helpers. Downstream handlers access user info using those helpers only; no global state.

## Section 2 — Components and Data Flow

**Config:** Extend `examples/hello-mysql/internal/platform/config` with JWT + demo auth fields. `Load()` pulls from env with defaults (e.g., username `demo`, password `demo`, issuer `hello-mysql`, TTL `1h`).

**Auth Module:**
- `module.go`: registers module name, exports provider tokens.
- `providers.go`: builds middleware and login handler using config.
- `config.go`: holds auth config struct sourced from platform config.

**JWT Middleware:**
- Extracts bearer token, returns 401 on missing/invalid tokens.
- Verifies signature and expiry using HS256.
- On success, stores `AuthUser{ID, Email}` in context.

**Login Handler:**
- `POST /auth/login` expects JSON with username/password.
- Validates against demo config values.
- Returns `{ "token": "<jwt>" }` on success.

**Typed Context Helpers:**
- `WithUser(ctx, user)` and `UserFromContext(ctx)` in `context.go`.
- Used by handlers and tests to show how to access authenticated user.

## Section 3 — Error Handling, Tests, and Docs

**Errors:** Use existing `httpapi.WriteProblem` for auth errors with status `401`. Validation errors for login payload are `400`. Internal issues return `500` with explicit context wrapping.

**Tests:**
- Unit tests for middleware and context helpers (valid/invalid token, missing header).
- Integration tests for `/auth/login` and protected routes (valid/invalid creds).
- Use table-driven tests for token validation cases.

**Documentation:** Update `examples/hello-mysql/README.md` with login example, token usage, and which `/users` routes require auth. Keep examples aligned with code paths.

---

**Defaults (chosen):**
- `AUTH_USERNAME=demo`
- `AUTH_PASSWORD=demo`
- `JWT_ISSUER=hello-mysql`
- `JWT_TTL=1h`
- `JWT_SECRET=dev-secret-change-me`
