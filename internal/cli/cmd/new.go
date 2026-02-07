// Package cmd implements the CLI commands.
package cmd

import (
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new app, module, provider, or controller",
}

func init() {
	rootCmd.AddCommand(newCmd)
}
