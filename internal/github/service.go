package github

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
)

// Service provides GitHub API operations
type Service struct {
	client *Client
}

// NewService creates a new GitHub service
func NewService() (*Service, error) {
	client, err := NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	return &Service{
		client: client,
	}, nil
}

// FetchPROptions contains options for fetching pull requests
type FetchPROptions struct {
	Owner           string
	Repo            string
	Since           string
	Until           string
	TargetUsers     []string
	ExcludeWeekends bool
	Limit           int
}

// FetchMergedPRMetrics fetches and calculates metrics for merged pull requests
func (s *Service) FetchMergedPRMetrics(ctx context.Context, opts FetchPROptions) ([]PRMetrics, error) {
	// Build the search query
	queryBuilder := NewSearchQueryBuilder()
	queryBuilder.AddRepository(opts.Owner, opts.Repo)
	queryBuilder.AddDateRange(opts.Since, opts.Until)

	searchQuery := queryBuilder.Build()
	slog.Info("Executing search query", "query", searchQuery)

	var allPRs []PullRequest
	var cursor *string

	// Paginate through all results
	for {
		variables := QueryVariables{
			Query:  searchQuery,
			Cursor: cursor,
		}

		var response SearchResponse
		if err := s.client.ExecuteQuery(pullRequestQuery, variables, &response); err != nil {
			return nil, fmt.Errorf("failed to execute GraphQL query: %w", err)
		}

		for _, pr := range response.Search.Nodes {
			if slices.Contains(opts.TargetUsers, pr.Author.Login) || len(opts.TargetUsers) == 0 {
				allPRs = append(allPRs, pr)
			}

			if opts.Limit > 0 && len(allPRs) >= opts.Limit {
				slog.Info("Reached limit of PRs to fetch", "limit", opts.Limit)
				break
			}
		}
		slog.Info("Fetched PRs", "count", len(response.Search.Nodes), "total", len(allPRs))
		if opts.Limit > 0 && len(allPRs) >= opts.Limit {
			break
		}

		if !response.Search.PageInfo.HasNextPage {
			break
		}

		cursor = &response.Search.PageInfo.EndCursor
	}

	slog.Info("Total PRs fetched", "count", len(allPRs))

	// Calculate metrics for each PR
	metrics := make([]PRMetrics, 0, len(allPRs))
	for _, pr := range allPRs {
		prMetrics := CalculateMetrics(pr, opts.ExcludeWeekends)
		metrics = append(metrics, prMetrics)
	}

	return metrics, nil
}

// ParseTargetUsers parses the comma-separated target users string
func ParseTargetUsers(targetUsers string) []string {
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
