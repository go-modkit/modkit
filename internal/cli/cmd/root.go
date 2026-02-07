package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "modkit",
	Short: "A CLI tool for scaffolding modkit applications",
	Long: `modkit is a CLI tool for scaffolding and managing applications
built with the modkit framework. It automates repetitive tasks like creating
modules, providers, and controllers.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags can be defined here
}
