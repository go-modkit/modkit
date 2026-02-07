// Package cmd implements the CLI commands.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/go-modkit/modkit/internal/cli/templates"
)

var newModuleCmd = &cobra.Command{
	Use:   "module [name]",
	Short: "Create a new module",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		moduleName := args[0]
		return createNewModule(moduleName)
	},
}

func init() {
	newCmd.AddCommand(newModuleCmd)
}

func createNewModule(name string) error {
	// Assume we are in the project root or adjust path
	// For MVP, simplistic path resolution: internal/modules/<name>
	destDir := filepath.Join("internal", "modules", name)
	if err := os.MkdirAll(destDir, 0o750); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}

	destPath := filepath.Clean(filepath.Join(destDir, "module.go"))
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("module already exists at %s", destPath)
	}

	data := struct {
		Name       string
		Package    string
		Identifier string
	}{
		Name:       name,
		Package:    sanitizePackageName(name),
		Identifier: exportedIdentifier(name),
	}

	tplFS := templates.FS()
	tpl, err := template.ParseFS(tplFS, "module.go.tpl")

	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	f, err := os.Create(destPath)
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

	fmt.Printf("Created %s\n", destPath)
	return nil
}
