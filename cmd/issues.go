package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/kmtym1998/gh-metric/internal/github"
	"github.com/kmtym1998/gh-metric/internal/usecase"
)

type issuesFlags struct {
	repo   string
	since  string
	until  string
	output string
	state  string
	limit  int
}

var issuesCmd = &cobra.Command{
	Use:   "issues [owner/repo]",
	Short: "Aggregate issue metrics with project field values",
	Long: `Aggregate issue metrics including:
- Issue details (title, state, author, assignees, labels)
- Created and closed dates with week/month aggregation
- Project field values from GitHub Projects
- Milestone information

If no repository is specified, it will attempt to detect the repository
from the current directory's git remote.`,
	RunE: runIssues,
}

var issueFlags issuesFlags

func init() {
	rootCmd.AddCommand(issuesCmd)

	// Add flags
	issuesCmd.Flags().StringVar(&issueFlags.repo, "repo", "", "Specify the repository (owner/repo). If not provided, it will be detected from the current git remote.")
	issuesCmd.Flags().StringVar(&issueFlags.since, "since", "", "Only include issues created since this date (YYYY-MM-DD)")
	issuesCmd.Flags().StringVar(&issueFlags.until, "until", "", "Only include issues created until this date (YYYY-MM-DD)")
	issuesCmd.Flags().StringVar(&issueFlags.output, "output", "json", "Output format: json (CSV not yet supported)")
	issuesCmd.Flags().StringVar(&issueFlags.state, "state", "all", "Issue state: all, open, or closed")
	issuesCmd.Flags().IntVar(&issueFlags.limit, "limit", 100, "Limit the number of issues to analyze")
}

func runIssues(cmd *cobra.Command, args []string) error {
	slog.Info("Starting issue analysis",
		"repo", issueFlags.repo,
		"since", issueFlags.since,
		"until", issueFlags.until,
		"output", issueFlags.output,
		"state", issueFlags.state,
		"limit", issueFlags.limit,
	)

	slog.Debug("Command args", "args", args)
	slog.Debug("Flag values",
		"repo", issueFlags.repo,
		"since", issueFlags.since,
		"until", issueFlags.until,
		"output", issueFlags.output,
		"state", issueFlags.state,
		"limit", issueFlags.limit,
	)

	// Validate flags
	if err := validateIssueFlags(); err != nil {
		slog.Error("Flag validation failed", "error", err)
		return fmt.Errorf("flag validation failed: %w", err)
	}

	slog.Debug("Flags validated successfully")

	// Create dependencies
	githubService, err := github.NewService()
	if err != nil {
		slog.Error("Failed to create GitHub client", "error", err)
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}
	slog.Debug("GitHub client created successfully")

	// Create usecase
	slog.Debug("Creating aggregate issues usecase")
	aggregateUsecase := usecase.NewAggregateIssuesUsecase(githubService)

	var repoOwner, repoName string
	if issueFlags.repo != "" {
		slog.Debug("Using repo from flag", "repo", issueFlags.repo)
		parts := strings.Split(issueFlags.repo, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid repo format: %s. Use owner/repo", issueFlags.repo)
		}
		repoOwner = parts[0]
		repoName = parts[1]
	} else {
		slog.Debug("Detecting repo from git remote")
		repoInfo, err := github.GetRepoInfoFromArgs(os.Args)
		if err != nil {
			slog.Error("Failed to get repository info", "error", err)
			return fmt.Errorf("failed to get repository info: %w", err)
		}
		repoOwner = repoInfo.Owner
		repoName = repoInfo.Repo
		slog.Debug("Detected repository", "owner", repoOwner, "repo", repoName)
	}

	slog.Info("Analyzing repository", "owner", repoOwner, "repo", repoName)

	input := usecase.AggregateIssuesInput{
		Owner: repoOwner,
		Repo:  repoName,
		Since: issueFlags.since,
		Until: issueFlags.until,
		State: issueFlags.state,
		Limit: issueFlags.limit,
	}

	slog.Debug("Usecase input prepared", "input", fmt.Sprintf("%+v", input))

	output, err := aggregateUsecase.Execute(cmd.Context(), input)
	if err != nil {
		slog.Error("Failed to aggregate issue metrics", "error", err)
		return fmt.Errorf("failed to execute usecase: %w", err)
	}

	slog.Info("Analysis complete", "issueCount", len(output.Items))

	// Format and output results
	slog.Debug("Formatting output", "format", issueFlags.output)
	if err := outputIssueResults(output); err != nil {
		slog.Error("Failed to output results", "error", err)
		return fmt.Errorf("failed to output results: %w", err)
	}

	slog.Debug("Results output successfully")
	return nil
}

func validateIssueFlags() error {
	// Validate output format (currently only JSON is supported)
	if issueFlags.output != "json" {
		return fmt.Errorf("only JSON output is currently supported for issues")
	}

	// Validate state
	if issueFlags.state != "all" && issueFlags.state != "open" && issueFlags.state != "closed" {
		return fmt.Errorf("invalid state: %s. Must be 'all', 'open', or 'closed'", issueFlags.state)
	}

	// Validate date formats
	if issueFlags.since != "" {
		if _, err := time.Parse("2006-01-02", issueFlags.since); err != nil {
			return fmt.Errorf("invalid since date format: %s. Use YYYY-MM-DD", issueFlags.since)
		}
	}

	if issueFlags.until != "" {
		if _, err := time.Parse("2006-01-02", issueFlags.until); err != nil {
			return fmt.Errorf("invalid until date format: %s. Use YYYY-MM-DD", issueFlags.until)
		}
	}

	// Validate limit
	if issueFlags.limit <= 0 {
		return fmt.Errorf("limit must be greater than 0")
	}

	return nil
}

func outputIssueResults(results *usecase.AggregateIssuesOutput) error {
	// For now, only support JSON output for issues
	// TODO: Add CSV support for issues in the future
	if issueFlags.output != "json" {
		return fmt.Errorf("only JSON output is currently supported for issues")
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}