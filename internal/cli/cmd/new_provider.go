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

var newProviderCmd = &cobra.Command{
	Use:   "provider [name]",
	Short: "Create a new provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		providerName := args[0]
		moduleName, _ := cmd.Flags().GetString("module")
		return createNewProvider(providerName, moduleName)
	},
}

func init() {
	newProviderCmd.Flags().StringP("module", "m", "", "Module to add the provider to (defaults to current directory module)")
	newCmd.AddCommand(newProviderCmd)
}

// createNewProvider creates a new provider file and registers it in the module.
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
//	To complete manually, add to Definition().Providers:
//	  {Token: "<token>", Build: <build-func>}
//
// Full Failure:
// ✗ Failed to <operation>: <error-details>
//
//	Target: <target-path>
//	Remediation: <actionable-guidance>
func createNewProvider(name, moduleName string) error {
	if err := validateScaffoldName(name, "provider name"); err != nil {
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

	// snake_case file name
	providerFileName := strings.ToLower(strings.ReplaceAll(name, "-", "_")) + ".go"
	providerPath := filepath.Join(moduleDir, providerFileName)

	if _, err := os.Stat(providerPath); err == nil {
		return fmt.Errorf("provider file already exists at %s", providerPath)
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
	tpl, err := template.ParseFS(tplFS, "provider.go.tpl")

	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Clean path for safety
	providerPath = filepath.Clean(providerPath)
	f, err := os.Create(providerPath)
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

	tokenName := fmt.Sprintf("%s.%s", pkgName, strings.ToLower(name))
	buildFuncName := "New" + data.Identifier + "Service"

	// Attempt to register provider in module
	if err := ast.AddProvider(modulePath, tokenName, buildFuncName); err != nil {
		// Partial failure: file created but registration failed
		fmt.Printf("✓ Created: %s\n", providerPath)
		fmt.Printf("✗ Registration failed: %v\n", err)
		fmt.Printf("  Module: %s\n", modulePath)
		fmt.Printf("  To complete manually, add to Definition().Providers:\n")
		fmt.Printf("    {Token: %q, Build: %s}\n", tokenName, buildFuncName)
		return nil
	}

	// Success
	fmt.Printf("✓ Created: %s\n", providerPath)
	fmt.Printf("✓ Registered in: %s\n", modulePath)

	return nil
}
