package github

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// fakeSearch is the canned result of one search query.
// createdSince is the value expected in the "created:>=" condition; empty means no condition.
type fakeSearch struct {
	createdSince string
	issueCount   int
	pages        [][]PullRequest
}

// fakeExecutor serves canned search results and records the queries it received
type fakeExecutor struct {
	t        *testing.T
	searches []fakeSearch
	queries  []string
}

func (f *fakeExecutor) ExecuteQuery(_ string, variables any, result any) error {
	vars, ok := variables.(QueryVariables)
	require.True(f.t, ok, "variables must be QueryVariables")

	search := f.findSearch(vars.Query)
	require.NotNil(f.t, search, "unexpected search query: %s", vars.Query)

	pageIndex := 0
	if vars.Cursor != nil {
		_, err := fmt.Sscanf(*vars.Cursor, "page-%d", &pageIndex)
		require.NoError(f.t, err)
	} else {
		f.queries = append(f.queries, vars.Query)
	}
	require.Less(f.t, pageIndex, len(search.pages), "cursor beyond last page")

	resp, ok := result.(*SearchResponse)
	require.True(f.t, ok, "result must be *SearchResponse")
	resp.Search.IssueCount = search.issueCount
	resp.Search.Nodes = search.pages[pageIndex]
	if pageIndex+1 < len(search.pages) {
		resp.Search.PageInfo.HasNextPage = true
		resp.Search.PageInfo.EndCursor = fmt.Sprintf("page-%d", pageIndex+1)
	}
	return nil
}

func (f *fakeExecutor) findSearch(query string) *fakeSearch {
	_, createdSince, hasCreated := strings.Cut(query, "created:>=")
	if hasCreated {
		createdSince, _, _ = strings.Cut(createdSince, " ")
	}
	for i := range f.searches {
		if f.searches[i].createdSince == createdSince {
			return &f.searches[i]
		}
	}
	return nil
}
