package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gh-metric",
	Short: "A GitHub CLI extension to collect repository metrics",
	Long: `gh-metric is a GitHub CLI extension that collects and outputs various metrics 
about GitHub repositories using the GitHub API.`,
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add global flags here if needed
	slog.Debug("Initializing root command")
}
