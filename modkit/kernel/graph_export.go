package kernel

import (
	"sort"
	"strconv"
	"strings"
)

// GraphFormat selects the serialization format for graph export.
type GraphFormat string

const (
	// GraphFormatMermaid exports the graph as Mermaid flowchart text.
	GraphFormatMermaid GraphFormat = "mermaid"
	// GraphFormatDOT exports the graph as Graphviz DOT text.
	GraphFormatDOT GraphFormat = "dot"
)

// ExportAppGraph exports the app's module graph in the requested format.
func ExportAppGraph(app *App, format GraphFormat) (string, error) {
	if app == nil {
		return "", ErrNilApp
	}
	if app.Graph == nil {
		return "", ErrNilGraph
	}
	return ExportGraph(app.Graph, format)
}

// ExportGraph exports a graph in Mermaid or DOT format.
func ExportGraph(g *Graph, format GraphFormat) (string, error) {
	if g == nil {
		return "", ErrNilGraph
	}

	sortedModules := sortedModuleNames(g)

	switch format {
	case GraphFormatMermaid:
		return exportMermaid(g, sortedModules)
	case GraphFormatDOT:
		return exportDOT(g, sortedModules)
	default:
		return "", &UnsupportedGraphFormatError{Format: format}
	}
}

func exportMermaid(g *Graph, sortedModules []string) (string, error) {
	lines := make([]string, 0, len(sortedModules)*2+3)
	lines = append(lines, "graph TD")

	ids := make(map[string]string, len(sortedModules))
	for i, name := range sortedModules {
		id := "m" + strconv.Itoa(i)
		ids[name] = id
		lines = append(lines, "    "+id+"[\""+escapeMermaidLabel(name)+"\"]")
	}

	for _, name := range sortedModules {
		node, err := graphNodeByName(g, name)
		if err != nil {
			return "", err
		}
		imports := append([]string(nil), node.Imports...)
		sort.Strings(imports)
		fromID := ids[name]
		for _, imported := range imports {
			toID, ok := ids[imported]
			if !ok {
				continue
			}
			lines = append(lines, "    "+fromID+" --> "+toID)
		}
	}

	if rootID, ok := ids[g.Root]; ok {
		lines = append(lines, "    classDef root stroke-width:3px;", "    class "+rootID+" root;")
	}

	return strings.Join(lines, "\n"), nil
}

func exportDOT(g *Graph, sortedModules []string) (string, error) {
	lines := make([]string, 0, len(sortedModules)*2+4)
	lines = append(lines, "digraph modkit {", "    rankdir=LR;")

	for _, name := range sortedModules {
		quoted := dotQuote(name)
		lines = append(lines, "    "+quoted+";")
		if name == g.Root {
			lines = append(lines, "    "+quoted+" [shape=doublecircle];")
		}
	}

	for _, name := range sortedModules {
		node, err := graphNodeByName(g, name)
		if err != nil {
			return "", err
		}
		imports := append([]string(nil), node.Imports...)
		sort.Strings(imports)
		for _, imported := range imports {
			if _, ok := g.Nodes[imported]; !ok {
				continue
			}
			lines = append(lines, "    "+dotQuote(name)+" -> "+dotQuote(imported)+";")
		}
	}

	lines = append(lines, "}")
	return strings.Join(lines, "\n"), nil
}

func graphNodeByName(g *Graph, name string) (*ModuleNode, error) {
	node, ok := g.Nodes[name]
	if !ok || node == nil {
		return nil, &GraphNodeNotFoundError{Node: name}
	}
	return node, nil
}

func sortedModuleNames(g *Graph) []string {
	names := make([]string, 0, len(g.Modules))
	for i := range g.Modules {
		names = append(names, g.Modules[i].Name)
	}
	sort.Strings(names)
	return names
}

func escapeMermaidLabel(s string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"\"", "\\\"",
		"\n", "\\n",
		"\r", "\\r",
	)
	return replacer.Replace(s)
}

func dotQuote(s string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"\"", "\\\"",
		"\n", "\\n",
		"\r", "\\r",
	)
	return "\"" + replacer.Replace(s) + "\""
}
