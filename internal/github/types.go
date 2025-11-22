package github

import "time"

// GraphQL query response structures

type SearchResponse struct {
	Search struct {
		PageInfo struct {
			StartCursor string `json:"startCursor"`
			EndCursor   string `json:"endCursor"`
			HasNextPage bool   `json:"hasNextPage"`
		} `json:"pageInfo"`
		Nodes []PullRequest `json:"nodes"`
	} `json:"search"`
}

type PullRequest struct {
	Number       int       `json:"number"`
	Title        string    `json:"title"`
	CreatedAt    time.Time `json:"createdAt"`
	MergedAt     time.Time `json:"mergedAt"`
	ClosedAt     time.Time `json:"closedAt"`
	URL          string    `json:"url"`
	Additions    int       `json:"additions"`
	Deletions    int       `json:"deletions"`
	ChangedFiles int       `json:"changedFiles"`
	Comments     struct {
		TotalCount int `json:"totalCount"`
	} `json:"comments"`
	TimelineItems struct {
		PageInfo struct {
			StartCursor string `json:"startCursor"`
			EndCursor   string `json:"endCursor"`
			HasNextPage bool   `json:"hasNextPage"`
		} `json:"pageInfo"`
		Nodes []TimelineItem `json:"nodes"`
	} `json:"timelineItems"`
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
	Reviews struct {
		Nodes []Review `json:"nodes"`
	} `json:"reviews"`
}

type TimelineItem struct {
	Typename          string            `json:"__typename"`
	CreatedAt         time.Time         `json:"createdAt"`
	RequestedReviewer RequestedReviewer `json:"requestedReviewer"`
}

type RequestedReviewer struct {
	Typename string `json:"__typename"`
	Login    string `json:"login"`
	Name     string `json:"name"`
	Members  struct {
		Nodes []struct {
			Login string `json:"login"`
		} `json:"nodes"`
	} `json:"members"`
}

type Review struct {
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
	State       string    `json:"state"`
	SubmittedAt time.Time `json:"submittedAt"`
}

// GraphQL query variables
type QueryVariables struct {
	Query  string  `json:"query"`
	Cursor *string `json:"cursor,omitempty"`
}

// listIssuesResponse is the response structure for listIssueAndProjectFieldsQuery
type listIssuesResponse struct {
	Repository struct {
		Issues struct {
			TotalCount int `json:"totalCount"`
			PageInfo   struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
			Nodes []issueNode `json:"nodes"`
		} `json:"issues"`
	} `json:"repository"`
}

// issueNode represents a single issue with project fields
type issueNode struct {
	ID        string    `json:"id"`
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	ClosedAt  time.Time `json:"closedAt"`
	Author    struct {
		Login string `json:"login"`
	} `json:"author"`
	Assignees struct {
		Nodes []struct {
			Login string `json:"login"`
		} `json:"nodes"`
	} `json:"assignees"`
	Labels struct {
		Nodes []struct {
			Name string `json:"name"`
		} `json:"nodes"`
	} `json:"labels"`
	Milestone struct {
		Title  string `json:"title"`
		Number int    `json:"number"`
	} `json:"milestone"`
	ProjectItems struct {
		TotalCount int                `json:"totalCount"`
		Nodes      []projectItemNode `json:"nodes"`
	} `json:"projectItems"`
}

// projectItemNode represents a project item with field values
type projectItemNode struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Project   struct {
		ID     string `json:"id"`
		Number int    `json:"number"`
		Title  string `json:"title"`
	} `json:"project"`
	FieldValues struct {
		Nodes []interface{} `json:"nodes"`
	} `json:"fieldValues"`
}

// issueQueryVariables represents the variables for listIssueAndProjectFieldsQuery
type issueQueryVariables struct {
	Owner    string                 `json:"owner"`
	Repo     string                 `json:"repo"`
	First    int                    `json:"first,omitempty"`
	After    *string                `json:"after,omitempty"`
	States   []string               `json:"states,omitempty"`
	Labels   []string               `json:"labels,omitempty"`
	FilterBy map[string]interface{} `json:"filterBy,omitempty"`
	OrderBy  map[string]string      `json:"orderBy,omitempty"`
}
