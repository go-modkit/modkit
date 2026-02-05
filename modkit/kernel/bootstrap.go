package kernel

import (
	"context"
	"io"

	"github.com/go-modkit/modkit/modkit/module"
)

type App struct {
	Graph       *Graph
	container   *Container
	Controllers map[string]any
}

func controllerKey(moduleName, controllerName string) string {
	return moduleName + ":" + controllerName
}

func Bootstrap(root module.Module) (*App, error) {
	graph, err := BuildGraph(root)
	if err != nil {
		return nil, err
	}

	visibility, err := buildVisibility(graph)
	if err != nil {
		return nil, err
	}

	container, err := newContainer(graph, visibility)
	if err != nil {
		return nil, err
	}

	controllers := make(map[string]any)
	perModule := make(map[string]map[string]bool)
	for _, node := range graph.Modules {
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
