// Package sqlmodule defines shared SQL contract tokens and dialect values.
package sqlmodule

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/go-modkit/modkit/modkit/module"
)

const (
	// TokenDB resolves the shared SQL database handle provider.
	TokenDB module.Token = "database.db"
	// TokenDialect resolves the SQL dialect provider.
	TokenDialect module.Token = "database.dialect"
)

// Dialect identifies the SQL engine family for a database provider.
type Dialect string

const (
	// DialectPostgres identifies PostgreSQL providers.
	DialectPostgres Dialect = "postgres"
	// DialectSQLite identifies SQLite providers.
	DialectSQLite Dialect = "sqlite"
	// DialectMySQL identifies MySQL providers.
	DialectMySQL Dialect = "mysql"
)

// Tokens contains provider tokens for a SQL module instance.
type Tokens struct {
	DB      module.Token
	Dialect module.Token
}

// InvalidNameError reports invalid SQL module instance names.
type InvalidNameError struct {
	Name   string
	Reason string
}

func (e *InvalidNameError) Error() string {
	return fmt.Sprintf("invalid sql module name: %q reason=%s", e.Name, e.Reason)
}

// NamedTokens returns deterministic tokens for a SQL module instance name.
func NamedTokens(name string) (Tokens, error) {
	if name == "" {
		return Tokens{DB: TokenDB, Dialect: TokenDialect}, nil
	}

	if strings.TrimSpace(name) == "" {
		return Tokens{}, &InvalidNameError{Name: name, Reason: "name is empty after trim"}
	}

	if strings.IndexFunc(name, unicode.IsSpace) >= 0 {
		return Tokens{}, &InvalidNameError{Name: name, Reason: "name must not contain spaces"}
	}

	return Tokens{
		DB:      module.Token("database." + name + ".db"),
		Dialect: module.Token("database." + name + ".dialect"),
	}, nil
}
