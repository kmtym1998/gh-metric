package github

import (
	"fmt"
	"strings"
	"time"
)

// The GraphQL query as specified in the instructions
const pullRequestQuery = `
query PullRequestLeadTime($query: String!, $cursor: String) {
  search(query: $query, type: ISSUE, first: 50, after: $cursor) {
    issueCount
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

const listIssueAndProjectFieldsQuery = `
query GetIssuesProjectFields(
  $owner: String!
  $repo: String!
  $first: Int = 20
  $after: String
  $states: [IssueState!]
  $labels: [String!]
  $filterBy: IssueFilters
  $orderBy: IssueOrder
) {
  repository(owner: $owner, name: $repo) {
    issues(
      first: $first
      after: $after
      states: $states
      labels: $labels
      filterBy: $filterBy
      orderBy: $orderBy
    ) {
      totalCount
      pageInfo {
        hasNextPage
        endCursor
      }
      nodes {
        id
        number
        title
        state
        createdAt
        updatedAt
        closedAt
        author {
          login
        }
        assignees(first: 5) {
          nodes {
            login
          }
        }
        labels(first: 5) {
          nodes {
            name
          }
        }
        milestone {
          title
          number
        }

        # Issue に紐づくプロジェクトアイテム
        projectItems(first: 10) {
          totalCount
          nodes {
            id
            createdAt
            updatedAt

            # プロジェクト情報
            project {
              id
              number
              title
            }

            # すべてのフィールド値を取得
            fieldValues(first: 20) {
              nodes {
                __typename

                # テキストフィールド
                ... on ProjectV2ItemFieldTextValue {
                  id
                  text
                  createdAt
                  updatedAt
                  field {
                    ... on ProjectV2Field {
                      id
                      name
                    }
                  }
                }

                # 数値フィールド
                ... on ProjectV2ItemFieldNumberValue {
                  id
                  number
                  createdAt
                  updatedAt
                  field {
                    ... on ProjectV2Field {
                      id
                      name
                    }
                  }
                }

                # 単一選択フィールド
                ... on ProjectV2ItemFieldSingleSelectValue {
                  id
                  name
                  nameHTML
                  color
                  description
                  createdAt
                  updatedAt
                  field {
                    ... on ProjectV2SingleSelectField {
                      id
                      name
                    }
                  }
                }

                # 日付フィールド
                ... on ProjectV2ItemFieldDateValue {
                  id
                  date
                  createdAt
                  updatedAt
                  field {
                    ... on ProjectV2Field {
                      id
                      name
                    }
                  }
                }

                # イテレーションフィールド
                ... on ProjectV2ItemFieldIterationValue {
                  id
                  title
                  startDate
                  duration
                  createdAt
                  updatedAt
                  field {
                    ... on ProjectV2IterationField {
                      id
                      name
                      configuration {
                        duration
                        startDay
                        iterations {
                          id
                          title
                          startDate
                          duration
                        }
                      }
                    }
                  }
                }

                # マイルストーンフィールド
                ... on ProjectV2ItemFieldMilestoneValue {
                  milestone {
                    id
                    title
                    number
                    dueOn
                    state
                  }
                  field {
                    ... on ProjectV2Field {
                      id
                      name
                    }
                  }
                }

                # ラベルフィールド
                ... on ProjectV2ItemFieldLabelValue {
                  labels(first: 10) {
                    nodes {
                      id
                      name
                      color
                      description
                    }
                  }
                  field {
                    ... on ProjectV2Field {
                      id
                      name
                    }
                  }
                }

                # ユーザーフィールド
                ... on ProjectV2ItemFieldUserValue {
                  users(first: 10) {
                    nodes {
                      id
                      login
                      name
                      avatarUrl
                    }
                  }
                  field {
                    ... on ProjectV2Field {
                      id
                      name
                    }
                  }
                }

                # プルリクエストフィールド
                ... on ProjectV2ItemFieldPullRequestValue {
                  pullRequests(first: 10) {
                    nodes {
                      id
                      number
                      title
                      url
                      state
                    }
                  }
                  field {
                    ... on ProjectV2Field {
                      id
                      name
                    }
                  }
                }

                # レビュアーフィールド
                ... on ProjectV2ItemFieldReviewerValue {
                  reviewers(first: 10) {
                    nodes {
                      ... on User {
                        id
                        login
                        name
                      }
                      ... on Team {
                        id
                        name
                        slug
                      }
                    }
                  }
                  field {
                    ... on ProjectV2Field {
                      id
                      name
                    }
                  }
                }
              }
            }
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

// AddCreatedSince adds a lower bound on the PR creation time.
// Used as a cursor to continue fetching beyond the GitHub Search API result cap.
func (b *SearchQueryBuilder) AddCreatedSince(createdSince time.Time) *SearchQueryBuilder {
	b.parts = append(b.parts, fmt.Sprintf("created:>=%s", createdSince.UTC().Format(time.RFC3339)))
	return b
}

// AddRepository adds repository filter to the query
func (b *SearchQueryBuilder) AddRepository(owner, repo string) *SearchQueryBuilder {
	if owner != "" && repo != "" {
		b.parts = append(b.parts, fmt.Sprintf("repo:%s/%s", owner, repo))
	}
	return b
}

// AddSort adds sorting to the query (sort:created-asc only for now)
func (b *SearchQueryBuilder) AddSort() *SearchQueryBuilder {
	b.parts = append(b.parts, "sort:created-asc")
	return b
}

// Build creates the final search query string
func (b *SearchQueryBuilder) Build() string {
	return strings.Join(b.parts, " ")
}
