package cmd

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
)

var (
	debugMode bool
)

var rootCmd = &cobra.Command{
	Use:   "gh-metric",
	Short: "A GitHub CLI extension to collect repository metrics",
	Long: `gh-metric is a GitHub CLI extension that collects and outputs various metrics
about GitHub repositories using the GitHub API.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupLogger()
	},
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

// setupLogger configures the logger based on debug mode
func setupLogger() {
	var level slog.Level
	var addSource bool

	if debugMode {
		level = slog.LevelDebug
		addSource = true
		slog.Debug("Debug mode enabled")
	} else {
		level = slog.LevelWarn
		addSource = false
	}

	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:     level,
		AddSource: addSource,
	}))
	slog.SetDefault(logger)

	slog.Debug("Logger initialized", "level", level.String(), "addSource", addSource)
}

func init() {
	// Add global persistent flags
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, "Enable debug mode with verbose logging")
}
