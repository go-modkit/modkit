# Getting Started

This guide assumes you are comfortable with Go modules and `net/http`. The goal is to boot a minimal modkit app and expose a route.

## Install

```bash
go get github.com/aryeko/modkit
```

## Define Tokens, Providers, and Controllers

Define tokens for your providers and create module metadata.

```go
package app

import (
    "net/http"

    mkhttp "github.com/aryeko/modkit/modkit/http"
    "github.com/aryeko/modkit/modkit/module"
)

type Tokens struct{}

const (
    TokenGreeting module.Token = "greeting"
)

type GreetingController struct {
    greeting string
}

func (c *GreetingController) RegisterRoutes(r mkhttp.Router) {
    r.Handle(http.MethodGet, "/greet", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
        _, _ = w.Write([]byte(c.greeting))
    }))
}

type AppModule struct{}

func (m AppModule) Definition() module.ModuleDef {
    return module.ModuleDef{
        Name: "app",
        Providers: []module.ProviderDef{
            {
                Token: TokenGreeting,
                Build: func(r module.Resolver) (any, error) {
                    return "hello", nil
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

## Bootstrap and Serve

```go
package main

import (
    "log"

    mkhttp "github.com/aryeko/modkit/modkit/http"
    "github.com/aryeko/modkit/modkit/kernel"

    "your/module/app"
)

func main() {
    appInstance, err := kernel.Bootstrap(app.AppModule{})
    if err != nil {
        log.Fatal(err)
    }

    router := mkhttp.NewRouter()
    if err := mkhttp.RegisterRoutes(mkhttp.AsRouter(router), appInstance.Controllers); err != nil {
        log.Fatal(err)
    }

    if err := mkhttp.Serve(":8080", router); err != nil {
        log.Fatal(err)
    }
}
```

## Next Steps

- Read `docs/guides/modules.md` for module composition and visibility rules.
- Read `docs/guides/testing.md` for testing patterns.
- Review `docs/design/mvp.md` for architecture details.
