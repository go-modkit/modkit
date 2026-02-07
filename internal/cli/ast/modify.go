// Package ast provides AST modification utilities for modkit source files.
package ast

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

// AddProvider registers a new provider in the module definition.
func AddProvider(filePath, providerToken, buildFunc string) error {
	fset := token.NewFileSet()
	f, err := decorator.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	// Find Definition method
	found := false
	dst.Inspect(f, func(n dst.Node) bool {
		if found {
			return false
		}
		fn, ok := n.(*dst.FuncDecl)
		if !ok || fn.Name.Name != "Definition" {
			return true
		}

		// Find Providers slice in return statement
		for _, stmt := range fn.Body.List {
			ret, ok := stmt.(*dst.ReturnStmt)
			if !ok {
				continue
			}

			for _, expr := range ret.Results {
				comp, ok := expr.(*dst.CompositeLit)
				if !ok {
					continue
				}

				for _, el := range comp.Elts {
					kv, ok := el.(*dst.KeyValueExpr)
					if !ok {
						continue
					}

					key, ok := kv.Key.(*dst.Ident)
					if !ok || key.Name != "Providers" {
						continue
					}

					// Append new provider
					providersSlice, ok := kv.Value.(*dst.CompositeLit)
					if !ok {
						continue
					}

					newProvider := &dst.CompositeLit{
						Elts: []dst.Expr{
							&dst.KeyValueExpr{
								Key:   &dst.Ident{Name: "Token"},
								Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", providerToken)},
							},
							&dst.KeyValueExpr{
								Key:   &dst.Ident{Name: "Build"},
								Value: &dst.Ident{Name: buildFunc},
							},
						},
					}

					// Add newline for readability
					newProvider.Decs.Before = dst.NewLine
					newProvider.Decs.After = dst.NewLine

					providersSlice.Elts = append(providersSlice.Elts, newProvider)
					found = true
					return false // Stop traversing
				}
			}
		}
		return true
	})

	if !found {
		return fmt.Errorf("failed to find Providers field in Definition method")
	}

	// Write back to file
	// Clean path for safety
	filePath = filepath.Clean(filePath)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for writing: %w", err)
	}

	err = decorator.Fprint(file, f)
	closeErr := file.Close()

	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("failed to close file: %w", closeErr)
	}

	return nil
}
