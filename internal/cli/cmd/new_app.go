// Package cmd implements the CLI commands.
package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/go-modkit/modkit/internal/cli/templates"
)

const (
	scaffoldVersionOverrideEnv = "MODKIT_SCAFFOLD_VERSION"
	defaultModkitVersion       = "v0.14.0"
	defaultChiVersion          = "v5.2.4"
)

var newAppCmd = &cobra.Command{
	Use:   "app [name]",
	Short: "Create a new modkit application",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		appName := args[0]
		return createNewApp(appName)
	},
}

func init() {
	newCmd.AddCommand(newAppCmd)
}

func createNewApp(name string) error {
	if err := validateScaffoldName(name, "app name"); err != nil {
		return err
	}

	// Check if directory exists and is not empty
	if _, err := os.Stat(name); err == nil {
		entries, err := os.ReadDir(name)
		if err != nil {
			return fmt.Errorf("failed to read directory %s: %w", name, err)
		}
		if len(entries) > 0 {
			return fmt.Errorf("directory %s already exists and is not empty", name)
		}
	}

	// Create directory structure
	dirs := []string{
		name,
		filepath.Join(name, "cmd", "api"),
		filepath.Join(name, "internal", "modules", "app"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Template data
	data := struct {
		Name          string
		ModkitVersion string
		ChiVersion    string
	}{
		Name:          name,
		ModkitVersion: resolveScaffoldModkitVersion(),
		ChiVersion:    defaultChiVersion,
	}

	// Render templates
	files := map[string]string{
		"go.mod.tpl":        filepath.Join(name, "go.mod"),
		"main.go.tpl":       filepath.Join(name, "cmd", "api", "main.go"),
		"app_module.go.tpl": filepath.Join(name, "internal", "modules", "app", "module.go"),
	}

	tplFS := templates.FS()
	for tplName, destPath := range files {
		tpl, err := template.ParseFS(tplFS, tplName)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", tplName, err)
		}

		// Clean path for safety
		destPath = filepath.Clean(destPath)
		f, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", destPath, err)
		}

		err = tpl.Execute(f, data)
		closeErr := f.Close()

		if err != nil {
			return fmt.Errorf("failed to execute template %s: %w", tplName, err)
		}
		if closeErr != nil {
			return fmt.Errorf("failed to close file %s: %w", destPath, closeErr)
		}
		fmt.Printf("Created %s\n", destPath)
	}

	// Initialize go mod tidy
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = name
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run go mod tidy: %w", err)
	}

	fmt.Printf("\nSuccess! Created app '%s'\n", name)
	fmt.Printf("Run:\n  cd %s\n  go run cmd/api/main.go\n", name)

	return nil
}

func resolveScaffoldModkitVersion() string {
	if v := normalizeSemver(os.Getenv(scaffoldVersionOverrideEnv)); v != "" {
		return v
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		if v := normalizeSemver(info.Main.Version); v != "" {
			return v
		}
	}

	return defaultModkitVersion
}

func normalizeSemver(v string) string {
	v = strings.TrimSpace(v)
	if v == "" || v == "(devel)" {
		return ""
	}

	if strings.HasPrefix(v, "v") {
		if len(v) > 1 && v[1] >= '0' && v[1] <= '9' {
			return v
		}
		return ""
	}

	if v[0] < '0' || v[0] > '9' {
		return ""
	}

	return "v" + v
}
