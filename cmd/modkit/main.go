// Package main is the entry point for the modkit CLI.
package main

import (
	"os"

	"github.com/go-modkit/modkit/internal/cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
