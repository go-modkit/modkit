package kernel

import "github.com/aryeko/modkit/modkit/module"

type Visibility map[string]map[module.Token]bool

func buildVisibility(graph *Graph) Visibility {
	visibility := make(Visibility)

	for _, node := range graph.Modules {
		visible := make(map[module.Token]bool)
		for _, provider := range node.Def.Providers {
			visible[provider.Token] = true
		}
		for _, impName := range node.Imports {
			impNode := graph.Nodes[impName]
			if impNode == nil {
				continue
			}
			for _, token := range impNode.Def.Exports {
				visible[token] = true
			}
		}
		visibility[node.Name] = visible
	}

	return visibility
}
