package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/kmtym1998/gh-metric/internal/github"
	"github.com/kmtym1998/gh-metric/internal/output"
	"github.com/kmtym1998/gh-metric/internal/usecase"
)

type mergedPRFlags struct {
	repo            string
	since           string
	until           string
	output          string
	excludeWeekends bool
	targetUsers     string
	limit           int
}

var mergedPRCmd = &cobra.Command{
	Use:   "merged-pr [owner/repo]",
	Short: "Calculate lead time for merged Pull Requests",
	Long: `Calculate lead time metrics for merged Pull Requests including:
- Time until first review
- Time until first approval
- Time until merge
- Comment count, file changes, and reviewer information

If no repository is specified, it will attempt to detect the repository
from the current directory's git remote.`,
	RunE: runMergedPR,
}

var flags mergedPRFlags

func init() {
	rootCmd.AddCommand(mergedPRCmd)

	// Add flags
	mergedPRCmd.Flags().StringVar(&flags.repo, "repo", "", "Specify the repository (owner/repo). If not provided, it will be detected from the current git remote.")
	mergedPRCmd.Flags().StringVar(&flags.since, "since", "", "Only include PRs merged since this date (YYYY-MM-DD)")
	mergedPRCmd.Flags().StringVar(&flags.until, "until", "", "Only include PRs merged until this date (YYYY-MM-DD)")
	mergedPRCmd.Flags().StringVar(&flags.output, "output", "csv", "Output format: json or csv")
	mergedPRCmd.Flags().BoolVar(&flags.excludeWeekends, "exclude-weekends", false, "Exclude weekends from lead time calculations")
	mergedPRCmd.Flags().StringVar(&flags.targetUsers, "target-user", "", "Only include PRs created by these users (comma-separated)")
	mergedPRCmd.Flags().IntVar(&flags.limit, "limit", 10, "Limit the number of PRs to analyze")
}

func runMergedPR(cmd *cobra.Command, args []string) error {
	slog.Info("Starting merged PR analysis",
		"repo", flags.repo,
		"since", flags.since,
		"until", flags.until,
		"output", flags.output,
		"excludeWeekends", flags.excludeWeekends,
		"targetUsers", flags.targetUsers,
		"limit", flags.limit,
	)

	// Validate flags
	if err := validateFlags(); err != nil {
		return fmt.Errorf("flag validation failed: %w", err)
	}

	// Create dependencies
	githubService, err := github.NewService()
	if err != nil {
		return fmt.Errorf("failed to create GitHub service: %w", err)
	}

	// Create usecase
	aggregateUsecase := usecase.NewAggregateMergedPRUsecase(githubService)

	targetUsers := parseTargetUsers(flags.targetUsers)

	var repoOwner, repoName string
	if flags.repo != "" {
		parts := strings.Split(flags.repo, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid repo format: %s. Use owner/repo", flags.repo)
		}
		repoOwner = parts[0]
		repoName = parts[1]
	} else {
		repoInfo, err := github.GetRepoInfoFromArgs(os.Args)
		if err != nil {
			return fmt.Errorf("failed to get repository info: %w", err)
		}
		repoOwner = repoInfo.Owner
		repoName = repoInfo.Repo
	}

	slog.Info("Analyzing repository", "owner", repoOwner, "repo", repoName)

	input := usecase.AggregateMergedPRInput{
		Owner:           repoOwner,
		Repo:            repoName,
		Since:           flags.since,
		Until:           flags.until,
		TargetUsers:     targetUsers,
		ExcludeWeekends: flags.excludeWeekends,
		Limit:           flags.limit,
	}

	output, err := aggregateUsecase.Execute(cmd.Context(), input)
	if err != nil {
		return fmt.Errorf("failed to execute usecase: %w", err)
	}

	slog.Info("Analysis complete", "prCount", len(output.Items))

	// Format and output results
	if err := outputResults(output.Items); err != nil {
		return fmt.Errorf("failed to output results: %w", err)
	}

	return nil
}

func validateFlags() error {
	// Validate output format
	if flags.output != "json" && flags.output != "csv" {
		return fmt.Errorf("invalid output format: %s. Must be 'json' or 'csv'", flags.output)
	}

	// Validate date formats
	if flags.since != "" {
		if _, err := time.Parse("2006-01-02", flags.since); err != nil {
			return fmt.Errorf("invalid since date format: %s. Use YYYY-MM-DD", flags.since)
		}
	}

	if flags.until != "" {
		if _, err := time.Parse("2006-01-02", flags.until); err != nil {
			return fmt.Errorf("invalid until date format: %s. Use YYYY-MM-DD", flags.until)
		}
	}

	// Validate target users format (should be comma-separated)
	if flags.targetUsers != "" {
		users := strings.Split(flags.targetUsers, ",")
		for _, user := range users {
			if strings.TrimSpace(user) == "" {
				return fmt.Errorf("invalid target-user format: empty user name found")
			}
		}
	}

	return nil
}

func outputResults(metrics []usecase.PRMetric) error {
	formatter, err := output.GetFormatter(flags.output)
	if err != nil {
		return err
	}

	return formatter.Format(metrics, os.Stdout)
}

// parseTargetUsers parses the comma-separated target users string
func parseTargetUsers(targetUsers string) []string {
	if targetUsers == "" {
		return nil
	}

	users := strings.Split(targetUsers, ",")
	result := make([]string, 0, len(users))
	for _, user := range users {
		if trimmed := strings.TrimSpace(user); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
