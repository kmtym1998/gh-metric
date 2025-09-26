package github

import (
	"log/slog"
	"strings"
	"time"
)

// CalculateMetrics processes raw PR data and calculates lead time metrics
func CalculateMetrics(pr PullRequest, excludeWeekends bool) PRMetrics {
	metrics := PRMetrics{
		Number:       pr.Number,
		Title:        pr.Title,
		Author:       pr.Author.Login,
		CreatedAt:    pr.CreatedAt,
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

	// Find the earliest review time and first approval time
	var firstReviewTime *time.Time
	var firstApproveTime *time.Time
	reviewers := make(map[string]bool)

	for _, review := range pr.Reviews.Nodes {
		reviewers[review.Author.Login] = true

		if firstReviewTime == nil || review.SubmittedAt.Before(*firstReviewTime) {
			firstReviewTime = &review.SubmittedAt
		}

		if review.State == "APPROVED" && (firstApproveTime == nil || review.SubmittedAt.Before(*firstApproveTime)) {
			firstApproveTime = &review.SubmittedAt
		}
	}

	// Convert reviewers map to comma-separated string
	reviewerList := make([]string, 0, len(reviewers))
	for reviewer := range reviewers {
		reviewerList = append(reviewerList, reviewer)
	}
	metrics.Reviewers = strings.Join(reviewerList, ",")

	// Calculate lead times if we have a review assignment time
	if reviewAssignTime != nil {
		if firstReviewTime != nil {
			metrics.UntilFirstReview = calculateHoursDifference(*reviewAssignTime, *firstReviewTime, excludeWeekends)
		}

		if firstApproveTime != nil {
			metrics.UntilFirstApprove = calculateHoursDifference(*reviewAssignTime, *firstApproveTime, excludeWeekends)
		}

		if !pr.MergedAt.IsZero() {
			metrics.UntilMerge = calculateHoursDifference(*reviewAssignTime, pr.MergedAt, excludeWeekends)
		}
	} else {
		slog.Warn("No reviewer assignment found for PR", "number", pr.Number)
	}

	return metrics
}

// calculateHoursDifference calculates the difference between two times in hours
// If excludeWeekends is true, weekends are excluded from the calculation
func calculateHoursDifference(start, end time.Time, excludeWeekends bool) float64 {
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
