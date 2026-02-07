package ast

import (
	"errors"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func TestAddProvider(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "module.go")
	original := `package users

import "github.com/go-modkit/modkit/modkit/module"

type Module struct{}

func (m *Module) Definition() module.ModuleDef {
	return module.ModuleDef{
		Name: "users",
		Providers: []module.ProviderDef{},
	}
}
`
	if err := os.WriteFile(file, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := AddProvider(file, "users.auth", "buildAuth"); err != nil {
		t.Fatalf("AddProvider failed: %v", err)
	}

	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, `Token: "users.auth"`) {
		t.Fatalf("expected token in output:\n%s", s)
	}
	if !strings.Contains(s, `Build: buildAuth`) {
		t.Fatalf("expected build func in output:\n%s", s)
	}
}

func TestAddProviderNoProvidersField(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "module.go")
	content := `package users

import "github.com/go-modkit/modkit/modkit/module"

type Module struct{}

func (m *Module) Definition() module.ModuleDef {
	return module.ModuleDef{Name: "users"}
}
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := AddProvider(file, "users.auth", "buildAuth"); err == nil {
		t.Fatal("expected error when Providers field is missing")
	} else {
		var perr *ProviderError
		if !errors.As(err, &perr) {
			t.Fatalf("expected ProviderError, got %T", err)
		}
		if !errors.Is(err, ErrProvidersNotFound) {
			t.Fatalf("expected ErrProvidersNotFound, got %v", err)
		}
		if perr.Error() == "" {
			t.Fatal("expected non-empty ProviderError message")
		}
	}
}

func TestAddProviderParseError(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "module.go")
	if err := os.WriteFile(file, []byte("package users\nfunc ("), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := AddProvider(file, "users.auth", "buildAuth"); err == nil {
		t.Fatal("expected parse error")
	} else {
		var perr *ProviderError
		if !errors.As(err, &perr) {
			t.Fatalf("expected ProviderError, got %T", err)
		}
		if perr.Unwrap() == nil {
			t.Fatal("expected wrapped parse error")
		}
	}
}

func TestAddProviderNoDefinitionMethod(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "module.go")
	content := `package users

type Module struct{}
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := AddProvider(file, "users.auth", "buildAuth"); err == nil {
		t.Fatal("expected error when Definition method is missing")
	}
}

func TestAddProviderInvalidToken(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "module.go")
	content := `package users

import "github.com/go-modkit/modkit/modkit/module"

type Module struct{}

func (m *Module) Definition() module.ModuleDef {
	return module.ModuleDef{
		Name: "users",
		Providers: []module.ProviderDef{},
	}
}
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := AddProvider(file, "bad token", "buildAuth"); err == nil {
		t.Fatal("expected error for invalid token format")
	}
}

func TestAddProviderDuplicate(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "module.go")
	content := `package users

import "github.com/go-modkit/modkit/modkit/module"

type Module struct{}

func (m *Module) Definition() module.ModuleDef {
	return module.ModuleDef{
		Name: "users",
		Providers: []module.ProviderDef{{
			Token: "users.auth",
			Build: func(r module.Resolver) (any, error) { return nil, nil },
		}},
	}
}
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	// First call should succeed (no duplicate yet)
	if err := AddProvider(file, "users.service", "buildService"); err != nil {
		t.Fatalf("First AddProvider failed: %v", err)
	}

	// Second call with same token should be idempotent (no error, no duplicate)
	if err := AddProvider(file, "users.auth", "buildAuth"); err != nil {
		t.Fatalf("Duplicate AddProvider should succeed idempotently: %v", err)
	}

	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)

	// Count occurrences of users.auth token
	count := strings.Count(s, `Token: "users.auth"`)
	if count != 1 {
		t.Fatalf("expected 1 occurrence of users.auth token, got %d", count)
	}

	// Should still have users.service
	if !strings.Contains(s, `Token: "users.service"`) {
		t.Fatal("expected users.service token to be present")
	}
}

func TestAddController(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "module.go")
	original := `package users

import "github.com/go-modkit/modkit/modkit/module"

type Module struct{}

func (m *Module) Definition() module.ModuleDef {
	return module.ModuleDef{
		Name: "users",
		Controllers: []module.ControllerDef{},
	}
}
`
	if err := os.WriteFile(file, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := AddController(file, "UsersController", "NewUsersController"); err != nil {
		t.Fatalf("AddController failed: %v", err)
	}

	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, `Name: "UsersController"`) {
		t.Fatalf("expected controller name in output:\n%s", s)
	}
	if !strings.Contains(s, `Build: NewUsersController`) {
		t.Fatalf("expected build func in output:\n%s", s)
	}
}

func TestAddControllerDuplicate(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "module.go")
	content := `package users

import "github.com/go-modkit/modkit/modkit/module"

type Module struct{}

func (m *Module) Definition() module.ModuleDef {
	return module.ModuleDef{
		Name: "users",
		Controllers: []module.ControllerDef{{
			Name: "AuthController",
			Build: func(r module.Resolver) (any, error) { return nil, nil },
		}},
	}
}
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	// First call should succeed
	if err := AddController(file, "UsersController", "NewUsersController"); err != nil {
		t.Fatalf("First AddController failed: %v", err)
	}

	// Second call with same name should be idempotent
	if err := AddController(file, "AuthController", "NewAuthController"); err != nil {
		t.Fatalf("Duplicate AddController should succeed idempotently: %v", err)
	}

	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)

	count := strings.Count(s, "AuthController")
	if count != 1 {
		t.Fatalf("expected 1 occurrence of AuthController, got %d", count)
	}

	if !strings.Contains(s, "UsersController") {
		t.Fatal("expected UsersController to be present")
	}
}

func TestAddControllerNoControllersField(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "module.go")
	content := `package users

import "github.com/go-modkit/modkit/modkit/module"

type Module struct{}

func (m *Module) Definition() module.ModuleDef {
	return module.ModuleDef{Name: "users"}
}
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	err := AddController(file, "UsersController", "NewUsersController")
	if err == nil {
		t.Fatal("expected error when Controllers field is missing")
	}

	var cerr *ControllerError
	if !errors.As(err, &cerr) {
		t.Fatalf("expected ControllerError, got %T", err)
	}
	if !errors.Is(err, ErrControllersNotFound) {
		t.Fatalf("expected ErrControllersNotFound, got %v", err)
	}
	if cerr.Error() == "" {
		t.Fatal("expected non-empty ControllerError message")
	}
}

func TestAddControllerParseError(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "module.go")
	if err := os.WriteFile(file, []byte("package users\nfunc ("), 0o600); err != nil {
		t.Fatal(err)
	}

	err := AddController(file, "UsersController", "NewUsersController")
	if err == nil {
		t.Fatal("expected parse error")
	}

	var cerr *ControllerError
	if !errors.As(err, &cerr) {
		t.Fatalf("expected ControllerError, got %T", err)
	}
	if cerr.Unwrap() == nil {
		t.Fatal("expected wrapped parse error")
	}
}

func TestAddControllerNoDefinitionMethod(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "module.go")
	content := `package users

type Module struct{}
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	err := AddController(file, "UsersController", "NewUsersController")
	if err == nil {
		t.Fatal("expected error when Definition method is missing")
	}
}

func TestProviderExistsSkipsUnexpectedShapes(t *testing.T) {
	providers := &dst.CompositeLit{
		Elts: []dst.Expr{
			&dst.Ident{Name: "notAComposite"},
			&dst.CompositeLit{Elts: []dst.Expr{&dst.Ident{Name: "notAKeyValue"}}},
		},
	}

	if providerExists(providers, "users.auth") {
		t.Fatal("expected false when providers contain unexpected shapes")
	}
}

func TestWriteFileAtomicallyStatError(t *testing.T) {
	err := writeFileAtomically(filepath.Join(t.TempDir(), "missing.go"), &dst.File{}, "users.auth")
	if err == nil {
		t.Fatal("expected stat error")
	}

	var perr *ProviderError
	if !errors.As(err, &perr) {
		t.Fatalf("expected ProviderError, got %T", err)
	}
	if perr.Op != "stat" {
		t.Fatalf("expected stat op, got %q", perr.Op)
	}
}

func TestWriteFileAtomicallyRenameError(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "module.go")
	if err := os.WriteFile(src, []byte("package users\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()
	f, err := decorator.ParseFile(fset, src, nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	targetDir := filepath.Join(tmp, "target")
	if err := os.Mkdir(targetDir, 0o750); err != nil {
		t.Fatal(err)
	}

	err = writeFileAtomically(targetDir, f, "users.auth")
	if err == nil {
		t.Fatal("expected rename error")
	}

	var perr *ProviderError
	if !errors.As(err, &perr) {
		t.Fatalf("expected ProviderError, got %T", err)
	}
	if perr.Op != "rename" {
		t.Fatalf("expected rename op, got %q", perr.Op)
	}
}

func TestAddControllerDuplicateSkipsUnexpectedExistingShape(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "module.go")
	content := `package users

import "github.com/go-modkit/modkit/modkit/module"

func existingController() module.ControllerDef {
	return module.ControllerDef{Name: "AuthController"}
}

type Module struct{}

func (m *Module) Definition() module.ModuleDef {
	return module.ModuleDef{
		Name: "users",
		Controllers: []module.ControllerDef{existingController()},
	}
}
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := AddController(file, "UsersController", "NewUsersController"); err != nil {
		t.Fatalf("expected insertion to succeed: %v", err)
	}
}

func TestAddProviderSkipsUnsupportedTraversalShapes(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "module.go")
	content := `package users

import "github.com/go-modkit/modkit/modkit/module"

type Module struct{}

func getDef() module.ModuleDef {
	return module.ModuleDef{Name: "users"}
}

func buildProviders() []module.ProviderDef {
	return nil
}

func (m *Module) Definition() module.ModuleDef {
	x := 1
	_ = x
	if x == 2 {
		return getDef()
	}
	if x == 1 {
		return module.ModuleDef{"users"}
	}
	return module.ModuleDef{
		Name:      "users",
		Providers: buildProviders(),
	}
}
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	err := AddProvider(file, "users.auth", "buildAuth")
	if err == nil {
		t.Fatal("expected error when Providers field has unsupported shapes")
	}
}

func TestAddControllerSkipsUnsupportedTraversalShapes(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "module.go")
	content := `package users

import "github.com/go-modkit/modkit/modkit/module"

type Module struct{}

func getDef() module.ModuleDef {
	return module.ModuleDef{Name: "users"}
}

func buildControllers() []module.ControllerDef {
	return nil
}

func (m *Module) Definition() module.ModuleDef {
	x := 1
	_ = x
	if x == 2 {
		return getDef()
	}
	if x == 1 {
		return module.ModuleDef{"users"}
	}
	return module.ModuleDef{
		Name:        "users",
		Controllers: buildControllers(),
	}
}
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	err := AddController(file, "UsersController", "NewUsersController")
	if err == nil {
		t.Fatal("expected error when Controllers field has unsupported shapes")
	}
}

func TestAddProviderTempError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod-based permission errors are unreliable on Windows")
	}

	tmp := t.TempDir()
	moduleDir := filepath.Join(tmp, "module")
	if err := os.MkdirAll(moduleDir, 0o750); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(moduleDir, "module.go")
	content := `package users

import "github.com/go-modkit/modkit/modkit/module"

type Module struct{}

func (m *Module) Definition() module.ModuleDef {
	return module.ModuleDef{
		Name: "users",
		Providers: []module.ProviderDef{},
	}
}
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.Chmod(moduleDir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(moduleDir, 0o750) })

	err := AddProvider(file, "users.auth", "buildAuth")
	if err == nil {
		t.Fatal("expected temp file creation error")
	}
	var perr *ProviderError
	if !errors.As(err, &perr) {
		t.Fatalf("expected ProviderError, got %T", err)
	}
	if perr.Op != "temp" {
		t.Fatalf("expected temp op, got %q", perr.Op)
	}
}

func TestAddControllerTempError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod-based permission errors are unreliable on Windows")
	}

	tmp := t.TempDir()
	moduleDir := filepath.Join(tmp, "module")
	if err := os.MkdirAll(moduleDir, 0o750); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(moduleDir, "module.go")
	content := `package users

import "github.com/go-modkit/modkit/modkit/module"

type Module struct{}

func (m *Module) Definition() module.ModuleDef {
	return module.ModuleDef{
		Name: "users",
		Controllers: []module.ControllerDef{},
	}
}
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.Chmod(moduleDir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(moduleDir, 0o750) })

	err := AddController(file, "UsersController", "NewUsersController")
	if err == nil {
		t.Fatal("expected temp file creation error")
	}
	var cerr *ControllerError
	if !errors.As(err, &cerr) {
		t.Fatalf("expected ControllerError, got %T", err)
	}
	if cerr.Op != "temp" {
		t.Fatalf("expected temp op, got %q", cerr.Op)
	}
}

func TestAddProviderTopLevelReturnNonComposite(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "module.go")
	content := `package users

import "github.com/go-modkit/modkit/modkit/module"

type Module struct{}

func getDef() module.ModuleDef {
	return module.ModuleDef{Name: "users"}
}

func (m *Module) Definition() module.ModuleDef {
	return getDef()
}
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	err := AddProvider(file, "users.auth", "buildAuth")
	if err == nil {
		t.Fatal("expected error for non-composite top-level return")
	}
}

func TestAddProviderTopLevelCompositeWithNonKeyValueElements(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "module.go")
	content := `package users

import "github.com/go-modkit/modkit/modkit/module"

type Module struct{}

func (m *Module) Definition() module.ModuleDef {
	return module.ModuleDef{"users"}
}
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	err := AddProvider(file, "users.auth", "buildAuth")
	if err == nil {
		t.Fatal("expected error for composite literal without key-value elements")
	}
}

func TestAddControllerTopLevelReturnNonComposite(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "module.go")
	content := `package users

import "github.com/go-modkit/modkit/modkit/module"

type Module struct{}

func getDef() module.ModuleDef {
	return module.ModuleDef{Name: "users"}
}

func (m *Module) Definition() module.ModuleDef {
	return getDef()
}
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	err := AddController(file, "UsersController", "NewUsersController")
	if err == nil {
		t.Fatal("expected error for non-composite top-level return")
	}
}

func TestAddControllerTopLevelCompositeWithNonKeyValueElements(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "module.go")
	content := `package users

import "github.com/go-modkit/modkit/modkit/module"

type Module struct{}

func (m *Module) Definition() module.ModuleDef {
	return module.ModuleDef{"users"}
}
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	err := AddController(file, "UsersController", "NewUsersController")
	if err == nil {
		t.Fatal("expected error for composite literal without key-value elements")
	}
}

func TestAddControllerDuplicateScanSkipsNonKeyValueElements(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "module.go")
	content := `package users

import "github.com/go-modkit/modkit/modkit/module"

type Module struct{}

func (m *Module) Definition() module.ModuleDef {
	return module.ModuleDef{
		Name: "users",
		Controllers: []module.ControllerDef{{someInvalidExpr}},
	}
}
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	err := AddController(file, "UsersController", "NewUsersController")
	if err != nil {
		t.Fatalf("expected insertion to succeed despite invalid existing shape: %v", err)
	}
}
