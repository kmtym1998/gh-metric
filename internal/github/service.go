package github

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/samber/lo"
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
	slog.Debug("Starting FetchMergedPRs", "opts", fmt.Sprintf("%+v", opts))

	// Build the search query
	queryBuilder := NewSearchQueryBuilder()
	queryBuilder.AddRepository(opts.Owner, opts.Repo)
	queryBuilder.AddDateRange(opts.Since, opts.Until)
	queryBuilder.AddSort()

	searchQuery := queryBuilder.Build()
	slog.Info("Executing search query", "query", searchQuery)
	slog.Debug("Query builder configuration",
		"owner", opts.Owner,
		"repo", opts.Repo,
		"since", opts.Since,
		"until", opts.Until,
		"targetUsers", opts.TargetUsers,
		"limit", opts.Limit,
	)

	var allPRs []PullRequest
	var cursor *string
	pageCount := 0

	// Paginate through all results
	for {
		pageCount++
		slog.Debug("Fetching page", "pageNumber", pageCount, "cursor", lo.FromPtr(cursor))

		variables := QueryVariables{
			Query:  searchQuery,
			Cursor: cursor,
		}

		var response SearchResponse
		if err := s.client.ExecuteQuery(pullRequestQuery, variables, &response); err != nil {
			slog.Error("GraphQL query failed", "error", err, "variables", fmt.Sprintf("%+v", variables))
			return nil, fmt.Errorf("failed to execute GraphQL query: %w", err)
		}

		slog.Debug("GraphQL response received",
			"searchResultsCount", len(response.Search.Nodes),
			"hasNextPage", response.Search.PageInfo.HasNextPage,
			"endCursor", response.Search.PageInfo.EndCursor,
		)

		for _, pr := range response.Search.Nodes {
			if slices.Contains(opts.TargetUsers, pr.Author.Login) || len(opts.TargetUsers) == 0 {
				allPRs = append(allPRs, pr)
			} else {
				slog.Debug("Skipped PR due to target user filter",
					"prNumber", pr.Number,
					"author", pr.Author.Login,
					"targetUsers", opts.TargetUsers,
				)
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
			slog.Debug("No more pages available")
			break
		}

		cursor = &response.Search.PageInfo.EndCursor
	}

	slog.Info("Total PRs fetched", "count", len(allPRs), "pages", pageCount)
	return allPRs, nil
}

// FetchIssuesOptions contains options for fetching issues
type FetchIssuesOptions struct {
	Owner  string
	Repo   string
	Since  string   // Date in YYYY-MM-DD format
	Until  string   // Date in YYYY-MM-DD format
	State  string   // all, open, closed
	Labels []string // Optional label filters
	Limit  int
}

// FetchIssues fetches issues with project fields from GitHub API
func (s *Service) FetchIssues(ctx context.Context, opts FetchIssuesOptions) ([]IssueNode, error) {
	slog.Debug("Starting FetchIssues", "opts", fmt.Sprintf("%+v", opts))

	// Map state string to GraphQL IssueState array
	var states []string
	switch opts.State {
	case "open":
		states = []string{"OPEN"}
	case "closed":
		states = []string{"CLOSED"}
	case "all", "":
		states = []string{"OPEN", "CLOSED"}
	default:
		return nil, fmt.Errorf("invalid state: %s (must be 'open', 'closed', or 'all')", opts.State)
	}

	// Build filter for date range if provided
	filterBy := make(map[string]interface{})
	if opts.Since != "" {
		filterBy["since"] = opts.Since + "T00:00:00Z"
	}

	// Build order by for consistent ordering
	orderBy := map[string]string{
		"field":     "CREATED_AT",
		"direction": "DESC",
	}

	var allIssues []IssueNode
	var cursor *string
	pageCount := 0
	first := 100 // Max per page for GraphQL

	// Paginate through all results
	for {
		pageCount++
		slog.Debug("Fetching page", "pageNumber", pageCount, "cursor", lo.FromPtr(cursor))

		variables := map[string]interface{}{
			"owner":    opts.Owner,
			"repo":     opts.Repo,
			"first":    first,
			"after":    cursor,
			"states":   states,
			"labels":   opts.Labels,
			"filterBy": filterBy,
			"orderBy":  orderBy,
		}

		var response listIssuesResponse
		if err := s.client.ExecuteQuery(listIssueAndProjectFieldsQuery, variables, &response); err != nil {
			slog.Error("GraphQL query failed", "error", err, "variables", fmt.Sprintf("%+v", variables))
			return nil, fmt.Errorf("failed to execute GraphQL query: %w", err)
		}

		slog.Debug("GraphQL response received",
			"issuesCount", len(response.Repository.Issues.Nodes),
			"totalCount", response.Repository.Issues.TotalCount,
			"hasNextPage", response.Repository.Issues.PageInfo.HasNextPage,
			"endCursor", response.Repository.Issues.PageInfo.EndCursor,
		)

		// Filter issues by date if Until is specified (since GraphQL doesn't support until filter)
		for _, issue := range response.Repository.Issues.Nodes {
			if opts.Until != "" {
				untilDate := opts.Until + "T23:59:59Z"
				if issue.CreatedAt.Format(time.RFC3339) > untilDate {
					continue
				}
			}
			allIssues = append(allIssues, issue)

			if opts.Limit > 0 && len(allIssues) >= opts.Limit {
				slog.Info("Reached limit of issues to fetch", "limit", opts.Limit)
				break
			}
		}

		slog.Info("Fetched issues", "count", len(response.Repository.Issues.Nodes), "total", len(allIssues))

		if opts.Limit > 0 && len(allIssues) >= opts.Limit {
			break
		}

		if !response.Repository.Issues.PageInfo.HasNextPage {
			slog.Debug("No more pages available")
			break
		}

		cursor = &response.Repository.Issues.PageInfo.EndCursor
	}

	slog.Info("Total issues fetched", "count", len(allIssues), "pages", pageCount)
	return allIssues, nil
}
