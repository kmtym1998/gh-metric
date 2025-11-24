package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kmtym1998/gh-metric/internal/github"
)

// AggregateIssuesInput contains the input parameters for the aggregate issues use case
type AggregateIssuesInput struct {
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
	Since string `json:"since"` // Date in YYYY-MM-DD format
	Until string `json:"until"` // Date in YYYY-MM-DD format
	State string `json:"state"` // all, open, closed
	Limit int    `json:"limit"`
}

// AggregateIssuesOutput contains the result of the aggregate issues use case
type AggregateIssuesOutput struct {
	Repository string        `json:"repository"`
	Period     Period        `json:"period"`
	TotalCount int           `json:"total_count"`
	Items      []IssueMetric `json:"items"`
}

// Period represents the time period for the aggregation
type Period struct {
	Since string `json:"since"`
	Until string `json:"until"`
}

// IssueMetric represents the aggregated metrics for a single issue
type IssueMetric struct {
	Number           int                  `json:"number"`
	Title            string               `json:"title"`
	State            string               `json:"state"`
	Author           string               `json:"author"`
	Assignees        []string             `json:"assignees"`
	Labels           []string             `json:"labels"`
	MilestoneTitle   string               `json:"milestone_title,omitempty"`
	MilestoneNumber  int                  `json:"milestone_number,omitempty"`
	CreatedAt        time.Time            `json:"created_at"`
	CreatedWeek      time.Time            `json:"created_week"`  // The week (Monday) when the issue was created
	CreatedMonth     time.Time            `json:"created_month"` // The month (1st day) when the issue was created
	UpdatedAt        time.Time            `json:"updated_at"`
	ClosedAt         *time.Time           `json:"closed_at,omitempty"`
	ClosedWeek       *time.Time           `json:"closed_week,omitempty"`  // The week (Monday) when the issue was closed
	ClosedMonth      *time.Time           `json:"closed_month,omitempty"` // The month (1st day) when the issue was closed
	ProjectFieldValues []ProjectFieldValue `json:"project_field_values,omitempty"`
}

// ProjectFieldValue represents a single project field value
type ProjectFieldValue struct {
	ProjectID     string      `json:"project_id"`
	ProjectNumber int         `json:"project_number"`
	ProjectTitle  string      `json:"project_title"`
	FieldName     string      `json:"field_name"`
	FieldType     string      `json:"field_type"`     // text, number, single_select, date, iteration, milestone, label, user, pull_request, reviewer
	Value         interface{} `json:"value"`          // The actual value (varies by field type)
	UpdatedAt     time.Time   `json:"updated_at"`
}

// Repository represents the repository information
type Repository struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

// GitHubIssueRepository interface defines what we need from GitHub API for issues
type GitHubIssueRepository interface {
	FetchIssues(ctx context.Context, opts github.FetchIssuesOptions) ([]github.IssueNode, error)
}

// AggregateIssuesUsecase handles the business logic for aggregating issue metrics
type AggregateIssuesUsecase struct {
	githubRepo *github.Service
}

// NewAggregateIssuesUsecase creates a new instance of AggregateIssuesUsecase
func NewAggregateIssuesUsecase(githubRepo *github.Service) *AggregateIssuesUsecase {
	return &AggregateIssuesUsecase{
		githubRepo: githubRepo,
	}
}

// Execute performs the aggregation of issue metrics
func (uc *AggregateIssuesUsecase) Execute(ctx context.Context, req AggregateIssuesInput) (*AggregateIssuesOutput, error) {
	slog.Info("Starting issue metrics aggregation",
		"owner", req.Owner,
		"repo", req.Repo,
		"since", req.Since,
		"until", req.Until,
		"state", req.State,
		"limit", req.Limit)

	slog.Debug("input parameters", "request", fmt.Sprintf("%+v", req))

	// Convert request to GitHub API options
	opts := github.FetchIssuesOptions{
		Owner: req.Owner,
		Repo:  req.Repo,
		Since: req.Since,
		Until: req.Until,
		State: req.State,
		Limit: req.Limit,
	}

	slog.Debug("GitHub API options", "opts", fmt.Sprintf("%+v", opts))

	// Fetch issues from GitHub
	slog.Debug("Fetching issues from GitHub")
	issues, err := uc.githubRepo.FetchIssues(ctx, opts)
	if err != nil {
		slog.Error("Failed to fetch issues", "error", err)
		return nil, fmt.Errorf("failed to fetch issues: %w", err)
	}

	slog.Info("Fetched issues from GitHub", "count", len(issues))

	// Calculate metrics for each issue
	slog.Debug("Calculating metrics for issues", "count", len(issues))
	metrics := make([]IssueMetric, 0, len(issues))
	repository := fmt.Sprintf("%s/%s", req.Owner, req.Repo)

	for _, issue := range issues {
		metric := uc.calculateIssueMetrics(issue)
		metrics = append(metrics, metric)
	}

	slog.Info("Calculated issue metrics", "count", len(metrics))

	return &AggregateIssuesOutput{
		Repository: repository,
		Period: Period{
			Since: req.Since,
			Until: req.Until,
		},
		TotalCount: len(metrics),
		Items:      metrics,
	}, nil
}

// calculateIssueMetrics processes raw issue data and calculates metrics
func (uc *AggregateIssuesUsecase) calculateIssueMetrics(issue github.IssueNode) IssueMetric {
	// FIXME: Timezone should be configurable
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	jstCreatedAt := issue.CreatedAt.In(jst)
	jstUpdatedAt := issue.UpdatedAt.In(jst)

	metric := IssueMetric{
		Number:          issue.Number,
		Title:           issue.Title,
		State:           issue.State,
		Author:          issue.Author.Login,
		CreatedAt:       jstCreatedAt,
		CreatedWeek:     getWeekStart(jstCreatedAt),
		CreatedMonth:    getMonthStart(jstCreatedAt),
		UpdatedAt:       jstUpdatedAt,
		MilestoneTitle:  issue.Milestone.Title,
		MilestoneNumber: issue.Milestone.Number,
	}

	// Process assignees
	for _, assignee := range issue.Assignees.Nodes {
		metric.Assignees = append(metric.Assignees, assignee.Login)
	}

	// Process labels
	for _, label := range issue.Labels.Nodes {
		metric.Labels = append(metric.Labels, label.Name)
	}

	// Process closed date if closed
	if !issue.ClosedAt.IsZero() {
		jstClosedAt := issue.ClosedAt.In(jst)
		metric.ClosedAt = &jstClosedAt
		closedWeek := getWeekStart(jstClosedAt)
		metric.ClosedWeek = &closedWeek
		closedMonth := getMonthStart(jstClosedAt)
		metric.ClosedMonth = &closedMonth
	}

	// Extract project field values
	projectFields := uc.extractProjectFieldValues(issue)
	metric.ProjectFieldValues = projectFields

	return metric
}

// getWeekStart returns the Monday of the week for the given time
func getWeekStart(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday -> 7
	}
	// Calculate days to subtract to get to Monday (weekday 1)
	daysToSubtract := weekday - 1
	return time.Date(t.Year(), t.Month(), t.Day()-daysToSubtract, 0, 0, 0, 0, t.Location())
}

// getMonthStart returns the first day of the month for the given time
func getMonthStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// extractProjectFieldValues extracts and processes project field values from an issue
func (uc *AggregateIssuesUsecase) extractProjectFieldValues(issue github.IssueNode) []ProjectFieldValue {
	var fieldValues []ProjectFieldValue

	// Iterate through all project items
	for _, projectItem := range issue.ProjectItems.Nodes {
		projectID := projectItem.Project.ID
		projectNumber := projectItem.Project.Number
		projectTitle := projectItem.Project.Title

		// Process each field value
		for _, fieldValue := range projectItem.FieldValues.Nodes {
			if fieldValue == nil {
				continue
			}

			// Type assert to map to access the fields
			fieldMap, ok := fieldValue.(map[string]interface{})
			if !ok {
				slog.Debug("Failed to assert field value to map", "fieldValue", fieldValue)
				continue
			}

			// Get the type name to determine how to process the field
			typeName, ok := fieldMap["__typename"].(string)
			if !ok {
				continue
			}

			pfv := ProjectFieldValue{
				ProjectID:     projectID,
				ProjectNumber: projectNumber,
				ProjectTitle:  projectTitle,
				UpdatedAt:     projectItem.UpdatedAt,
			}

			// Process based on field type
			switch typeName {
			case "ProjectV2ItemFieldTextValue":
				pfv.FieldType = "text"
				if text, ok := fieldMap["text"].(string); ok {
					pfv.Value = text
				}
				if field, ok := fieldMap["field"].(map[string]interface{}); ok {
					if name, ok := field["name"].(string); ok {
						pfv.FieldName = name
					}
				}

			case "ProjectV2ItemFieldNumberValue":
				pfv.FieldType = "number"
				if number, ok := fieldMap["number"].(float64); ok {
					pfv.Value = number
				}
				if field, ok := fieldMap["field"].(map[string]interface{}); ok {
					if name, ok := field["name"].(string); ok {
						pfv.FieldName = name
					}
				}

			case "ProjectV2ItemFieldSingleSelectValue":
				pfv.FieldType = "single_select"
				if name, ok := fieldMap["name"].(string); ok {
					pfv.Value = name
				}
				if field, ok := fieldMap["field"].(map[string]interface{}); ok {
					if name, ok := field["name"].(string); ok {
						pfv.FieldName = name
					}
				}

			case "ProjectV2ItemFieldDateValue":
				pfv.FieldType = "date"
				if date, ok := fieldMap["date"].(string); ok {
					pfv.Value = date
				}
				if field, ok := fieldMap["field"].(map[string]interface{}); ok {
					if name, ok := field["name"].(string); ok {
						pfv.FieldName = name
					}
				}

			case "ProjectV2ItemFieldIterationValue":
				pfv.FieldType = "iteration"
				if title, ok := fieldMap["title"].(string); ok {
					iterationInfo := map[string]interface{}{
						"title": title,
					}
					if startDate, ok := fieldMap["startDate"].(string); ok {
						iterationInfo["startDate"] = startDate
					}
					if duration, ok := fieldMap["duration"].(float64); ok {
						iterationInfo["duration"] = duration
					}
					pfv.Value = iterationInfo
				}
				if field, ok := fieldMap["field"].(map[string]interface{}); ok {
					if name, ok := field["name"].(string); ok {
						pfv.FieldName = name
					}
				}

			case "ProjectV2ItemFieldMilestoneValue":
				pfv.FieldType = "milestone"
				if milestone, ok := fieldMap["milestone"].(map[string]interface{}); ok {
					milestoneInfo := map[string]interface{}{}
					if title, ok := milestone["title"].(string); ok {
						milestoneInfo["title"] = title
					}
					if number, ok := milestone["number"].(float64); ok {
						milestoneInfo["number"] = int(number)
					}
					if state, ok := milestone["state"].(string); ok {
						milestoneInfo["state"] = state
					}
					if dueOn, ok := milestone["dueOn"].(string); ok {
						milestoneInfo["dueOn"] = dueOn
					}
					pfv.Value = milestoneInfo
				}
				if field, ok := fieldMap["field"].(map[string]interface{}); ok {
					if name, ok := field["name"].(string); ok {
						pfv.FieldName = name
					}
				}

			case "ProjectV2ItemFieldLabelValue":
				pfv.FieldType = "label"
				var labelNames []string
				if labels, ok := fieldMap["labels"].(map[string]interface{}); ok {
					if nodes, ok := labels["nodes"].([]interface{}); ok {
						for _, node := range nodes {
							if label, ok := node.(map[string]interface{}); ok {
								if name, ok := label["name"].(string); ok {
									labelNames = append(labelNames, name)
								}
							}
						}
					}
				}
				pfv.Value = labelNames
				if field, ok := fieldMap["field"].(map[string]interface{}); ok {
					if name, ok := field["name"].(string); ok {
						pfv.FieldName = name
					}
				}

			case "ProjectV2ItemFieldUserValue":
				pfv.FieldType = "user"
				var userLogins []string
				if users, ok := fieldMap["users"].(map[string]interface{}); ok {
					if nodes, ok := users["nodes"].([]interface{}); ok {
						for _, node := range nodes {
							if user, ok := node.(map[string]interface{}); ok {
								if login, ok := user["login"].(string); ok {
									userLogins = append(userLogins, login)
								}
							}
						}
					}
				}
				pfv.Value = userLogins
				if field, ok := fieldMap["field"].(map[string]interface{}); ok {
					if name, ok := field["name"].(string); ok {
						pfv.FieldName = name
					}
				}

			case "ProjectV2ItemFieldPullRequestValue":
				pfv.FieldType = "pull_request"
				var prInfo []map[string]interface{}
				if prs, ok := fieldMap["pullRequests"].(map[string]interface{}); ok {
					if nodes, ok := prs["nodes"].([]interface{}); ok {
						for _, node := range nodes {
							if pr, ok := node.(map[string]interface{}); ok {
								prData := map[string]interface{}{}
								if number, ok := pr["number"].(float64); ok {
									prData["number"] = int(number)
								}
								if title, ok := pr["title"].(string); ok {
									prData["title"] = title
								}
								if url, ok := pr["url"].(string); ok {
									prData["url"] = url
								}
								if state, ok := pr["state"].(string); ok {
									prData["state"] = state
								}
								prInfo = append(prInfo, prData)
							}
						}
					}
				}
				pfv.Value = prInfo
				if field, ok := fieldMap["field"].(map[string]interface{}); ok {
					if name, ok := field["name"].(string); ok {
						pfv.FieldName = name
					}
				}

			case "ProjectV2ItemFieldReviewerValue":
				pfv.FieldType = "reviewer"
				var reviewerInfo []string
				if reviewers, ok := fieldMap["reviewers"].(map[string]interface{}); ok {
					if nodes, ok := reviewers["nodes"].([]interface{}); ok {
						for _, node := range nodes {
							if reviewer, ok := node.(map[string]interface{}); ok {
								// Check if it's a user or team
								if login, ok := reviewer["login"].(string); ok {
									reviewerInfo = append(reviewerInfo, login)
								} else if name, ok := reviewer["name"].(string); ok {
									reviewerInfo = append(reviewerInfo, name)
								}
							}
						}
					}
				}
				pfv.Value = reviewerInfo
				if field, ok := fieldMap["field"].(map[string]interface{}); ok {
					if name, ok := field["name"].(string); ok {
						pfv.FieldName = name
					}
				}

			default:
				// Unknown field type, skip
				slog.Debug("Unknown project field type", "typename", typeName)
				continue
			}

			// Only add if we have a field name and value
			if pfv.FieldName != "" && pfv.Value != nil {
				fieldValues = append(fieldValues, pfv)
			}
		}
	}

	return fieldValues
}