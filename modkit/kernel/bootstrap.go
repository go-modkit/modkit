package kernel

import "github.com/aryeko/modkit/modkit/module"

type App struct {
	Graph       *Graph
	Container   *Container
	Controllers map[string]any
}

func Bootstrap(root module.Module) (*App, error) {
	graph, err := BuildGraph(root)
	if err != nil {
		return nil, err
	}

	visibility := buildVisibility(graph)

	container, err := newContainer(graph, visibility)
	if err != nil {
		return nil, err
	}

	controllers := make(map[string]any)
	for _, node := range graph.Modules {
		resolver := container.resolverFor(node.Name)
		for _, controller := range node.Def.Controllers {
			instance, err := controller.Build(resolver)
			if err != nil {
				return nil, &ControllerBuildError{Module: node.Name, Controller: controller.Name, Err: err}
			}
			controllers[controller.Name] = instance
		}
	}

	return &App{
		Graph:       graph,
		Container:   container,
		Controllers: controllers,
	}, nil
}
