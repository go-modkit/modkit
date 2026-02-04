package kernel

import "github.com/aryeko/modkit/modkit/module"

type ModuleNode struct {
	Name    string
	Module  module.Module
	Def     module.ModuleDef
	Imports []string
}

type Graph struct {
	Root    string
	Modules []ModuleNode
	Nodes   map[string]*ModuleNode
}

func BuildGraph(root module.Module) (*Graph, error) {
	if root == nil {
		return nil, &DuplicateModuleNameError{Name: ""}
	}

	graph := &Graph{
		Nodes: make(map[string]*ModuleNode),
	}

	state := make(map[string]int)
	stack := make([]string, 0)

	var visit func(m module.Module) error
	visit = func(m module.Module) error {
		def := m.Definition()
		name := def.Name
		if name == "" {
			return &DuplicateModuleNameError{Name: name}
		}

		switch state[name] {
		case 1:
			idx := 0
			for i, n := range stack {
				if n == name {
					idx = i
					break
				}
			}
			path := append(append([]string{}, stack[idx:]...), name)
			return &ModuleCycleError{Path: path}
		case 2:
			return &DuplicateModuleNameError{Name: name}
		}

		state[name] = 1
		stack = append(stack, name)

		imports := make([]string, 0, len(def.Imports))
		for _, imp := range def.Imports {
			if err := visit(imp); err != nil {
				return err
			}
			imports = append(imports, imp.Definition().Name)
		}

		stack = stack[:len(stack)-1]
		state[name] = 2

		if _, exists := graph.Nodes[name]; exists {
			return &DuplicateModuleNameError{Name: name}
		}

		graph.Modules = append(graph.Modules, ModuleNode{
			Name:    name,
			Module:  m,
			Def:     def,
			Imports: imports,
		})
		graph.Nodes[name] = &graph.Modules[len(graph.Modules)-1]
		return nil
	}

	if err := visit(root); err != nil {
		return nil, err
	}

	graph.Root = root.Definition().Name

	providerTokens := make(map[module.Token]string)
	for _, node := range graph.Modules {
		for _, provider := range node.Def.Providers {
			if existing, ok := providerTokens[provider.Token]; ok {
				return nil, &DuplicateProviderTokenError{
					Token:   provider.Token,
					Modules: []string{existing, node.Name},
				}
			}
			providerTokens[provider.Token] = node.Name
		}
	}

	return graph, nil
}
