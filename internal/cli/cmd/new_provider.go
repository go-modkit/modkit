// Package cmd implements the CLI commands.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

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

func createNewProvider(name, moduleName string) error {
	var moduleDir string
	var err error

	if moduleName == "" {
		// Try to find module in current directory or parent
		moduleDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	} else {
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

	// Get package name from directory name if not explicit
	pkgName := filepath.Base(moduleDir)

	data := struct {
		Name    string
		Package string
		Title   func(string) string
	}{
		Name:    name,
		Package: pkgName,
		Title: func(s string) string {
			return cases.Title(language.English).String(strings.ReplaceAll(s, "-", " "))
		},
	}

	tplFS := templates.FS()
	tpl, err := template.New("provider.go.tpl").Funcs(template.FuncMap{
		"Title": data.Title,
	}).ParseFS(tplFS, "provider.go.tpl")

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

	fmt.Printf("Created %s\n", providerPath)

	tokenName := fmt.Sprintf("%s.service", strings.ToLower(name))

	fmt.Printf("TODO: Register provider in %s:\n", modulePath)
	fmt.Printf("  Token: %q\n", tokenName)
	fmt.Printf("  Build: func(r module.Resolver) (any, error) { return New%sService(), nil }\n", data.Title(name))

	return nil
}
