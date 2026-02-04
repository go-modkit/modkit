package app

import "testing"

func TestAppModuleAliasCompiles(t *testing.T) {
	var _ AppModule = Module{}
}
