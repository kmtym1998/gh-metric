package github

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
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

// FetchMergedPRs fetches merged pull requests from GitHub API
func (s *Service) FetchMergedPRs(ctx context.Context, opts FetchPROptions) ([]PullRequest, error) {
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
	return allPRs, nil
}
