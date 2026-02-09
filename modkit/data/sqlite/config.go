package sqlite

import (
	"time"

	"github.com/go-modkit/modkit/modkit/config"
	"github.com/go-modkit/modkit/modkit/module"
)

// DefaultConfigModule provides SQLite configuration from environment variables.
//
// Required:
// - SQLITE_PATH
//
// Optional:
// - SQLITE_BUSY_TIMEOUT
// - SQLITE_JOURNAL_MODE
// - SQLITE_CONNECT_TIMEOUT (default 0; disables provider ping)
func DefaultConfigModule() module.Module {
	return config.NewModule(
		config.WithTyped(TokenPath, config.ValueSpec[string]{
			Key:         "SQLITE_PATH",
			Required:    true,
			Description: "SQLite database path or DSN.",
			Parse:       config.ParseString,
		}, true),
		config.WithTyped(TokenBusyTimeout, config.ValueSpec[time.Duration]{
			Key:         "SQLITE_BUSY_TIMEOUT",
			Description: "Optional busy timeout to apply to the DSN.",
			Parse:       config.ParseDuration,
		}, true),
		config.WithTyped(TokenJournalMode, config.ValueSpec[string]{
			Key:         "SQLITE_JOURNAL_MODE",
			Description: "Optional journal mode to apply to the DSN.",
			Parse:       config.ParseString,
		}, true),
		config.WithTyped(TokenConnectTimeout, config.ValueSpec[time.Duration]{
			Key:         "SQLITE_CONNECT_TIMEOUT",
			Description: "Optional ping timeout on provider build. 0 disables ping.",
			Parse:       config.ParseDuration,
		}, true),
	)
}
