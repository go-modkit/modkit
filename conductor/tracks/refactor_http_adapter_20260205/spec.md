# Specification: Refactor Http Adapter

## Context
The current `modkit/http` adapter provides a basic wrapper around `go-chi`, but it lacks flexibility in middleware application and limits how users can customize route registration. As the framework grows, users need a more robust way to intercept requests, handle errors globally, and configure the server.

## Goals
1.  **Extensible Middleware:** Allow users to apply middleware globally (at the `App` level) and per-controller/route.
2.  **Enhanced Route Registration:** Refactor `RegisterRoutes` to support cleaner, explicit routing definitions that align with the `modkit` philosophy (no magic).
3.  **Standardized Error Handling:** Improve default error responses to align with RFC 7807 (Problem Details) and allow users to override the error encoder.
4.  **Zero-Reflection:** Maintain the core promise of the framework by avoiding reflection in the hot path.

## Proposed Design

### Middleware
We will introduce a standard Middleware type:
```go
type Middleware func(http.Handler) http.Handler
```
And allow it to be chained explicitly.

### Server Options
The `Server` struct/constructor should accept Functional Options to configure timeouts, logging, and global middleware.

### Route Registration
We will likely keep the explicit `RegisterRoutes` but allow it to accept a configuration object or options to inject route-specific middleware.

## Non-Goals
-   Replacing `go-chi`. We will continue to build *on top* of it.
-   Adding auto-discovery of controllers.

## Success Criteria
-   All existing tests pass.
-   New tests cover middleware execution order.
-   The `hello-simple` and `hello-mysql` examples are updated and functional.
