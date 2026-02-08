package kernel_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/go-modkit/modkit/modkit/kernel"
	"github.com/go-modkit/modkit/modkit/module"
)

func TestExportGraphDeterministic(t *testing.T) {
	db := mod("db", nil, nil, nil, nil)
	auth := mod("auth", nil, nil, nil, nil)
	users := mod("users", []module.Module{db}, nil, nil, nil)
	app := mod("app", []module.Module{users, auth}, nil, nil, nil)

	g, err := kernel.BuildGraph(app)
	if err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}

	tests := []struct {
		name   string
		format kernel.GraphFormat
		want   string
	}{
		{
			name:   "Mermaid",
			format: kernel.GraphFormatMermaid,
			want: strings.Join([]string{
				"graph TD",
				"    m0[\"app\"]",
				"    m1[\"auth\"]",
				"    m2[\"db\"]",
				"    m3[\"users\"]",
				"    m0 --> m1",
				"    m0 --> m3",
				"    m3 --> m2",
				"    classDef root stroke-width:3px;",
				"    class m0 root;",
			}, "\n"),
		},
		{
			name:   "DOT",
			format: kernel.GraphFormatDOT,
			want: strings.Join([]string{
				"digraph modkit {",
				"    rankdir=LR;",
				"    \"app\";",
				"    \"app\" [shape=doublecircle];",
				"    \"auth\";",
				"    \"db\";",
				"    \"users\";",
				"    \"app\" -> \"auth\";",
				"    \"app\" -> \"users\";",
				"    \"users\" -> \"db\";",
				"}",
			}, "\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := kernel.ExportGraph(g, tt.format)
			if err != nil {
				t.Fatalf("ExportGraph failed: %v", err)
			}
			if got != tt.want {
				t.Fatalf("unexpected output\n--- got ---\n%s\n--- want ---\n%s", got, tt.want)
			}
		})
	}
}

func TestExportGraphSingleModule(t *testing.T) {
	root := mod("root", nil, nil, nil, nil)
	g, err := kernel.BuildGraph(root)
	if err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}

	got, err := kernel.ExportGraph(g, kernel.GraphFormatMermaid)
	if err != nil {
		t.Fatalf("ExportGraph failed: %v", err)
	}

	want := strings.Join([]string{
		"graph TD",
		"    m0[\"root\"]",
		"    classDef root stroke-width:3px;",
		"    class m0 root;",
	}, "\n")

	if got != want {
		t.Fatalf("unexpected output\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestExportGraphReExportStillImportsOnly(t *testing.T) {
	token := module.Token("shared.token")
	shared := mod("shared", nil, []module.ProviderDef{{Token: token, Build: buildNoop}}, nil, []module.Token{token})
	mid := mod("mid", []module.Module{shared}, nil, nil, []module.Token{token})
	root := mod("root", []module.Module{mid}, nil, nil, nil)

	g, err := kernel.BuildGraph(root)
	if err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}

	got, err := kernel.ExportGraph(g, kernel.GraphFormatDOT)
	if err != nil {
		t.Fatalf("ExportGraph failed: %v", err)
	}

	if strings.Contains(got, "\"root\" -> \"shared\";") {
		t.Fatalf("expected no transitive edge, got output:\n%s", got)
	}
	if !strings.Contains(got, "\"root\" -> \"mid\";") {
		t.Fatalf("expected direct import edge root->mid, got output:\n%s", got)
	}
	if !strings.Contains(got, "\"mid\" -> \"shared\";") {
		t.Fatalf("expected direct import edge mid->shared, got output:\n%s", got)
	}
}

func TestExportGraphErrors(t *testing.T) {
	_, err := kernel.ExportGraph(nil, kernel.GraphFormatMermaid)
	if !errors.Is(err, kernel.ErrNilGraph) {
		t.Fatalf("expected ErrNilGraph, got %v", err)
	}

	graph, buildErr := kernel.BuildGraph(mod("app", nil, nil, nil, nil))
	if buildErr != nil {
		t.Fatalf("BuildGraph failed: %v", buildErr)
	}

	_, err = kernel.ExportGraph(graph, kernel.GraphFormat("json"))
	if err == nil {
		t.Fatalf("expected unsupported format error")
	}
	var unsupported *kernel.UnsupportedGraphFormatError
	if !errors.As(err, &unsupported) {
		t.Fatalf("expected UnsupportedGraphFormatError, got %T", err)
	}
	if unsupported.Format != kernel.GraphFormat("json") {
		t.Fatalf("unexpected format: %q", unsupported.Format)
	}

	malformed := &kernel.Graph{
		Root: "app",
		Modules: []kernel.ModuleNode{
			{Name: "app"},
		},
		Nodes: map[string]*kernel.ModuleNode{},
	}

	_, err = kernel.ExportGraph(malformed, kernel.GraphFormatMermaid)
	if err == nil {
		t.Fatalf("expected missing-node error for malformed graph")
	}
	var nodeMissing *kernel.GraphNodeNotFoundError
	if !errors.As(err, &nodeMissing) {
		t.Fatalf("expected GraphNodeNotFoundError, got %T", err)
	}
	if nodeMissing.Node != "app" {
		t.Fatalf("unexpected missing node name: %q", nodeMissing.Node)
	}
	if !errors.Is(err, kernel.ErrGraphNodeNotFound) {
		t.Fatalf("expected ErrGraphNodeNotFound, got %v", err)
	}
}

func TestExportAppGraphErrors(t *testing.T) {
	_, err := kernel.ExportAppGraph(nil, kernel.GraphFormatMermaid)
	if !errors.Is(err, kernel.ErrNilApp) {
		t.Fatalf("expected ErrNilApp, got %v", err)
	}

	app := &kernel.App{}
	_, err = kernel.ExportAppGraph(app, kernel.GraphFormatMermaid)
	if !errors.Is(err, kernel.ErrNilGraph) {
		t.Fatalf("expected ErrNilGraph, got %v", err)
	}
}

func TestExportAppGraphMatchesExportGraph(t *testing.T) {
	root := mod("app", []module.Module{mod("users", nil, nil, nil, nil)}, nil, nil, nil)
	app, err := kernel.Bootstrap(root)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	want, err := kernel.ExportGraph(app.Graph, kernel.GraphFormatMermaid)
	if err != nil {
		t.Fatalf("ExportGraph failed: %v", err)
	}

	got, err := kernel.ExportAppGraph(app, kernel.GraphFormatMermaid)
	if err != nil {
		t.Fatalf("ExportAppGraph failed: %v", err)
	}

	if got != want {
		t.Fatalf("expected ExportAppGraph to match ExportGraph\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestExportGraphDoesNotMutateGraph(t *testing.T) {
	modA := mod("a", nil, nil, nil, nil)
	modB := mod("b", nil, nil, nil, nil)
	modC := mod("c", nil, nil, nil, nil)
	root := mod("root", []module.Module{modC, modA, modB}, nil, nil, nil)

	g, err := kernel.BuildGraph(root)
	if err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}

	beforeModules := append([]kernel.ModuleNode(nil), g.Modules...)
	beforeRootImports := append([]string(nil), g.Nodes["root"].Imports...)

	if _, err := kernel.ExportGraph(g, kernel.GraphFormatMermaid); err != nil {
		t.Fatalf("ExportGraph mermaid failed: %v", err)
	}
	if _, err := kernel.ExportGraph(g, kernel.GraphFormatDOT); err != nil {
		t.Fatalf("ExportGraph dot failed: %v", err)
	}

	if !reflect.DeepEqual(g.Modules, beforeModules) {
		t.Fatalf("graph modules mutated\n--- got ---\n%#v\n--- want ---\n%#v", g.Modules, beforeModules)
	}
	if !reflect.DeepEqual(g.Nodes["root"].Imports, beforeRootImports) {
		t.Fatalf("root imports mutated\n--- got ---\n%v\n--- want ---\n%v", g.Nodes["root"].Imports, beforeRootImports)
	}
}

func TestExportGraphEscapesMermaidLabels(t *testing.T) {
	dep := mod("dep\\node", nil, nil, nil, nil)
	root := mod("app\"root", []module.Module{dep}, nil, nil, nil)

	g, err := kernel.BuildGraph(root)
	if err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}

	got, err := kernel.ExportGraph(g, kernel.GraphFormatMermaid)
	if err != nil {
		t.Fatalf("ExportGraph failed: %v", err)
	}

	want := strings.Join([]string{
		"graph TD",
		"    m0[\"app\\\"root\"]",
		"    m1[\"dep\\\\node\"]",
		"    m0 --> m1",
		"    classDef root stroke-width:3px;",
		"    class m0 root;",
	}, "\n")

	if got != want {
		t.Fatalf("unexpected output\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestExportGraphEscapesDOTIdentifiers(t *testing.T) {
	root := mod("line\nroot", nil, nil, nil, nil)

	g, err := kernel.BuildGraph(root)
	if err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}

	got, err := kernel.ExportGraph(g, kernel.GraphFormatDOT)
	if err != nil {
		t.Fatalf("ExportGraph failed: %v", err)
	}

	want := strings.Join([]string{
		"digraph modkit {",
		"    rankdir=LR;",
		"    \"line\\nroot\";",
		"    \"line\\nroot\" [shape=doublecircle];",
		"}",
	}, "\n")

	if got != want {
		t.Fatalf("unexpected output\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
