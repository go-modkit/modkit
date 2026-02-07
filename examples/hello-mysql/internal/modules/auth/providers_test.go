package auth

import (
	"errors"
	"net/http"
	"testing"
	"time"

	configmodule "github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/config"
	"github.com/go-modkit/modkit/modkit/module"
)

type mapResolver map[module.Token]any

func (r mapResolver) Get(token module.Token) (any, error) {
	v, ok := r[token]
	if !ok {
		return nil, errors.New("missing token")
	}
	return v, nil
}

func TestAuthProviders_BuildsHandlerAndMiddleware(t *testing.T) {
	defs := Providers()
	r := mapResolver{
		configmodule.TokenJWTSecret:    "secret",
		configmodule.TokenJWTIssuer:    "issuer",
		configmodule.TokenJWTTTL:       time.Minute,
		configmodule.TokenAuthUsername: "demo",
		configmodule.TokenAuthPassword: "demo",
	}

	var handlerBuilt, mwBuilt bool
	for _, def := range defs {
		value, err := def.Build(r)
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		switch def.Token {
		case TokenHandler:
			_, handlerBuilt = value.(*Handler)
		case TokenMiddleware:
			_, mwBuilt = value.(func(http.Handler) http.Handler)
		}
	}
	if !handlerBuilt || !mwBuilt {
		t.Fatalf("handler=%v middleware=%v", handlerBuilt, mwBuilt)
	}
}
