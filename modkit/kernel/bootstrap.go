// Package kernel provides the dependency injection container and module bootstrapping logic.
package kernel

import (
	"context"
	"errors"
	"io"
	"slices"
	"strconv"
	"sync/atomic"

	"github.com/go-modkit/modkit/modkit/module"
)

type bootstrapConfig struct {
	providerOverrides []ProviderOverride
	firstOptionByTok  map[module.Token]int
	optionNames       map[module.Token][]string
	currentOptionIdx  int
	err               error
}

func newBootstrapConfig() bootstrapConfig {
	return bootstrapConfig{
		providerOverrides: make([]ProviderOverride, 0),
		firstOptionByTok:  make(map[module.Token]int),
		optionNames:       make(map[module.Token][]string),
	}
}

func (c *bootstrapConfig) setCurrentOption(index int) {
	c.currentOptionIdx = index
}

func (c *bootstrapConfig) addProviderOverrides(overrides []ProviderOverride) {
	if c.err != nil {
		return
	}

	seenInThisOption := make(map[module.Token]bool)
	optionName := c.providerOverrideOptionName(c.currentOptionIdx)
	for _, override := range overrides {
		if override.Build == nil {
			c.err = &OverrideBuildNilError{Token: override.Token}
			return
		}

		if seenInThisOption[override.Token] {
			c.err = &DuplicateOverrideTokenError{Token: override.Token}
			return
		}
		seenInThisOption[override.Token] = true

		if firstIdx, ok := c.firstOptionByTok[override.Token]; ok && firstIdx != c.currentOptionIdx {
			names := append(slices.Clone(c.optionNames[override.Token]), optionName)
			c.err = &BootstrapOptionConflictError{Token: override.Token, Options: names}
			return
		}

		c.firstOptionByTok[override.Token] = c.currentOptionIdx
		c.optionNames[override.Token] = append(c.optionNames[override.Token], optionName)
		c.providerOverrides = append(c.providerOverrides, override)
	}
}

func (c *bootstrapConfig) providerOverrideOptionName(index int) string {
	return "WithProviderOverrides#" + strconv.Itoa(index+1)
}

// App represents a bootstrapped modkit application with its dependency graph,
// container, and instantiated controllers.
type App struct {
	Graph       *Graph
	container   *Container
	Controllers map[string]any
	closed      atomic.Bool
	closing     atomic.Bool
}

func controllerKey(moduleName, controllerName string) string {
	return moduleName + ":" + controllerName
}

// Bootstrap constructs a modkit application from a root module.
// It builds the module graph, validates dependencies, creates the DI container,
// and instantiates all controllers.
func Bootstrap(root module.Module) (*App, error) {
	return BootstrapWithOptions(root)
}

// BootstrapWithOptions constructs a modkit application from a root module and explicit bootstrap options.
func BootstrapWithOptions(root module.Module, opts ...BootstrapOption) (*App, error) {
	graph, err := BuildGraph(root)
	if err != nil {
		return nil, err
	}

	visibility, err := buildVisibility(graph)
	if err != nil {
		return nil, err
	}

	cfg := newBootstrapConfig()
	for idx, opt := range opts {
		if opt == nil {
			return nil, &NilBootstrapOptionError{Index: idx}
		}
		cfg.setCurrentOption(idx)
		opt.apply(&cfg)
		if cfg.err != nil {
			return nil, cfg.err
		}
	}

	providers, err := providerEntriesFromGraph(graph)
	if err != nil {
		return nil, err
	}

	for _, override := range cfg.providerOverrides {
		entry, ok := providers[override.Token]
		if !ok {
			return nil, &OverrideTokenNotFoundError{Token: override.Token}
		}
		if !visibility[graph.Root][override.Token] {
			return nil, &OverrideTokenNotVisibleFromRootError{Root: graph.Root, Token: override.Token}
		}
		entry.build = override.Build
		entry.cleanup = override.Cleanup
		providers[override.Token] = entry
	}

	container := newContainerWithProviders(providers, visibility)

	controllers := make(map[string]any)
	perModule := make(map[string]map[string]bool)
	for i := range graph.Modules {
		node := &graph.Modules[i]
		if perModule[node.Name] == nil {
			perModule[node.Name] = make(map[string]bool)
		}
		resolver := container.resolverFor(node.Name)
		for _, controller := range node.Def.Controllers {
			if perModule[node.Name][controller.Name] {
				return nil, &DuplicateControllerNameError{Module: node.Name, Name: controller.Name}
			}
			perModule[node.Name][controller.Name] = true
			instance, err := controller.Build(resolver)
			if err != nil {
				return nil, &ControllerBuildError{Module: node.Name, Controller: controller.Name, Err: err}
			}
			controllers[controllerKey(node.Name, controller.Name)] = instance
		}
	}

	return &App{
		Graph:       graph,
		container:   container,
		Controllers: controllers,
	}, nil
}

// Resolver returns a root-scoped resolver that enforces module visibility.
func (a *App) Resolver() module.Resolver {
	return a.container.resolverFor(a.Graph.Root)
}

// Get resolves a token from the root module scope.
func (a *App) Get(token module.Token) (any, error) {
	return a.Resolver().Get(token)
}

// CleanupHooks returns provider cleanup hooks in LIFO order.
func (a *App) CleanupHooks() []func(context.Context) error {
	return a.container.cleanupHooksLIFO()
}

// Closers returns provider closers in build order.
func (a *App) Closers() []io.Closer {
	return a.container.closersInBuildOrder()
}

// Close closes providers implementing io.Closer in reverse build order.
func (a *App) Close() error {
	return a.CloseContext(context.Background())
}

// CloseContext closes providers implementing io.Closer in reverse build order.
func (a *App) CloseContext(ctx context.Context) error {
	if a.closed.Load() {
		return nil
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if !a.closing.CompareAndSwap(false, true) {
		return nil
	}
	defer a.closing.Store(false)

	var errs []error
	for _, closer := range a.container.closersLIFO() {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := closer.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		a.closed.Store(true)
		return nil
	}
	a.closed.Store(true)
	return errors.Join(errs...)
}
