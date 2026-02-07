// Package cmd implements the CLI commands.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/go-modkit/modkit/internal/cli/ast"
	"github.com/go-modkit/modkit/internal/cli/templates"
)

var newControllerCmd = &cobra.Command{
	Use:   "controller [name]",
	Short: "Create a new controller",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		controllerName := args[0]
		moduleName, _ := cmd.Flags().GetString("module")
		return createNewController(controllerName, moduleName)
	},
}

func init() {
	newControllerCmd.Flags().StringP("module", "m", "", "Module to add the controller to (defaults to current directory module)")
	newCmd.AddCommand(newControllerCmd)
}

// createNewController creates a new controller file and registers it in the module.
//
// UX/Error Contract:
//
// Success:
// ✓ Created: <file-path>
// ✓ Registered in: <module-path>
//
// Partial Failure (File created, registration failed):
// ✓ Created: <file-path>
// ✗ Registration failed: <error-details>
//
//	Module: <module-path>
//	To complete manually, add to Definition().Controllers:
//	  {Name: "<name>", Build: <build-func>}
//
// Full Failure:
// ✗ Failed to <operation>: <error-details>
//
//	Target: <target-path>
//	Remediation: <actionable-guidance>
func createNewController(name, moduleName string) error {
	if err := validateScaffoldName(name, "controller name"); err != nil {
		return err
	}

	var moduleDir string
	var err error

	if moduleName == "" {
		// Try to find module in current directory or parent
		moduleDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	} else {
		if err := validateScaffoldName(moduleName, "module name"); err != nil {
			return err
		}
		// Assume standard structure: internal/modules/<name>
		moduleDir = filepath.Join("internal", "modules", moduleName)
	}

	modulePath := filepath.Join(moduleDir, "module.go")
	if _, err := os.Stat(modulePath); err != nil {
		return fmt.Errorf("module file not found at %s. Please specify a valid module with --module flag", modulePath)
	}

	// Generate controller file
	controllerFileName := strings.ToLower(strings.ReplaceAll(name, "-", "_")) + "_controller.go"
	controllerPath := filepath.Join(moduleDir, controllerFileName)

	if _, err := os.Stat(controllerPath); err == nil {
		return fmt.Errorf("controller file already exists at %s", controllerPath)
	}

	pkgName := sanitizePackageName(filepath.Base(moduleDir))

	data := struct {
		Name       string
		Package    string
		Identifier string
	}{
		Name:       name,
		Package:    pkgName,
		Identifier: exportedIdentifier(name),
	}

	tplFS := templates.FS()
	tpl, err := template.ParseFS(tplFS, "controller.go.tpl")

	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Clean path for safety
	controllerPath = filepath.Clean(controllerPath)
	f, err := os.Create(controllerPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	err = tpl.Execute(f, data)
	closeErr := f.Close()

	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("failed to close file: %w", closeErr)
	}

	controllerName := fmt.Sprintf("%sController", data.Identifier)
	buildFuncName := "New" + controllerName

	if err := ast.AddController(modulePath, controllerName, buildFuncName); err != nil {
		fmt.Printf("✓ Created: %s\n", controllerPath)
		fmt.Printf("✗ Registration failed: %v\n", err)
		fmt.Printf("  Module: %s\n", modulePath)
		fmt.Printf("  To complete manually, add to Definition().Controllers:\n")
		fmt.Printf("    {Name: %q, Build: %s}\n", controllerName, buildFuncName)
		return nil
	}

	fmt.Printf("✓ Created: %s\n", controllerPath)
	fmt.Printf("✓ Registered in: %s\n", modulePath)

	return nil
}
