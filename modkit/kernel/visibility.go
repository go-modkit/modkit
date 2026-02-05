package kernel

import "github.com/go-modkit/modkit/modkit/module"

type Visibility map[string]map[module.Token]bool

func BuildVisibility(graph *Graph) (Visibility, error) {
	if graph == nil {
		return nil, ErrNilGraph
	}
	return buildVisibility(graph)
}

func buildVisibility(graph *Graph) (Visibility, error) {
	visibility := make(Visibility)
	effectiveExports := make(map[string]map[module.Token]bool)

	for _, node := range graph.Modules {
		visible := make(map[module.Token]bool)
		importExporters := make(map[module.Token][]string)
		for _, provider := range node.Def.Providers {
			visible[provider.Token] = true
		}

		for _, impName := range node.Imports {
			for token := range effectiveExports[impName] {
				visible[token] = true
				importExporters[token] = append(importExporters[token], impName)
			}
		}

		exports := make(map[module.Token]bool)
		for _, token := range node.Def.Exports {
			if !visible[token] {
				return nil, &ExportNotVisibleError{Module: node.Name, Token: token}
			}
			if len(importExporters[token]) > 1 {
				return nil, &ExportAmbiguousError{
					Module:  node.Name,
					Token:   token,
					Imports: importExporters[token],
				}
			}
			exports[token] = true
		}

		visibility[node.Name] = visible
		effectiveExports[node.Name] = exports
	}

	return visibility, nil
}
