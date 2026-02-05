package kernel

import (
	"errors"
	"testing"

	"github.com/go-modkit/modkit/modkit/module"
)

type testContainer struct {
	container *Container
}

type modHelper struct {
	def module.ModuleDef
}

func (m *modHelper) Definition() module.ModuleDef {
	return m.def
}

func mod(
	name string,
	imports []module.Module,
	providers []module.ProviderDef,
	controllers []module.ControllerDef,
	exports []module.Token,
) module.Module {
	return &modHelper{
		def: module.ModuleDef{
			Name:        name,
			Imports:     imports,
			Providers:   providers,
			Controllers: controllers,
			Exports:     exports,
		},
	}
}

func newTestContainer(t *testing.T, graph *Graph) *testContainer {
	t.Helper()
	visibility, err := BuildVisibility(graph)
	if err != nil {
		t.Fatalf("build visibility: %v", err)
	}
	container, err := newContainer(graph, visibility)
	if err != nil {
		t.Fatalf("new container: %v", err)
	}
	return &testContainer{container: container}
}

func (c *testContainer) Get(moduleName string, token module.Token) (any, error) {
	visibility := c.container.visibility[moduleName]
	if !visibility[token] {
		return nil, &TokenNotVisibleError{Module: moduleName, Token: token}
	}
	return c.container.Get(token)
}

func TestContainerGet_ReturnsProviderBuildErrorOnBuildFailure(t *testing.T) {
	boom := errors.New("boom")
	provider := module.ProviderDef{
		Token: "test.provider",
		Build: func(module.Resolver) (any, error) { return nil, boom },
	}
	mod := mod("Mod", nil, []module.ProviderDef{provider}, nil, nil)

	g, err := BuildGraph(mod)
	if err != nil {
		t.Fatalf("BuildGraph: %v", err)
	}
	c := newTestContainer(t, g)

	_, err = c.Get("Mod", provider.Token)
	if err == nil {
		t.Fatal("expected error")
	}

	var buildErr *ProviderBuildError
	if !errors.As(err, &buildErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if buildErr.Module != "Mod" || buildErr.Token != provider.Token {
		t.Fatalf("unexpected error fields")
	}
	if !errors.Is(err, boom) {
		t.Fatalf("expected wrapped error")
	}
}

func TestContainerGet_ReturnsProviderBuildErrorOnMissingDependency(t *testing.T) {
	missing := module.Token("missing")
	provider := module.ProviderDef{
		Token: "test.provider",
		Build: func(r module.Resolver) (any, error) {
			_, err := r.Get(missing)
			if err != nil {
				return nil, err
			}
			return "ok", nil
		},
	}
	mod := mod("Mod", nil, []module.ProviderDef{provider}, nil, nil)

	g, err := BuildGraph(mod)
	if err != nil {
		t.Fatalf("BuildGraph: %v", err)
	}
	c := newTestContainer(t, g)

	_, err = c.Get("Mod", provider.Token)
	if err == nil {
		t.Fatal("expected error")
	}

	var buildErr *ProviderBuildError
	if !errors.As(err, &buildErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if buildErr.Module != "Mod" || buildErr.Token != provider.Token {
		t.Fatalf("unexpected error fields")
	}

	var notVisible *TokenNotVisibleError
	if !errors.As(err, &notVisible) {
		t.Fatalf("expected TokenNotVisibleError, got %T", err)
	}
	if notVisible.Module != "Mod" || notVisible.Token != missing {
		t.Fatalf("unexpected nested error fields")
	}
}
