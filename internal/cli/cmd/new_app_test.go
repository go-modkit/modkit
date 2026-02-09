package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCreateNewApp(t *testing.T) {
	tmp := t.TempDir()
	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	binDir := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(binDir, 0o750); err != nil {
		t.Fatal(err)
	}
	shim := filepath.Join(binDir, "go")
	content := "#!/bin/sh\nexit 0\n"
	if runtime.GOOS == "windows" {
		shim = filepath.Join(binDir, "go.bat")
		content = "@echo off\r\nexit /b 0\r\n"
	}
	if err := os.WriteFile(shim, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}

	oldPath := os.Getenv("PATH")
	t.Cleanup(func() { _ = os.Setenv("PATH", oldPath) })
	if err := os.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath); err != nil {
		t.Fatal(err)
	}

	if err := createNewApp("demo"); err != nil {
		t.Fatalf("createNewApp failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "demo", "go.mod")); err != nil {
		t.Fatalf("expected go.mod, got %v", err)
	}

	modBytes, err := os.ReadFile(filepath.Join(tmp, "demo", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(modBytes), "github.com/go-chi/chi/v5 v5.2.4") {
		t.Fatalf("expected scaffolded go.mod to use chi v5.2.4, got:\n%s", string(modBytes))
	}
	if !strings.Contains(string(modBytes), "github.com/go-modkit/modkit v0.14.0") {
		t.Fatalf("expected scaffolded go.mod to use modkit v0.14.0 fallback, got:\n%s", string(modBytes))
	}
}

func TestResolveScaffoldModkitVersionOverride(t *testing.T) {
	t.Setenv(scaffoldVersionOverrideEnv, "0.15.1")

	got := resolveScaffoldModkitVersion()
	if got != "v0.15.1" {
		t.Fatalf("expected override version v0.15.1, got %q", got)
	}
}

func TestNormalizeSemver(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "add v prefix", in: "1.2.3", want: "v1.2.3"},
		{name: "keep v prefix", in: "v1.2.3", want: "v1.2.3"},
		{name: "reject devel", in: "(devel)", want: ""},
		{name: "reject invalid prefixed", in: "vnext", want: ""},
		{name: "reject invalid", in: "main", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeSemver(tt.in)
			if got != tt.want {
				t.Fatalf("normalizeSemver(%q)=%q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCreateNewAppInvalidName(t *testing.T) {
	if err := createNewApp("../bad"); err == nil {
		t.Fatal("expected error for invalid app name")
	}
}

func TestCreateNewAppDirectoryNotEmpty(t *testing.T) {
	tmp := t.TempDir()
	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("demo", 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("demo", "existing.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := createNewApp("demo"); err == nil {
		t.Fatal("expected error when directory exists and is not empty")
	}
}

func TestCreateNewAppExistingEmptyDirectory(t *testing.T) {
	tmp := t.TempDir()
	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll("demo", 0o750); err != nil {
		t.Fatal(err)
	}

	binDir := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(binDir, 0o750); err != nil {
		t.Fatal(err)
	}
	shim := filepath.Join(binDir, "go")
	content := "#!/bin/sh\nexit 0\n"
	if runtime.GOOS == "windows" {
		shim = filepath.Join(binDir, "go.bat")
		content = "@echo off\r\nexit /b 0\r\n"
	}
	if err := os.WriteFile(shim, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	oldPath := os.Getenv("PATH")
	t.Cleanup(func() { _ = os.Setenv("PATH", oldPath) })
	if err := os.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath); err != nil {
		t.Fatal(err)
	}

	if err := createNewApp("demo"); err != nil {
		t.Fatalf("expected createNewApp to reuse empty directory, got %v", err)
	}
}

func TestCreateNewAppGoModTidyFailure(t *testing.T) {
	tmp := t.TempDir()
	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	oldPath := os.Getenv("PATH")
	t.Cleanup(func() { _ = os.Setenv("PATH", oldPath) })
	if err := os.Setenv("PATH", ""); err != nil {
		t.Fatal(err)
	}

	if err := createNewApp("demo"); err == nil {
		t.Fatal("expected error when go mod tidy cannot run")
	}
}

func TestCreateNewAppPathIsFile(t *testing.T) {
	tmp := t.TempDir()
	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile("demo", []byte("not a dir"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := createNewApp("demo"); err == nil {
		t.Fatal("expected error when app path exists as file")
	}
}

func TestCreateNewAppRunE(t *testing.T) {
	tmp := t.TempDir()
	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	binDir := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(binDir, 0o750); err != nil {
		t.Fatal(err)
	}
	shim := filepath.Join(binDir, "go")
	content := "#!/bin/sh\nexit 0\n"
	if runtime.GOOS == "windows" {
		shim = filepath.Join(binDir, "go.bat")
		content = "@echo off\r\nexit /b 0\r\n"
	}
	if err := os.WriteFile(shim, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	oldPath := os.Getenv("PATH")
	t.Cleanup(func() { _ = os.Setenv("PATH", oldPath) })
	if err := os.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath); err != nil {
		t.Fatal(err)
	}

	if err := newAppCmd.RunE(&cobra.Command{}, []string{"runetest"}); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
}
