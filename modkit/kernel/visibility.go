package kernel

import "github.com/aryeko/modkit/modkit/module"

type Visibility map[string]map[module.Token]bool

func buildVisibility(graph *Graph) (Visibility, error) {
	visibility := make(Visibility)
	effectiveExports := make(map[string]map[module.Token]bool)

	for _, node := range graph.Modules {
		visible := make(map[module.Token]bool)
		for _, provider := range node.Def.Providers {
			visible[provider.Token] = true
		}

		for _, impName := range node.Imports {
			for token := range effectiveExports[impName] {
				visible[token] = true
			}
		}

		exports := make(map[module.Token]bool)
		for _, token := range node.Def.Exports {
			if !visible[token] {
				return nil, &ExportNotVisibleError{Module: node.Name, Token: token}
			}
			exports[token] = true
		}

		visibility[node.Name] = visible
		effectiveExports[node.Name] = exports
	}

	return visibility, nil
}
