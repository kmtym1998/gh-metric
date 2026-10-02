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

// FetchMergedPRs fetches merged pull requests from GitHub API.
//
// The GitHub Search API returns at most 1000 results per query regardless of
// pagination. To fetch beyond that, the search is re-issued with a
// `created:>=<createdAt of the last PR>` condition, which works as a cursor
// because results are sorted by creation time in ascending order.
func (s *Service) FetchMergedPRs(ctx context.Context, opts FetchPROptions) ([]PullRequest, error) {
	slog.Debug("Starting FetchMergedPRs", "opts", fmt.Sprintf("%+v", opts))

	var allPRs []PullRequest
	seen := make(map[int]struct{})
	var createdSince *time.Time
	pageCount := 0

	limitReached := func() bool {
		return opts.Limit > 0 && len(allPRs) >= opts.Limit
	}

	for {
		searchQuery := buildMergedPRSearchQuery(opts, createdSince)
		slog.Info("Executing search query", "query", searchQuery)

		fetchedInQuery := 0
		issueCount := 0
		var lastCreatedAt time.Time
		var cursor *string

		// Paginate through the results of a single search query
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
				"issueCount", response.Search.IssueCount,
				"hasNextPage", response.Search.PageInfo.HasNextPage,
				"endCursor", response.Search.PageInfo.EndCursor,
			)

			issueCount = response.Search.IssueCount
			fetchedInQuery += len(response.Search.Nodes)
			if n := len(response.Search.Nodes); n > 0 {
				lastCreatedAt = response.Search.Nodes[n-1].CreatedAt
			}

			for _, pr := range response.Search.Nodes {
				// The created:>= cursor is inclusive, so PRs at the boundary can appear twice
				if _, ok := seen[pr.Number]; ok {
					continue
				}
				seen[pr.Number] = struct{}{}

				if len(opts.TargetUsers) > 0 && !slices.Contains(opts.TargetUsers, pr.Author.Login) {
					slog.Debug("Skipped PR due to target user filter",
						"prNumber", pr.Number,
						"author", pr.Author.Login,
						"targetUsers", opts.TargetUsers,
					)
					continue
				}

				allPRs = append(allPRs, pr)
				if limitReached() {
					slog.Info("Reached limit of PRs to fetch", "limit", opts.Limit)
					break
				}
			}
			slog.Info("Fetched PRs", "count", len(response.Search.Nodes), "total", len(allPRs))

			if limitReached() || !response.Search.PageInfo.HasNextPage {
				break
			}
			cursor = &response.Search.PageInfo.EndCursor
		}

		// Pagination ends when either the limit is reached or the search query is exhausted.
		// A query is exhausted when every matching PR has been fetched; otherwise the result cap was hit.
		if limitReached() || fetchedInQuery >= issueCount || fetchedInQuery == 0 {
			break
		}

		if createdSince != nil && !lastCreatedAt.After(*createdSince) {
			return nil, fmt.Errorf("cannot advance search cursor: more than %d PRs created at %s", fetchedInQuery, lastCreatedAt)
		}
		slog.Info("Search result cap reached, continuing with created cursor",
			"fetched", fetchedInQuery,
			"issueCount", issueCount,
			"createdSince", lastCreatedAt,
		)
		createdSince = &lastCreatedAt
	}

	slog.Info("Total PRs fetched", "count", len(allPRs), "pages", pageCount)
	return allPRs, nil
}

func buildMergedPRSearchQuery(opts FetchPROptions, createdSince *time.Time) string {
	queryBuilder := NewSearchQueryBuilder()
	queryBuilder.AddRepository(opts.Owner, opts.Repo)
	queryBuilder.AddDateRange(opts.Since, opts.Until)
	if createdSince != nil {
		queryBuilder.AddCreatedSince(*createdSince)
	}
	queryBuilder.AddSort()
	return queryBuilder.Build()
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
