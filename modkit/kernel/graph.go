package kernel

import (
	"fmt"
	"reflect"

	"github.com/go-modkit/modkit/modkit/module"
)

// ModuleNode represents a module in the dependency graph with its definition and import names.
type ModuleNode struct {
	Name    string
	Module  module.Module
	Def     module.ModuleDef
	Imports []string
}

// Graph represents the complete module dependency graph with all nodes and import relationships.
type Graph struct {
	Root    string
	Modules []ModuleNode
	Nodes   map[string]*ModuleNode
}

// BuildGraph constructs the module dependency graph starting from the root module.
// It validates module metadata, checks for cycles, and ensures all imports are valid.
func BuildGraph(root module.Module) (*Graph, error) {
	if root == nil {
		return nil, &RootModuleNilError{}
	}
	rootVal := reflect.ValueOf(root)
	if rootVal.Kind() == reflect.Ptr && rootVal.IsNil() {
		return nil, &RootModuleNilError{}
	}

	graph := &Graph{
		Nodes: make(map[string]*ModuleNode),
	}

	state := make(map[string]int)
	stack := make([]string, 0)
	identities := make(map[string]uintptr)

	var visit func(m module.Module) error
	visit = func(m module.Module) error {
		if m == nil {
			return &NilImportError{Module: "", Index: -1}
		}
		val := reflect.ValueOf(m)
		if val.Kind() == reflect.Ptr && val.IsNil() {
			return &NilImportError{Module: "", Index: -1}
		}
		def := m.Definition()
		if val.Kind() != reflect.Ptr {
			return &ModuleNotPointerError{Module: def.Name}
		}
		if err := validateModuleDef(&def); err != nil {
			return err
		}
		name := def.Name

		id := uintptr(0)
		id = val.Pointer()

		if id == 0 {
			if _, ok := identities[name]; ok {
				return &DuplicateModuleNameError{Name: name}
			}
		} else if existing, ok := identities[name]; ok {
			if existing != id {
				return &DuplicateModuleNameError{Name: name}
			}
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
			return nil
		}

		state[name] = 1
		if id == 0 {
			identities[name] = 0
		} else {
			identities[name] = id
		}
		stack = append(stack, name)

		imports := make([]string, 0, len(def.Imports))
		for idx, imp := range def.Imports {
			if imp == nil {
				return &NilImportError{Module: name, Index: idx}
			}
			impVal := reflect.ValueOf(imp)
			if impVal.Kind() == reflect.Ptr && impVal.IsNil() {
				return &NilImportError{Module: name, Index: idx}
			}
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
	for i := range graph.Modules {
		node := &graph.Modules[i]
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

func validateModuleDef(def *module.ModuleDef) error {
	if def.Name == "" {
		return &InvalidModuleDefError{Module: def.Name, Reason: "module name is empty"}
	}
	for i, provider := range def.Providers {
		if provider.Token == "" {
			return &InvalidModuleDefError{Module: def.Name, Reason: fmt.Sprintf("provider[%d] token is empty", i)}
		}
		if provider.Build == nil {
			return &InvalidModuleDefError{Module: def.Name, Reason: fmt.Sprintf("provider[%d] build is nil", i)}
		}
	}
	for i, controller := range def.Controllers {
		if controller.Name == "" {
			return &InvalidModuleDefError{Module: def.Name, Reason: fmt.Sprintf("controller[%d] name is empty", i)}
		}
		if controller.Build == nil {
			return &InvalidModuleDefError{Module: def.Name, Reason: fmt.Sprintf("controller[%d] build is nil", i)}
		}
	}
	for i, token := range def.Exports {
		if token == "" {
			return &InvalidModuleDefError{Module: def.Name, Reason: fmt.Sprintf("export[%d] token is empty", i)}
		}
	}
	return nil
}
