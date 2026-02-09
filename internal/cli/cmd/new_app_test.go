package cmd

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"text/template"

	"github.com/spf13/cobra"
)

func restoreNewAppHooks(t *testing.T) {
	t.Helper()
	origReadBuildInfo := readBuildInfo
	origMkdirAll := mkdirAll
	origParseTemplateFS := parseTemplateFS
	origCreateFile := createFile
	origExecuteTemplate := executeTemplate
	origCloseWriteFile := closeWriteFile
	t.Cleanup(func() {
		readBuildInfo = origReadBuildInfo
		mkdirAll = origMkdirAll
		parseTemplateFS = origParseTemplateFS
		createFile = origCreateFile
		executeTemplate = origExecuteTemplate
		closeWriteFile = origCloseWriteFile
	})
}

func chdirTempDir(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
}

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
	expected := "github.com/go-modkit/modkit " + defaultModkitVersion
	if !strings.Contains(string(modBytes), expected) {
		t.Fatalf("expected scaffolded go.mod to use %s, got:\n%s", expected, string(modBytes))
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

func TestIsStableTagVersion(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "stable tag", in: "v0.14.1", want: true},
		{name: "pseudo version", in: "v0.0.0-20260209085719-e619ca0f7c81", want: false},
		{name: "prerelease", in: "v1.0.0-rc.1", want: false},
		{name: "build metadata", in: "v1.0.0+meta", want: false},
		{name: "empty", in: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isStableTagVersion(tt.in)
			if got != tt.want {
				t.Fatalf("isStableTagVersion(%q)=%v, want %v", tt.in, got, tt.want)
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

func TestCreateNewAppMkdirFailure(t *testing.T) {
	restoreNewAppHooks(t)
	chdirTempDir(t)

	mkdirAll = func(string, os.FileMode) error {
		return errors.New("mkdir boom")
	}

	if err := createNewApp("demo"); err == nil || !strings.Contains(err.Error(), "failed to create directory") {
		t.Fatalf("expected mkdir failure, got %v", err)
	}
}

func TestCreateNewAppTemplateParseFailure(t *testing.T) {
	restoreNewAppHooks(t)
	chdirTempDir(t)

	parseTemplateFS = func(fs.FS, ...string) (*template.Template, error) {
		return nil, errors.New("parse boom")
	}

	if err := createNewApp("demo"); err == nil || !strings.Contains(err.Error(), "failed to parse template") {
		t.Fatalf("expected parse failure, got %v", err)
	}
}

func TestCreateNewAppFileCreateFailure(t *testing.T) {
	restoreNewAppHooks(t)
	chdirTempDir(t)

	createFile = func(string) (io.WriteCloser, error) {
		return nil, errors.New("create boom")
	}

	if err := createNewApp("demo"); err == nil || !strings.Contains(err.Error(), "failed to create file") {
		t.Fatalf("expected file create failure, got %v", err)
	}
}

func TestCreateNewAppTemplateExecuteFailure(t *testing.T) {
	restoreNewAppHooks(t)
	chdirTempDir(t)

	executeTemplate = func(*template.Template, io.Writer, any) error {
		return errors.New("execute boom")
	}

	if err := createNewApp("demo"); err == nil || !strings.Contains(err.Error(), "failed to execute template") {
		t.Fatalf("expected execute failure, got %v", err)
	}
}

func TestCreateNewAppCloseFailure(t *testing.T) {
	restoreNewAppHooks(t)
	chdirTempDir(t)

	closeWriteFile = func(io.Closer) error {
		return errors.New("close boom")
	}

	if err := createNewApp("demo"); err == nil || !strings.Contains(err.Error(), "failed to close file") {
		t.Fatalf("expected close failure, got %v", err)
	}
}

func TestResolveScaffoldModkitVersionFromBuildInfo(t *testing.T) {
	restoreNewAppHooks(t)
	t.Setenv(scaffoldVersionOverrideEnv, "")

	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "v0.14.2"}}, true
	}

	got := resolveScaffoldModkitVersion()
	if got != "v0.14.2" {
		t.Fatalf("expected build info version v0.14.2, got %q", got)
	}
}

func TestResolveScaffoldModkitVersionIgnoresUnstableBuildInfo(t *testing.T) {
	restoreNewAppHooks(t)
	t.Setenv(scaffoldVersionOverrideEnv, "")

	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "v0.0.0-20260209085719-e619ca0f7c81"}}, true
	}

	got := resolveScaffoldModkitVersion()
	if got != defaultModkitVersion {
		t.Fatalf("expected fallback default %q, got %q", defaultModkitVersion, got)
	}
}
