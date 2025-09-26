package main

import (
	"log/slog"
	"os"

	"github.com/kmtym1998/gh-metric/cmd"
	"github.com/lmittmann/tint"
)

func main() {
	// Initialize structured logging
	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))
	slog.SetDefault(logger)

	// Execute the root command
	if err := cmd.Execute(); err != nil {
		slog.Error("Command execution failed", "error", err)
		os.Exit(1)
	}
}
