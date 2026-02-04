# hello-simple

A minimal modkit example with no external dependencies (no Docker, no database).

## What This Example Shows

- Single module with providers and a controller
- Dependency injection via token resolution
- HTTP route registration
- Stateful singleton provider (counter)

## Run

```bash
go run main.go
```

## Test

```bash
# Health check
curl http://localhost:8080/health
# {"status":"ok"}

# Greeting (counter increments each call)
curl http://localhost:8080/greet
# {"count":1,"message":"Hello from modkit!"}

curl http://localhost:8080/greet
# {"count":2,"message":"Hello from modkit!"}
```

## Code Structure

```
hello-simple/
└── main.go    # Everything in one file for simplicity
```

## Key Concepts

### Module Definition

```go
func (m *AppModule) Definition() module.ModuleDef {
    return module.ModuleDef{
        Name:        "app",
        Providers:   []module.ProviderDef{...},
        Controllers: []module.ControllerDef{...},
    }
}
```

### Provider Resolution

```go
Build: func(r module.Resolver) (any, error) {
    msg, err := r.Get(TokenGreeting)
    if err != nil {
        return nil, err
    }
    return &Controller{message: msg.(string)}, nil
}
```

### Route Registration

```go
func (c *GreetingController) RegisterRoutes(r mkhttp.Router) {
    r.Handle(http.MethodGet, "/greet", http.HandlerFunc(c.Greet))
}
```

## Next Steps

- See [hello-mysql](../hello-mysql/) for a full CRUD example with database
- Read the [Getting Started Guide](../../docs/guides/getting-started.md)
- Explore [Modules](../../docs/guides/modules.md) for multi-module apps
