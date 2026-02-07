package ast

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
