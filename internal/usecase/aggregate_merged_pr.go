package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kmtym1998/gh-metric/internal/github"
)

// AggregateMergedPRInput contains the input parameters for the aggregate merged PR use case
type AggregateMergedPRInput struct {
	Owner           string
	Repo            string
	Since           string
	Until           string
	TargetUsers     []string
	ExcludeWeekends bool
	Limit           int
	IncludeBot      bool
}

// AggregateMergedPROutput contains the result of the aggregate merged PR use case
type AggregateMergedPROutput struct {
	Items []PRMetric
}

// PRMetric represents the calculated metrics for a single Pull Request
type PRMetric struct {
	Repository        string    `json:"repository"`
	Number            int       `json:"number"`
	Title             string    `json:"title"`
	Author            string    `json:"author"`
	CreatedAt         time.Time `json:"created_at"`
	CreatedWeek       time.Time `json:"created_week"`  // The week (Monday) when the PR was created
	CreatedMonth      time.Time `json:"created_month"` // The month (1st day) when the PR was created
	MergedAt          time.Time `json:"merged_at"`
	MergedWeek        time.Time `json:"merged_week"`  // The week (Monday) when the PR was merged
	MergedMonth       time.Time `json:"merged_month"` // The month (1st day) when the PR was merged
	URL               string    `json:"url"`
	UntilFirstReview  float64   `json:"until_first_review"`  // hours
	UntilFirstApprove float64   `json:"until_first_approve"` // hours
	UntilMerge        float64   `json:"until_merge"`         // hours
	CommentCount      int       `json:"comment_count"`
	ChangedFiles      int       `json:"changed_files"`
	Additions         int       `json:"additions"`
	Deletions         int       `json:"deletions"`
	ApprovedBy        []string  `json:"approved_by"`       // list of user IDs who approved
	FirstReviewedBy   string    `json:"first_reviewed_by"` // user ID who gave the first review
	FirstApprovedBy   string    `json:"first_approved_by"` // user ID who gave the first approval
}

// GitHubRepository interface defines what we need from GitHub API
type GitHubRepository interface {
	FetchMergedPRs(ctx context.Context, opts github.FetchPROptions) ([]github.PullRequest, error)
}

// AggregateMergedPRUsecase handles the business logic for aggregating merged PR metrics
type AggregateMergedPRUsecase struct {
	githubRepo GitHubRepository
}

// NewAggregateMergedPRUsecase creates a new instance of AggregateMergedPRUsecase
func NewAggregateMergedPRUsecase(githubRepo GitHubRepository) *AggregateMergedPRUsecase {
	return &AggregateMergedPRUsecase{
		githubRepo: githubRepo,
	}
}

// Execute performs the aggregation of merged PR metrics
func (uc *AggregateMergedPRUsecase) Execute(ctx context.Context, req AggregateMergedPRInput) (*AggregateMergedPROutput, error) {
	slog.Info("Starting PR metrics aggregation",
		"owner", req.Owner,
		"repo", req.Repo,
		"since", req.Since,
		"until", req.Until,
		"targetUsers", req.TargetUsers,
		"excludeWeekends", req.ExcludeWeekends,
		"includeBot", req.IncludeBot)

	slog.Debug("input parameters", "request", fmt.Sprintf("%+v", req))

	// Convert request to GitHub API options
	opts := github.FetchPROptions{
		Owner:           req.Owner,
		Repo:            req.Repo,
		Since:           req.Since,
		Until:           req.Until,
		TargetUsers:     req.TargetUsers,
		ExcludeWeekends: req.ExcludeWeekends,
		Limit:           req.Limit,
	}

	slog.Debug("GitHub API options", "opts", fmt.Sprintf("%+v", opts))

	// Fetch PRs from GitHub
	slog.Debug("Fetching PRs from GitHub")
	prs, err := uc.githubRepo.FetchMergedPRs(ctx, opts)
	if err != nil {
		slog.Error("Failed to fetch merged PRs", "error", err)
		return nil, fmt.Errorf("failed to fetch merged PRs: %w", err)
	}

	slog.Info("Fetched PRs from GitHub", "count", len(prs))

	// Calculate metrics for each PR
	slog.Debug("Calculating metrics for PRs", "count", len(prs))
	metrics := make([]PRMetric, 0, len(prs))
	repository := fmt.Sprintf("%s/%s", req.Owner, req.Repo)
	for _, pr := range prs {
		metric := uc.calculatePRMetrics(repository, pr, req.ExcludeWeekends, req.IncludeBot)
		metrics = append(metrics, metric)
	}

	slog.Info("Calculated PR metrics", "count", len(metrics))

	return &AggregateMergedPROutput{
		Items: metrics,
	}, nil
}

// calculatePRMetrics processes raw PR data and calculates lead time metrics
func (uc *AggregateMergedPRUsecase) calculatePRMetrics(repository string, pr github.PullRequest, excludeWeekends bool, includeBot bool) PRMetric {
	// FIXME: Timezone should be configurable
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	jstCreatedAt := pr.CreatedAt.In(jst)
	jstMergedAt := pr.MergedAt.In(jst)
	metric := PRMetric{
		Repository:   repository,
		Number:       pr.Number,
		Title:        pr.Title,
		Author:       pr.Author.Login,
		CreatedAt:    jstCreatedAt,
		CreatedWeek:  jstCreatedAt.AddDate(0, 0, -int(jstCreatedAt.Weekday())+1), // Adjust to Monday
		CreatedMonth: time.Date(jstCreatedAt.Year(), jstCreatedAt.Month(), 1, 0, 0, 0, 0, jst),
		MergedAt:     jstMergedAt,
		MergedWeek:   jstMergedAt.AddDate(0, 0, -int(jstMergedAt.Weekday())+1), // Adjust to Monday
		MergedMonth:  time.Date(jstMergedAt.Year(), jstMergedAt.Month(), 1, 0, 0, 0, 0, jst),
		URL:          pr.URL,
		CommentCount: pr.Comments.TotalCount,
		ChangedFiles: pr.ChangedFiles,
		Additions:    pr.Additions,
		Deletions:    pr.Deletions,
	}

	// Find the earliest reviewer assignment time
	var reviewAssignTime *time.Time
	for _, item := range pr.TimelineItems.Nodes {
		if item.Typename == "ReviewRequestedEvent" {
			if reviewAssignTime == nil || item.CreatedAt.Before(*reviewAssignTime) {
				reviewAssignTime = &item.CreatedAt
			}
		}
	}

	// Find the earliest review time, first approval time, and collect review information
	var firstReviewTime *time.Time
	var firstApproveTime *time.Time
	var firstReviewedBy string
	var firstApprovedBy string
	approvedByMap := make(map[string]bool) // Use map to avoid duplicates

	for _, review := range pr.Reviews.Nodes {
		author := review.Author.Login

		// Skip bot reviews if includeBot is false
		if !includeBot && IsBot(author) {
			continue
		}

		// Skip author's own reviews
		if author == pr.Author.Login {
			continue
		}

		// Track first review (any review type)
		if firstReviewTime == nil || review.SubmittedAt.Before(*firstReviewTime) {
			firstReviewTime = &review.SubmittedAt
			firstReviewedBy = author
		}

		// Track approvals
		if review.State == "APPROVED" || review.State == "DISMISSED" {
			approvedByMap[author] = true

			// Track first approval
			if firstApproveTime == nil || review.SubmittedAt.Before(*firstApproveTime) {
				firstApproveTime = &review.SubmittedAt
				firstApprovedBy = author
			}
		}
	}

	// Convert approvers map to slice
	approvedBy := make([]string, 0, len(approvedByMap))
	for approver := range approvedByMap {
		approvedBy = append(approvedBy, approver)
	}

	// Set the review information
	metric.ApprovedBy = approvedBy
	metric.FirstReviewedBy = firstReviewedBy
	metric.FirstApprovedBy = firstApprovedBy

	// Calculate lead times if we have a review assignment time
	if reviewAssignTime != nil {
		if firstReviewTime != nil {
			metric.UntilFirstReview = uc.calculateHoursDifference(*reviewAssignTime, *firstReviewTime, excludeWeekends)
		}

		if firstApproveTime != nil {
			metric.UntilFirstApprove = uc.calculateHoursDifference(*reviewAssignTime, *firstApproveTime, excludeWeekends)
		}

		if !pr.MergedAt.IsZero() {
			metric.UntilMerge = uc.calculateHoursDifference(*reviewAssignTime, pr.MergedAt, excludeWeekends)
		}

	} else {
		slog.Warn("No reviewer assignment found for PR", "number", pr.Number)
	}

	return metric
}

// calculateHoursDifference calculates the difference between two times in hours
// If excludeWeekends is true, weekends are excluded from the calculation
func (uc *AggregateMergedPRUsecase) calculateHoursDifference(start, end time.Time, excludeWeekends bool) float64 {
	if !excludeWeekends {
		// Simple calculation without weekend exclusion
		duration := end.Sub(start)
		return duration.Hours()
	}

	// Calculate business hours (excluding weekends)
	totalHours := 0.0
	current := start

	for current.Before(end) {
		// Skip weekends (Saturday = 6, Sunday = 0)
		if current.Weekday() != time.Saturday && current.Weekday() != time.Sunday {
			// Calculate how much time to add
			nextDay := current.Add(24 * time.Hour)
			nextDay = time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, nextDay.Location())

			var endTime time.Time
			if end.Before(nextDay) {
				endTime = end
			} else {
				endTime = nextDay
			}

			if endTime.After(current) {
				totalHours += endTime.Sub(current).Hours()
			}
		}

		// Move to next day
		current = time.Date(current.Year(), current.Month(), current.Day()+1, 0, 0, 0, 0, current.Location())
	}

	return totalHours
}
