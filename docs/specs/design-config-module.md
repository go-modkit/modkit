# Design Spec: modkit Core Config Module

**Status:** Draft
**Date:** 2026-02-07
**Author:** Sisyphus (AI Agent)
**Related PRD:** `docs/specs/prd-modkit-core.md` (Phase 2 `Config Module`)

## 1. Overview

This document specifies a reusable Config Module for core `modkit`.

Current configuration loading exists in example apps (`examples/hello-mysql/internal/config` and `examples/hello-mysql/internal/platform/config`) but not as a core framework capability. This spec defines a standard way to load typed config from environment variables into the DI container, aligned with modkit principles: explicit wiring, deterministic bootstrap, and no reflection magic.

## 2. Goals

- Provide a first-class, reusable configuration pattern for modkit apps.
- Support typed config values resolved via `module.Get[T]`.
- Make defaults, required keys, and parse errors explicit and deterministic.
- Preserve module visibility/export rules for config tokens.
- Keep implementation lightweight and standard-library-first.

## 3. Non-Goals

- Full-featured external config framework replacement (Viper/Koanf class scope).
- Automatic binding from struct tags using reflection.
- Dynamic hot-reload/watch mode in v1.
- Secret manager integrations (Vault/SSM/etc.) in v1.

## 4. Design Principles

- **Explicit schema:** app defines exactly which keys are read.
- **Deterministic bootstrap:** config is loaded at provider build and fails early on invalid required values.
- **No hidden globals:** no package-level mutable config singleton.
- **Composable modules:** config providers can be private or exported via normal module metadata.
- **Clear errors:** include env key, expected type, and module/token context.

## 5. Proposed Package and API Shape

## 5.1. Package Location

- New core package: `modkit/config`

## 5.2. Core Types (Conceptual)

```go
type KeySpec struct {
    Key         string
    Required    bool
    Default     *string
    Description string
}

type Source interface {
    Get(key string) (string, bool)
}

type LoadError struct {
    Key    string
    Reason string
}
```

Notes:

- `Source` defaults to environment source in v1 (`os.LookupEnv`).
- `KeySpec` captures required/default semantics explicitly.
- Parsing helpers return typed values and rich errors.

## 5.3. Module Construction Pattern

The config package should provide helper constructors (names illustrative):

- `config.NewModule(opts ...Option) module.Module`
- `config.WithToken(token module.Token)`
- `config.WithSchema(schema any)` or explicit field registration options
- `config.WithSource(src Source)`

No reflection auto-binding is required for v1. If schema struct support is added, it must remain explicit and deterministic.

## 6. Token and Visibility Model

## 6.1. Token Convention

- Base token prefix: `config.`
- Recommended exported tokens:
  - `config.raw` (optional map/string view)
  - `config.app` (typed app config struct)

## 6.2. Visibility

- Config module can keep raw internals private.
- Apps should export only typed config tokens needed by importers.
- Standard modkit visibility rules apply with no exceptions.

## 7. Loading and Validation Semantics

## 7.1. String Resolution

For each configured key:

1. Read from source (`LookupEnv`).
2. Trim spaces.
3. If empty and default exists, use default.
4. If required and still empty, return typed load error.

## 7.2. Type Parsing Helpers

Provide explicit helpers for common types:

- `String`
- `Int`
- `Float64`
- `Bool`
- `Duration`
- `CSV []string`

Each helper must include key name in parse error context.

## 7.3. Error Model

Expected categories:

- Missing required key
- Invalid parse for key/type
- Invalid schema/spec definition (developer error)

Error messages should be structured for test assertions and operator troubleshooting.

## 8. Security Considerations

- Never log secret values in errors.
- Allow marking keys as sensitive to force redaction in diagnostics.
- Optional debug dump (if any) must redact sensitive keys by default.
- Docs should recommend secrets from environment/secret store injection, not committed files.

## 9. Integration Pattern

Example module wiring shape:

```go
type AppConfig struct {
    HTTPAddr  string
    JWTSecret string
}

const TokenAppConfig module.Token = "config.app"

func (m *AppModule) Definition() module.ModuleDef {
    cfgModule := config.NewModule(
        // explicit schema/options
    )

    return module.ModuleDef{
        Name:    "app",
        Imports: []module.Module{cfgModule},
        Providers: []module.ProviderDef{{
            Token: "app.service",
            Build: func(r module.Resolver) (any, error) {
                cfg, err := module.Get[AppConfig](r, TokenAppConfig)
                if err != nil {
                    return nil, err
                }
                return NewService(cfg), nil
            },
        }},
    }
}
```

## 10. Testing Strategy

## 10.1. Unit Tests

- Key resolution (present, empty, missing, defaulted).
- Typed parsing for supported primitives.
- Redaction behavior for sensitive keys.
- Error typing and error message content.

## 10.2. Integration Tests

- Bootstrapping app with config module succeeds with valid env.
- Bootstrapping fails fast for missing required keys.
- Visibility checks when config token is not exported.

## 10.3. Compatibility Tests

- Ensure example-app migration preserves behavior for existing env vars.

## 11. Adoption and Migration Plan

1. Add core `modkit/config` package with basic env source and typed parsing helpers.
2. Introduce guide docs showing recommended config module composition.
3. Migrate `examples/hello-mysql` incrementally to consume core config module patterns.
4. Keep example-specific wrappers only where app-specific semantics differ.

## 12. Acceptance Criteria

This PRD item is complete when all are true:

1. A core config package exists in `modkit/` and is documented.
2. Apps can load typed config via module providers and `module.Get[T]`.
3. Missing/invalid required config fails at bootstrap with descriptive errors.
4. Sensitive keys are redacted in any config diagnostics.
5. At least one example app demonstrates the core pattern.

## 13. Open Questions

1. Should v1 expose a map-like raw config token, or typed-only tokens?
2. Should duration parsing support strict format only (Go duration) or aliases?
3. Should `.env` file support be included in core or kept out-of-scope?
4. Should config module own validation rules, or only load and parse while feature modules validate domain constraints?

## 14. Future Enhancements (Not in v1)

- Multi-source layering (env + file + flags).
- Live reload callbacks.
- Secret manager source adapters.
- Schema export for docs generation.
