package kernel_test

import (
	"errors"
	"testing"

	"github.com/go-modkit/modkit/modkit/kernel"
	"github.com/go-modkit/modkit/modkit/module"
)

func buildNoop(module.Resolver) (any, error) {
	return struct{}{}, nil
}

func TestVisibilityAllowsReExportFromImport(t *testing.T) {
	token := module.Token("shared.token")
	imported := mod("Imported", nil, []module.ProviderDef{{Token: token, Build: buildNoop}}, nil, []module.Token{token})
	reexporter := mod("Reexporter", []module.Module{imported}, nil, nil, []module.Token{token})

	g, err := kernel.BuildGraph(reexporter)
	if err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}

	if _, err := kernel.BuildVisibility(g); err != nil {
		t.Fatalf("BuildVisibility failed: %v", err)
	}
}

func TestVisibilityRejectsReExportOfNonExportedImportToken(t *testing.T) {
	token := module.Token("private.token")
	imported := mod("Imported", nil, []module.ProviderDef{{Token: token, Build: buildNoop}}, nil, nil)
	reexporter := mod("Reexporter", []module.Module{imported}, nil, nil, []module.Token{token})

	g, err := kernel.BuildGraph(reexporter)
	if err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}

	_, err = kernel.BuildVisibility(g)
	if err == nil {
		t.Fatalf("expected error for re-exporting non-exported token")
	}

	var exportErr *kernel.ExportNotVisibleError
	if !errors.As(err, &exportErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if exportErr.Module != "Reexporter" {
		t.Fatalf("unexpected module: %q", exportErr.Module)
	}
	if exportErr.Token != token {
		t.Fatalf("unexpected token: %q", exportErr.Token)
	}
}

func TestVisibilityRejectsAmbiguousReExport(t *testing.T) {
	token := module.Token("shared.token")
	shared := mod("Shared", nil, []module.ProviderDef{{Token: token, Build: buildNoop}}, nil, []module.Token{token})
	left := mod("Left", []module.Module{shared}, nil, nil, []module.Token{token})
	right := mod("Right", []module.Module{shared}, nil, nil, []module.Token{token})
	reexporter := mod("Reexporter", []module.Module{left, right}, nil, nil, []module.Token{token})

	g, err := kernel.BuildGraph(reexporter)
	if err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}

	_, err = kernel.BuildVisibility(g)
	if err == nil {
		t.Fatalf("expected ambiguity error")
	}
	if !errors.Is(err, kernel.ErrExportAmbiguous) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildVisibilityRejectsNilGraph(t *testing.T) {
	_, err := kernel.BuildVisibility(nil)
	if err == nil {
		t.Fatalf("expected error for nil graph")
	}
	if !errors.Is(err, kernel.ErrNilGraph) {
		t.Fatalf("unexpected error: %v", err)
	}
}
