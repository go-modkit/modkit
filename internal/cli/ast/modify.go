// Package ast provides AST modification utilities for modkit source files.
package ast

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

var providerTokenPattern = regexp.MustCompile(`^[A-Za-z0-9_]+\.[A-Za-z0-9_]+$`)

// ProviderError represents an error during provider registration
type ProviderError struct {
	Op    string // operation: "parse", "validate", "find", "insert"
	Token string
	File  string
	Err   error
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("provider %s failed for %q in %s: %v", e.Op, e.Token, e.File, e.Err)
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

// Common errors
var (
	ErrDefinitionNotFound = errors.New("Definition method not found")
	ErrProvidersNotFound  = errors.New("Providers field not found in Definition")
	ErrTokenExists        = errors.New("provider token already exists")
)

// providerExists checks if a provider token already exists in the providers slice
func providerExists(providersSlice *dst.CompositeLit, providerToken string) bool {
	for _, existing := range providersSlice.Elts {
		existingComp, ok := existing.(*dst.CompositeLit)
		if !ok {
			continue
		}
		for _, elt := range existingComp.Elts {
			kv, ok := elt.(*dst.KeyValueExpr)
			if !ok {
				continue
			}
			key, ok := kv.Key.(*dst.Ident)
			if !ok || key.Name != "Token" {
				continue
			}
			if lit, ok := kv.Value.(*dst.BasicLit); ok {
				tokenValue := strings.Trim(lit.Value, `"`)
				if tokenValue == providerToken {
					return true
				}
			}
		}
	}
	return false
}

// AddProvider registers a new provider in the module definition.
func AddProvider(filePath, providerToken, buildFunc string) error {
	if !providerTokenPattern.MatchString(providerToken) {
		return &ProviderError{
			Op:    "validate",
			Token: providerToken,
			File:  filePath,
			Err:   fmt.Errorf("provider token must be in module.component format"),
		}
	}

	fset := token.NewFileSet()
	f, err := decorator.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return &ProviderError{
			Op:    "parse",
			Token: providerToken,
			File:  filePath,
			Err:   err,
		}
	}

	found, modified := findAndInsertProvider(f, providerToken, buildFunc)

	if !found {
		return &ProviderError{
			Op:    "find",
			Token: providerToken,
			File:  filePath,
			Err:   ErrProvidersNotFound,
		}
	}

	if !modified {
		return nil
	}

	return writeFileAtomically(filePath, f, providerToken)
}

// findAndInsertProvider finds the Providers slice and inserts the provider if not duplicate
func findAndInsertProvider(f *dst.File, providerToken, buildFunc string) (found, modified bool) {
	dst.Inspect(f, func(n dst.Node) bool {
		if found {
			return false
		}
		fn, ok := n.(*dst.FuncDecl)
		if !ok || fn.Name.Name != "Definition" {
			return true
		}

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

					providersSlice, ok := kv.Value.(*dst.CompositeLit)
					if !ok {
						continue
					}

					if providerExists(providersSlice, providerToken) {
						found = true
						return false
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

					newProvider.Decs.Before = dst.NewLine
					newProvider.Decs.After = dst.NewLine

					providersSlice.Elts = append(providersSlice.Elts, newProvider)
					found = true
					modified = true
					return false
				}
			}
		}
		return true
	})

	return found, modified
}

// writeFileAtomically writes the modified AST back to file atomically
func writeFileAtomically(filePath string, f *dst.File, providerToken string) error {
	filePath = filepath.Clean(filePath)
	st, err := os.Stat(filePath)
	if err != nil {
		return &ProviderError{
			Op:    "stat",
			Token: providerToken,
			File:  filePath,
			Err:   err,
		}
	}

	dir := filepath.Dir(filePath)
	tmp, err := os.CreateTemp(dir, ".modkit-*.go")
	if err != nil {
		return &ProviderError{
			Op:    "temp",
			Token: providerToken,
			File:  filePath,
			Err:   err,
		}
	}
	defer func() {
		_ = os.Remove(tmp.Name())
	}()
	if err := os.Chmod(tmp.Name(), st.Mode()); err != nil {
		return &ProviderError{
			Op:    "chmod",
			Token: providerToken,
			File:  filePath,
			Err:   err,
		}
	}

	err = decorator.Fprint(tmp, f)
	closeErr := tmp.Close()

	if err != nil {
		return &ProviderError{
			Op:    "write",
			Token: providerToken,
			File:  filePath,
			Err:   err,
		}
	}
	if closeErr != nil {
		return &ProviderError{
			Op:    "close",
			Token: providerToken,
			File:  filePath,
			Err:   closeErr,
		}
	}
	if err := os.Rename(tmp.Name(), filePath); err != nil {
		return &ProviderError{
			Op:    "rename",
			Token: providerToken,
			File:  filePath,
			Err:   err,
		}
	}

	return nil
}

// ControllerError represents an error during controller registration
type ControllerError struct {
	Op   string
	Name string
	File string
	Err  error
}

func (e *ControllerError) Error() string {
	return fmt.Sprintf("controller %s failed for %q in %s: %v", e.Op, e.Name, e.File, e.Err)
}

func (e *ControllerError) Unwrap() error {
	return e.Err
}

// ErrControllersNotFound is returned when Controllers field is not found in Definition
var ErrControllersNotFound = errors.New("Controllers field not found in Definition")

// AddController registers a new controller in the module definition
func AddController(filePath, controllerName, buildFunc string) error {
	fset := token.NewFileSet()
	f, err := decorator.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return &ControllerError{
			Op:   "parse",
			Name: controllerName,
			File: filePath,
			Err:  err,
		}
	}

	found := false
	modified := false
	dst.Inspect(f, func(n dst.Node) bool {
		if found {
			return false
		}
		fn, ok := n.(*dst.FuncDecl)
		if !ok || fn.Name.Name != "Definition" {
			return true
		}

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
					if !ok || key.Name != "Controllers" {
						continue
					}

					controllersSlice, ok := kv.Value.(*dst.CompositeLit)
					if !ok {
						continue
					}

					// Check for duplicate
					for _, existing := range controllersSlice.Elts {
						existingComp, ok := existing.(*dst.CompositeLit)
						if !ok {
							continue
						}
						for _, elt := range existingComp.Elts {
							kv, ok := elt.(*dst.KeyValueExpr)
							if !ok {
								continue
							}
							key, ok := kv.Key.(*dst.Ident)
							if !ok || key.Name != "Name" {
								continue
							}
							if lit, ok := kv.Value.(*dst.BasicLit); ok {
								nameValue := strings.Trim(lit.Value, `"`)
								if nameValue == controllerName {
									found = true
									return false
								}
							}
						}
					}

					newController := &dst.CompositeLit{
						Elts: []dst.Expr{
							&dst.KeyValueExpr{
								Key:   &dst.Ident{Name: "Name"},
								Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", controllerName)},
							},
							&dst.KeyValueExpr{
								Key:   &dst.Ident{Name: "Build"},
								Value: &dst.Ident{Name: buildFunc},
							},
						},
					}

					newController.Decs.Before = dst.NewLine
					newController.Decs.After = dst.NewLine

					controllersSlice.Elts = append(controllersSlice.Elts, newController)
					found = true
					modified = true
					return false
				}
			}
		}
		return true
	})

	if !found {
		return &ControllerError{
			Op:   "find",
			Name: controllerName,
			File: filePath,
			Err:  ErrControllersNotFound,
		}
	}

	if !modified {
		return nil
	}

	filePath = filepath.Clean(filePath)
	st, err := os.Stat(filePath)
	if err != nil {
		return &ControllerError{
			Op:   "stat",
			Name: controllerName,
			File: filePath,
			Err:  err,
		}
	}

	dir := filepath.Dir(filePath)
	tmp, err := os.CreateTemp(dir, ".modkit-*.go")
	if err != nil {
		return &ControllerError{
			Op:   "temp",
			Name: controllerName,
			File: filePath,
			Err:  err,
		}
	}
	defer func() {
		_ = os.Remove(tmp.Name())
	}()
	if err := os.Chmod(tmp.Name(), st.Mode()); err != nil {
		return &ControllerError{
			Op:   "chmod",
			Name: controllerName,
			File: filePath,
			Err:  err,
		}
	}

	err = decorator.Fprint(tmp, f)
	closeErr := tmp.Close()

	if err != nil {
		return &ControllerError{
			Op:   "write",
			Name: controllerName,
			File: filePath,
			Err:  err,
		}
	}
	if closeErr != nil {
		return &ControllerError{
			Op:   "close",
			Name: controllerName,
			File: filePath,
			Err:  closeErr,
		}
	}
	if err := os.Rename(tmp.Name(), filePath); err != nil {
		return &ControllerError{
			Op:   "rename",
			Name: controllerName,
			File: filePath,
			Err:  err,
		}
	}

	return nil
}
