package sqlite

import "github.com/go-modkit/modkit/modkit/data/sqlmodule"

// BuildStage identifies the provider build step.
type BuildStage = sqlmodule.BuildStage

const (
	// StageResolveConfig indicates a failure resolving config tokens.
	StageResolveConfig = sqlmodule.StageResolveConfig
	// StageInvalidConfig indicates invalid config values (e.g. negative settings).
	StageInvalidConfig = sqlmodule.StageInvalidConfig
	// StageOpen indicates a failure opening the database handle.
	StageOpen = sqlmodule.StageOpen
	// StagePing indicates a failure pinging the database.
	StagePing = sqlmodule.StagePing
)

// BuildError is returned when the SQLite provider fails to build.
type BuildError = sqlmodule.BuildError
