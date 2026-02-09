# Providers

Providers are the building blocks of your application's business logic. They encapsulate services, repositories, and any other dependencies that controllers or other providers need.

## What is a Provider?

A provider is a value registered in a module that can be injected into controllers or other providers. In modkit, providers are:

- **Lazy:** Built on first `Get()` call, not at bootstrap
- **Singletons:** Same instance returned for all subsequent calls
- **Scoped:** Only accessible according to module visibility rules

## Defining Providers

Providers are defined in your module's `Providers` field:

```go
type DatabaseModule struct {
    dsn string
}

func (m *DatabaseModule) Definition() module.ModuleDef {
    return module.ModuleDef{
        Name: "database",
        Providers: []module.ProviderDef{
            {
                Token: "db.connection",
                Build: func(r module.Resolver) (any, error) {
                    return sql.Open("mysql", m.dsn)
                },
            },
        },
        Exports: []module.Token{"db.connection"},
    }
}
```

## ProviderDef Fields

| Field | Type | Description |
|-------|------|-------------|
| `Token` | `module.Token` | Unique identifier for the provider |
| `Build` | `func(Resolver) (any, error)` | Factory function that creates the provider value |

## Tokens

Tokens are string identifiers that uniquely identify a provider within the module graph:

```go
// Define tokens as constants for type safety and reuse
const (
    TokenDB          module.Token = "database.connection"
    TokenUsersRepo   module.Token = "users.repository"
    TokenUsersService module.Token = "users.service"
)
```

**Naming conventions:**
- Use dot notation: `module.component` (e.g., `users.service`)
- Keep tokens lowercase
- Be descriptive but concise

## Provider Lifecycle

```mermaid
stateDiagram-v2
    direction LR
    
    [*] --> Registered: Bootstrap
    Registered --> Building: First Get()
    Building --> Cached: Build success
    Building --> Error: Build fails
    Cached --> Cached: Subsequent Get()
```

1. **Registered:** At bootstrap, the factory function is stored (not called)
2. **Building:** On first `Get()`, the factory is invoked
3. **Cached:** The result is stored and returned for all future `Get()` calls

## Resolving Dependencies

Use the `Resolver` to get other providers:

```go
module.ProviderDef{
    Token: TokenUsersService,
    Build: func(r module.Resolver) (any, error) {
        // Get a dependency
        db, err := module.Get[*sql.DB](r, TokenDB)
        if err != nil {
            return nil, err
        }
        
        // Get another dependency
        logger, err := module.Get[Logger](r, TokenLogger)
        if err != nil {
            return nil, err
        }
        
        return NewUsersService(db, logger), nil
    },
}
```

## Error Handling

The `Build` function returns an error for failed initialization:

```go
Build: func(r module.Resolver) (any, error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }
    
    // Verify connection
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("database ping failed: %w", err)
    }
    
    return db, nil
}
```

Build errors are wrapped in `ProviderBuildError` with context about which provider failed.

## Common Patterns

### Value Provider

For simple values that don't need a factory:

```go
module.ProviderDef{
    Token: "config.port",
    Build: func(r module.Resolver) (any, error) {
        return 8080, nil
    },
}
```

### Factory with Configuration

Capture configuration in the module struct:

```go
type EmailModule struct {
    smtpHost string
    smtpPort int
}

func (m *EmailModule) Definition() module.ModuleDef {
    return module.ModuleDef{
        Name: "email",
        Providers: []module.ProviderDef{{
            Token: "email.sender",
            Build: func(r module.Resolver) (any, error) {
                return NewEmailSender(m.smtpHost, m.smtpPort), nil
            },
        }},
        Exports: []module.Token{"email.sender"},
    }
}
```

### Interface-Based Providers

Define providers against interfaces for testability:

```go
// Interface
type UserRepository interface {
    FindByID(ctx context.Context, id int) (*User, error)
    Create(ctx context.Context, user *User) error
}

// Implementation
type MySQLUserRepository struct {
    db *sql.DB
}

// Provider returns interface type
Build: func(r module.Resolver) (any, error) {
    db, err := module.Get[*sql.DB](r, TokenDB)
    if err != nil {
        return nil, err
    }
    return &MySQLUserRepository{db: db}, nil
}
```

### Cleanup and Shutdown

For providers that need cleanup (database connections, file handles), handle shutdown in your application:

```go
func main() {
    app, err := kernel.Bootstrap(&AppModule{})
    if err != nil {
        log.Fatal(err)
    }
    
    // ... run server ...
    
    // Cleanup on shutdown
    if db, err := module.Get[*sql.DB](app, "db.connection"); err == nil {
        db.Close()
    }
}
```

## Cycle Detection

modkit detects circular dependencies at build time:

```go
// This will fail with ProviderCycleError
Provider A → depends on → Provider B → depends on → Provider A
```

Error message:
```text
provider cycle detected: a.service → b.service → a.service
```

**Solution:** Refactor to break the cycle, often by extracting a shared dependency.

## Tips

- Keep provider factories simple—complex logic belongs in the provider itself
- Use interfaces for providers that may have multiple implementations
- Export only the tokens that other modules actually need
- Handle errors explicitly in `Build` functions
- Test providers in isolation before integration testing modules

## Conventions (Recommended)

1. Use token format `module.component` (for example, `users.service`).
2. Keep token constants close to module definitions to avoid drift.
3. Resolve dependencies with `module.Get[T](r, token)` and return early on errors.
4. Return interface types from providers when consumers should not depend on concrete implementations.
5. Keep `Build` focused on wiring and construction, not business logic.
6. Export only tokens that are intentionally public to importing modules.

## Anti-Patterns to Avoid

- **Stringly-typed token scattering**: repeating raw token strings across packages instead of central constants.
- **Hidden cross-module dependency**: resolving a token from another module that was not exported.
- **Heavy factory logic**: embedding retries, polling, or business workflows in `Build` instead of a service method.
