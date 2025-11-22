package usecase

import (
	"time"
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