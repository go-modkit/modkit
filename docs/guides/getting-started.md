# Getting Started

This guide walks you through building your first modkit app: a simple HTTP server with one module, one provider, and one controller.

**What you'll build:** A `/greet` endpoint that returns a greeting message.

**Prerequisites:** Go 1.22+ and familiarity with `net/http`.

## Install

```bash
go get github.com/aryeko/modkit
```

## Define Your Module

Create a module with a provider (the greeting string) and a controller (the HTTP handler):

```go
package app

import (
    "net/http"

    mkhttp "github.com/aryeko/modkit/modkit/http"
    "github.com/aryeko/modkit/modkit/module"
)

// Token identifies the greeting provider
const TokenGreeting module.Token = "greeting"

// Controller handles HTTP requests
type GreetingController struct {
    greeting string
}

func (c *GreetingController) RegisterRoutes(r mkhttp.Router) {
    r.Handle(http.MethodGet, "/greet", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
        w.Write([]byte(c.greeting))
    }))
}

// Module defines the app structure
type AppModule struct{}

func (m *AppModule) Definition() module.ModuleDef {
    return module.ModuleDef{
        Name: "app",
        Providers: []module.ProviderDef{
            {
                Token: TokenGreeting,
                Build: func(r module.Resolver) (any, error) {
                    return "Hello, modkit!", nil
                },
            },
        },
        Controllers: []module.ControllerDef{
            {
                Name: "GreetingController",
                Build: func(r module.Resolver) (any, error) {
                    value, err := r.Get(TokenGreeting)
                    if err != nil {
                        return nil, err
                    }
                    return &GreetingController{greeting: value.(string)}, nil
                },
            },
        },
        Exports: []module.Token{TokenGreeting},
    }
}
```

**Key points:**
- Modules must be passed as pointers (`&AppModule{}`) for stable identity
- Controllers must implement `mkhttp.RouteRegistrar`
- Providers are built on first access and cached as singletons

## Bootstrap and Serve

Create your `main.go`:

```go
package main

import (
    "log"

    mkhttp "github.com/aryeko/modkit/modkit/http"
    "github.com/aryeko/modkit/modkit/kernel"

    "your/module/app"
)

func main() {
    // Bootstrap the app
    appInstance, err := kernel.Bootstrap(&app.AppModule{})
    if err != nil {
        log.Fatal(err)
    }

    // Create router and register controllers
    router := mkhttp.NewRouter()
    if err := mkhttp.RegisterRoutes(mkhttp.AsRouter(router), appInstance.Controllers); err != nil {
        log.Fatal(err)
    }

    // Start server
    log.Println("Listening on :8080")
    if err := mkhttp.Serve(":8080", router); err != nil {
        log.Fatal(err)
    }
}
```

## Verify It Works

Run your app:

```bash
go run main.go
```

Test the endpoint:

```bash
curl http://localhost:8080/greet
# Hello, modkit!
```

## Next Steps

- [Modules Guide](modules.md) — Learn about imports, exports, and visibility
- [Testing Guide](testing.md) — Testing patterns for modkit apps
- [Architecture Guide](../architecture.md) — How modkit works under the hood
- [Example App](../../examples/hello-mysql/) — Full CRUD API with MySQL, migrations, and Swagger
