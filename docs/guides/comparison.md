# Comparison with Alternatives

This guide compares modkit with other Go dependency injection and application frameworks.

## Quick Comparison

| Feature | modkit | google/wire | uber-go/fx | samber/do |
|---------|--------|-------------|------------|-----------|
| Module boundaries | ✅ Yes | ❌ No | ❌ No | ❌ No |
| Visibility enforcement | ✅ Yes | ❌ No | ❌ No | ❌ No |
| Reflection-free | ✅ Yes | ✅ Yes | ❌ No | ❌ No |
| Code generation | ❌ No | ✅ Yes | ❌ No | ❌ No |
| Lifecycle hooks | ❌ No | ❌ No | ✅ Yes | ❌ No |
| HTTP routing | ✅ Built-in | ❌ No | ❌ No | ❌ No |
| Learning curve | Low | Medium | Medium | Low |

## google/wire

[Wire](https://github.com/google/wire) is a compile-time dependency injection tool that uses code generation.

### Wire Approach

```go
// wire.go
//go:build wireinject

func InitializeApp() (*App, error) {
    wire.Build(
        NewDatabase,
        NewUserRepository,
        NewUserService,
        NewApp,
    )
    return nil, nil
}
```

Run `wire` to generate the injector code.

### modkit Approach

```go
func (m *AppModule) Definition() module.ModuleDef {
    return module.ModuleDef{
        Name: "app",
        Providers: []module.ProviderDef{
            {Token: "db", Build: buildDB},
            {Token: "users.repo", Build: buildUserRepo},
            {Token: "users.service", Build: buildUserService},
        },
    }
}
```

### When to Choose

| Choose Wire when... | Choose modkit when... |
|---------------------|----------------------|
| You want compile-time safety | You want module boundaries |
| You have a flat dependency graph | You have feature modules with visibility rules |
| You prefer code generation | You prefer explicit runtime wiring |
| You don't need HTTP routing | You want integrated HTTP support |

## uber-go/fx

[Fx](https://github.com/uber-go/fx) is a dependency injection framework using reflection.

### Fx Approach

```go
func main() {
    fx.New(
        fx.Provide(
            NewDatabase,
            NewUserRepository,
            NewUserService,
        ),
        fx.Invoke(StartServer),
    ).Run()
}
```

### modkit Approach

```go
func main() {
    app, _ := kernel.Bootstrap(&AppModule{})
    router := mkhttp.NewRouter()
    mkhttp.RegisterRoutes(mkhttp.AsRouter(router), app.Controllers)
    mkhttp.Serve(":8080", router)
}
```

### Key Differences

| Aspect | Fx | modkit |
|--------|-----|--------|
| Injection | Automatic via reflection | Explicit via `module.Get[T](r, token)` |
| Lifecycle | `OnStart`/`OnStop` hooks | Manual cleanup |
| Module system | Groups (no visibility) | Full visibility enforcement |
| Type safety | Runtime type matching | Compile-time (explicit casts) |

### When to Choose

| Choose Fx when... | Choose modkit when... |
|-------------------|----------------------|
| You want automatic injection | You prefer explicit wiring |
| You need lifecycle management | You manage lifecycle manually |
| You're okay with reflection | You want no reflection |
| You have an existing Fx codebase | You want NestJS-style modules |

## samber/do

[do](https://github.com/samber/do) is a lightweight DI container using generics.

### do Approach

```go
injector := do.New()
do.Provide(injector, NewDatabase)
do.Provide(injector, NewUserService)

svc := do.MustInvoke[UserService](injector)
```

### modkit Approach

```go
app, _ := kernel.Bootstrap(&AppModule{})
svc, _ := module.Get[UserService](app, "users.service")
```

### Key Differences

| Aspect | do | modkit |
|--------|-----|--------|
| Type safety | Generics | String tokens + cast |
| Scopes | Global/Named | Module visibility |
| HTTP | Not included | Built-in |
| Size | Minimal | Includes routing |

### When to Choose

| Choose do when... | Choose modkit when... |
|-------------------|----------------------|
| You want minimal overhead | You want module organization |
| You prefer generics | You want NestJS-style structure |
| You only need DI | You need HTTP + DI |

## No Framework (Manual DI)

Many Go projects wire dependencies manually in `main()`.

### Manual Approach

```go
func main() {
    db := NewDatabase(os.Getenv("DB_DSN"))
    userRepo := NewUserRepository(db)
    userService := NewUserService(userRepo)
    userHandler := NewUserHandler(userService)
    
    router := chi.NewRouter()
    router.Get("/users", userHandler.List)
    // ...
}
```

### modkit Approach

```go
func main() {
    dbModule := database.NewModule(os.Getenv("DB_DSN"))
    usersModule := users.NewModule(dbModule)
    appModule := app.NewModule(usersModule)
    
    app, _ := kernel.Bootstrap(appModule)
    router := mkhttp.NewRouter()
    mkhttp.RegisterRoutes(mkhttp.AsRouter(router), app.Controllers)
    mkhttp.Serve(":8080", router)
}
```

### Trade-offs

| Aspect | Manual DI | modkit |
|--------|-----------|--------|
| Simplicity | Very simple | Slightly more structure |
| Refactoring | Harder at scale | Easier with modules |
| Testing | Ad-hoc mocking | Module-level testing |
| Visibility | No enforcement | Explicit exports |
| Boilerplate | Low initially, grows | Consistent |

### When to Choose

| Choose Manual DI when... | Choose modkit when... |
|--------------------------|----------------------|
| Small service | Medium-large service |
| Few dependencies | Many interconnected features |
| Solo developer | Team collaboration |
| Prototype | Long-lived codebase |

## NestJS (TypeScript)

modkit is directly inspired by [NestJS](https://nestjs.com/), bringing similar concepts to Go.

### Concept Mapping

| NestJS | modkit |
|--------|--------|
| `@Module()` decorator | `ModuleDef` struct |
| `@Injectable()` | `ProviderDef` |
| `@Controller()` | `ControllerDef` |
| Constructor injection | `module.Get[T](r, token)` |
| `imports` | `Imports` |
| `providers` | `Providers` |
| `exports` | `Exports` |
| `@Get()`, `@Post()` | `r.Handle()` |
| Guards | Auth middleware |
| Pipes | Explicit validation |
| Interceptors | Middleware wrappers |
| Exception filters | Error middleware |

### What's Different

| Aspect | NestJS | modkit |
|--------|--------|--------|
| Language | TypeScript | Go |
| Reflection | Heavy (decorators) | None |
| Metadata | Runtime decorators | Explicit structs |
| Routing | Decorator-based | Explicit registration |
| Scopes | Request/Transient/Singleton | Singleton only |

## Summary

**Choose modkit if you want:**
- NestJS-style module organization in Go
- Explicit visibility boundaries between modules
- No reflection or code generation
- Integrated HTTP routing
- Deterministic, debuggable bootstrap

**Consider alternatives if:**
- You need automatic lifecycle management → Fx
- You prefer compile-time DI → Wire
- You want minimal overhead → do or manual DI
- You need request-scoped providers → Fx
