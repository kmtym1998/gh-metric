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
