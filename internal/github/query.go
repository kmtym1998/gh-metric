package github

import (
	"fmt"
	"strings"
)

// The GraphQL query as specified in the instructions
const pullRequestQuery = `
query PullRequestLeadTime($query: String!, $cursor: String) {
  search(query: $query, type: ISSUE, first: 50, after: $cursor) {
    pageInfo {
      startCursor
      endCursor
      hasNextPage
    }
    nodes {
      ... on PullRequest {
        number
        title
        createdAt
        mergedAt
        closedAt
        url
        additions
        deletions
        changedFiles
        comments {
          totalCount
        }
        timelineItems(first: 30, itemTypes: [REVIEW_REQUESTED_EVENT]) {
          pageInfo {
            startCursor
            endCursor
            hasNextPage
          }
          nodes {
            __typename
            ... on ReviewRequestedEvent {
              __typename
              createdAt
              requestedReviewer {
                ... on Bot {
                  __typename
                  login
                }
                ... on Mannequin {
                  __typename
                  login
                }
                ... on Team {
                  __typename
                  name
                  members {
                    nodes {
                      login
                    }
                  }
                }
                ... on User {
                  __typename
                  login
                }
              }
            }
          }
        }
        author {
          login
        }
        reviews(first: 50) {
          nodes {
            author {
              login
            }
            state
            submittedAt
          }
        }
      }
    }
  }
}
`

// SearchQueryBuilder builds a GitHub search query string
type SearchQueryBuilder struct {
	parts []string
}

// NewSearchQueryBuilder creates a new search query builder
func NewSearchQueryBuilder() *SearchQueryBuilder {
	return &SearchQueryBuilder{
		parts: []string{"is:merged"},
	}
}

// AddDateRange adds date filters to the query
func (b *SearchQueryBuilder) AddDateRange(since, until string) *SearchQueryBuilder {
	if since != "" {
		b.parts = append(b.parts, fmt.Sprintf("merged:>=%s", since))
	}
	if until != "" {
		b.parts = append(b.parts, fmt.Sprintf("merged:<=%s", until))
	}
	return b
}

// AddAuthors adds author filters to the query
func (b *SearchQueryBuilder) AddAuthors(authors []string) *SearchQueryBuilder {
	if len(authors) > 0 {
		authorQueries := make([]string, len(authors))
		for i, author := range authors {
			authorQueries[i] = fmt.Sprintf("author:%s", strings.TrimSpace(author))
		}
		b.parts = append(b.parts, fmt.Sprintf("(%s)", strings.Join(authorQueries, " OR ")))
	}
	return b
}

// AddRepository adds repository filter to the query
func (b *SearchQueryBuilder) AddRepository(owner, repo string) *SearchQueryBuilder {
	if owner != "" && repo != "" {
		b.parts = append(b.parts, fmt.Sprintf("repo:%s/%s", owner, repo))
	}
	return b
}

// Build creates the final search query string
func (b *SearchQueryBuilder) Build() string {
	return strings.Join(b.parts, " ")
}
