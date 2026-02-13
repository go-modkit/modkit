package sqlmodule

import (
	"fmt"

	"github.com/go-modkit/modkit/modkit/module"
)

// BuildStage identifies the provider build step.
type BuildStage string

const (
	// StageResolveConfig indicates a failure resolving config tokens.
	StageResolveConfig BuildStage = "resolve_config"
	// StageInvalidConfig indicates invalid config values (e.g. negative settings).
	StageInvalidConfig BuildStage = "invalid_config"
	// StageOpen indicates a failure opening the database handle.
	StageOpen BuildStage = "open"
	// StagePing indicates a failure pinging the database.
	StagePing BuildStage = "ping"
)

// BuildError is returned when a SQL provider fails to build.
type BuildError struct {
	Provider string
	Token    module.Token
	Stage    BuildStage
	Err      error
}

func (e *BuildError) Error() string {
	if e.Provider == "" {
		return fmt.Sprintf("sql provider build failed: token=%q stage=%s: %v", e.Token, e.Stage, e.Err)
	}
	return fmt.Sprintf("%s provider build failed: token=%q stage=%s: %v", e.Provider, e.Token, e.Stage, e.Err)
}

func (e *BuildError) Unwrap() error {
	return e.Err
}
