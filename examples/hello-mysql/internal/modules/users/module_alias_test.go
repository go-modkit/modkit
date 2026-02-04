package users

import "testing"

func TestUsersModuleAliasCompiles(t *testing.T) {
	var _ UsersModule = Module{}
}
