package sqlite

import "testing"

func TestDefaultConfigModuleReturnsNewInstance(t *testing.T) {
	first := DefaultConfigModule()
	second := DefaultConfigModule()
	if first == second {
		t.Fatal("expected DefaultConfigModule to return a new module instance")
	}
}
