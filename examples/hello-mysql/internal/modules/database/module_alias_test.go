package database

import "testing"

func TestDatabaseModuleAliasCompiles(t *testing.T) {
	var _ DatabaseModule = Module{}
}
