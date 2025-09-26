package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/kmtym1998/gh-metric/internal/github"
	"github.com/kmtym1998/gh-metric/internal/output"
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
		"targetUsers", flags.targetUsers)

	// Validate flags
	if err := validateFlags(); err != nil {
		return fmt.Errorf("flag validation failed: %w", err)
	}

	// Get repository information
	repoInfo, err := getRepositoryInfo(args)
	if err != nil {
		return fmt.Errorf("failed to determine repository: %w", err)
	}

	slog.Info("Analyzing repository", "owner", repoInfo.Owner, "repo", repoInfo.Repo)

	// Create GitHub service
	service, err := createGitHubService()
	if err != nil {
		return fmt.Errorf("failed to create GitHub service: %w", err)
	}

	// Fetch PR metrics
	metrics, err := fetchPRMetrics(service, repoInfo)
	if err != nil {
		return fmt.Errorf("failed to fetch PR metrics: %w", err)
	}

	slog.Info("Analysis complete", "prCount", len(metrics))

	// Format and output results
	if err := outputResults(metrics); err != nil {
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

func getRepositoryInfo(args []string) (*github.RepoInfo, error) {
	return github.GetRepoInfoFromArgs(args)
}

func createGitHubService() (*github.Service, error) {
	return github.NewService()
}

func fetchPRMetrics(service *github.Service, repoInfo *github.RepoInfo) ([]github.PRMetrics, error) {
	targetUsers := github.ParseTargetUsers(flags.targetUsers)

	var repoOwner, repoName string
	if flags.repo != "" {
		parts := strings.Split(flags.repo, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid repository format: %s (expected owner/repo)", flags.repo)
		}
		repoOwner = parts[0]
		repoName = parts[1]
	} else {
		repoOwner = repoInfo.Owner
		repoName = repoInfo.Repo
	}

	opts := github.FetchPROptions{
		Owner:           repoOwner,
		Repo:            repoName,
		Since:           flags.since,
		Until:           flags.until,
		TargetUsers:     targetUsers,
		ExcludeWeekends: flags.excludeWeekends,
		Limit:           flags.limit,
	}

	return service.FetchMergedPRMetrics(context.Background(), opts)
}

func outputResults(metrics []github.PRMetrics) error {
	formatter, err := output.GetFormatter(flags.output)
	if err != nil {
		return err
	}

	return formatter.Format(metrics, os.Stdout)
}
