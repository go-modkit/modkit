package kernel

import "github.com/aryeko/modkit/modkit/module"

type Visibility map[string]map[module.Token]bool

func buildVisibility(graph *Graph) Visibility {
	visibility := make(Visibility)
	effectiveExports := make(map[string]map[module.Token]bool)

	for _, node := range graph.Modules {
		visible := make(map[module.Token]bool)
		for _, provider := range node.Def.Providers {
			visible[provider.Token] = true
		}

		moduleExports := make(map[module.Token]bool)
		for _, token := range node.Def.Exports {
			moduleExports[token] = false
		}

		for _, impName := range node.Imports {
			for token := range effectiveExports[impName] {
				visible[token] = true
			}
		}

		for token := range moduleExports {
			if visible[token] {
				moduleExports[token] = true
				continue
			}
			for _, impName := range node.Imports {
				if effectiveExports[impName][token] {
					moduleExports[token] = true
					break
				}
			}
		}

		exports := make(map[module.Token]bool)
		for token, allowed := range moduleExports {
			if allowed {
				exports[token] = true
			}
		}

		visibility[node.Name] = visible
		effectiveExports[node.Name] = exports
		for token := range exports {
			if _, ok := visibility[node.Name][token]; ok {
				continue
			}
			visibility[node.Name][token] = true
		}
	}

	return visibility
}
