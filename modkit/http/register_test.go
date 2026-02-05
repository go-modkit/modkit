package http

import (
	"net/http"
	"strings"
	"testing"
)

type testController struct{ called bool }

type testControllerB struct{ called bool }

func (c *testController) RegisterRoutes(router Router) {
	c.called = true
	router.Handle(http.MethodGet, "/ping", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
}

func (c *testControllerB) RegisterRoutes(_ Router) {
	c.called = true
}

type orderedController struct {
	name  string
	order *[]string
}

func (c *orderedController) RegisterRoutes(_ Router) {
	*c.order = append(*c.order, c.name)
}

func TestRegisterRoutes_InvokesControllers(t *testing.T) {
	router := NewRouter()
	ctrl := &testController{}

	err := RegisterRoutes(AsRouter(router), map[string]any{"Test": ctrl})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !ctrl.called {
		t.Fatalf("expected controller RegisterRoutes to be called")
	}
}

func TestRegisterRoutes_ErrsOnMissingRegistrar(t *testing.T) {
	router := NewRouter()

	err := RegisterRoutes(AsRouter(router), map[string]any{"Test": struct{}{}})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestRegisterRoutes_DoesNotPartiallyRegister(t *testing.T) {
	router := NewRouter()
	ctrlA := &testController{}
	ctrlB := &testControllerB{}

	err := RegisterRoutes(AsRouter(router), map[string]any{
		"A": ctrlA,
		"B": struct{}{},
		"C": ctrlB,
	})
	if err == nil {
		t.Fatalf("expected error")
	}

	if ctrlA.called || ctrlB.called {
		t.Fatalf("expected no controllers to be registered on error")
	}
}

func TestRegisterRoutes_SortsControllerNames(t *testing.T) {
	router := NewRouter()
	order := []string{}
	controllers := map[string]any{
		"Zeta":  &orderedController{name: "Zeta", order: &order},
		"Alpha": &orderedController{name: "Alpha", order: &order},
		"Beta":  &orderedController{name: "Beta", order: &order},
	}

	err := RegisterRoutes(AsRouter(router), controllers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	joined := strings.Join(order, ",")
	if joined != "Alpha,Beta,Zeta" {
		t.Fatalf("expected alphabetical order, got %q", joined)
	}
}
