package main

import (
	"log/slog"
	"os"

	"github.com/kmtym1998/gh-metric/cmd"
)

func main() {
	// Execute the root command
	if err := cmd.Execute(); err != nil {
		slog.Error("Command execution failed", "error", err)
		os.Exit(1)
	}
}
