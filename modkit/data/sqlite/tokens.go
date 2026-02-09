package sqlite

import "github.com/go-modkit/modkit/modkit/module"

const (
	// TokenPath resolves the SQLite database path or DSN.
	TokenPath module.Token = "sqlite.path" //nolint:gosec // token name, not credential
	// TokenBusyTimeout resolves the optional busy timeout setting.
	TokenBusyTimeout module.Token = "sqlite.busy_timeout" //nolint:gosec // token name, not credential
	// TokenJournalMode resolves the optional journal mode setting.
	TokenJournalMode module.Token = "sqlite.journal_mode" //nolint:gosec // token name, not credential
	// TokenConnectTimeout resolves the optional provider ping timeout.
	TokenConnectTimeout module.Token = "sqlite.connect_timeout" //nolint:gosec // token name, not credential
)
