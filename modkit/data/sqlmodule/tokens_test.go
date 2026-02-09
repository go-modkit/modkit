package sqlmodule

import (
	"errors"
	"testing"
)

func TestNamedTokens_DefaultName(t *testing.T) {
	tokens, err := NamedTokens("")
	if err != nil {
		t.Fatalf("NamedTokens(\"\") error = %v", err)
	}

	if tokens.DB != TokenDB {
		t.Fatalf("DB token = %q, want %q", tokens.DB, TokenDB)
	}
	if tokens.Dialect != TokenDialect {
		t.Fatalf("dialect token = %q, want %q", tokens.Dialect, TokenDialect)
	}
}

func TestNamedTokens_Namespace(t *testing.T) {
	tokens, err := NamedTokens("analytics")
	if err != nil {
		t.Fatalf("NamedTokens(\"analytics\") error = %v", err)
	}

	if tokens.DB != "database.analytics.db" {
		t.Fatalf("DB token = %q, want %q", tokens.DB, "database.analytics.db")
	}
	if tokens.Dialect != "database.analytics.dialect" {
		t.Fatalf("dialect token = %q, want %q", tokens.Dialect, "database.analytics.dialect")
	}
}

func TestNamedTokens_InvalidName(t *testing.T) {
	testCases := []string{"   ", "analytics reporting"}

	for _, name := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := NamedTokens(name)
			if err == nil {
				t.Fatalf("NamedTokens(%q) expected error", name)
			}

			var invalidNameErr *InvalidNameError
			if !errors.As(err, &invalidNameErr) {
				t.Fatalf("error %T is not *InvalidNameError", err)
			}
		})
	}
}
