package github

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pr(number int, author string, createdAt time.Time) PullRequest {
	p := PullRequest{Number: number, CreatedAt: createdAt}
	p.Author.Login = author
	return p
}

func numbers(prs []PullRequest) []int {
	result := make([]int, 0, len(prs))
	for _, p := range prs {
		result = append(result, p.Number)
	}
	return result
}

var (
	day1 = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	day2 = time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	day3 = time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)
	day4 = time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC)
)

func TestFetchMergedPRs(t *testing.T) {
	opts := FetchPROptions{Owner: "o", Repo: "r", Since: "2026-01-01"}

	t.Run("pages through a single search when all results are available", func(t *testing.T) {
		fake := &fakeExecutor{t: t, searches: []fakeSearch{{
			issueCount: 3,
			pages: [][]PullRequest{
				{pr(1, "alice", day1), pr(2, "alice", day2)},
				{pr(3, "alice", day3)},
			},
		}}}

		prs, err := (&Service{client: fake}).FetchMergedPRs(t.Context(), opts)

		require.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, numbers(prs))
		require.Len(t, fake.queries, 1)
		assert.NotContains(t, fake.queries[0], "created:>=")
	})

	t.Run("continues with a created cursor when the result cap is hit", func(t *testing.T) {
		// The first search reports 5 matches but only 3 are retrievable (result cap).
		// The second search starts at the created time of the last PR, so #3 appears in both.
		fake := &fakeExecutor{t: t, searches: []fakeSearch{
			{
				issueCount: 5,
				pages:      [][]PullRequest{{pr(1, "alice", day1), pr(2, "alice", day2), pr(3, "alice", day3)}},
			},
			{
				createdSince: "2026-01-03T00:00:00Z",
				issueCount:   3,
				pages:        [][]PullRequest{{pr(3, "alice", day3), pr(4, "alice", day3), pr(5, "alice", day4)}},
			},
		}}

		prs, err := (&Service{client: fake}).FetchMergedPRs(t.Context(), opts)

		require.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3, 4, 5}, numbers(prs), "boundary PR must not be duplicated")
		require.Len(t, fake.queries, 2)
		assert.Contains(t, fake.queries[1], "created:>=2026-01-03T00:00:00Z")
	})

	t.Run("advances the cursor even when every PR so far was filtered out", func(t *testing.T) {
		fake := &fakeExecutor{t: t, searches: []fakeSearch{
			{
				issueCount: 3,
				pages:      [][]PullRequest{{pr(1, "bob", day1), pr(2, "bob", day2)}},
			},
			{
				createdSince: "2026-01-02T00:00:00Z",
				issueCount:   2,
				pages:        [][]PullRequest{{pr(2, "bob", day2), pr(3, "alice", day3)}},
			},
		}}
		filtered := opts
		filtered.TargetUsers = []string{"alice"}

		prs, err := (&Service{client: fake}).FetchMergedPRs(t.Context(), filtered)

		require.NoError(t, err)
		assert.Equal(t, []int{3}, numbers(prs))
		require.Len(t, fake.queries, 2)
	})

	t.Run("stops at the limit without issuing another search", func(t *testing.T) {
		fake := &fakeExecutor{t: t, searches: []fakeSearch{{
			issueCount: 5,
			pages: [][]PullRequest{
				{pr(1, "alice", day1), pr(2, "alice", day2)},
				{pr(3, "alice", day3)},
			},
		}}}
		limited := opts
		limited.Limit = 2

		prs, err := (&Service{client: fake}).FetchMergedPRs(t.Context(), limited)

		require.NoError(t, err)
		assert.Equal(t, []int{1, 2}, numbers(prs))
		require.Len(t, fake.queries, 1)
	})

	t.Run("fails when the cursor cannot advance", func(t *testing.T) {
		// Every retrievable PR shares the same created time, so re-searching returns the same set forever
		samePage := [][]PullRequest{{pr(1, "alice", day1), pr(2, "alice", day1)}}
		fake := &fakeExecutor{t: t, searches: []fakeSearch{
			{issueCount: 3, pages: samePage},
			{createdSince: "2026-01-01T00:00:00Z", issueCount: 3, pages: samePage},
		}}

		_, err := (&Service{client: fake}).FetchMergedPRs(t.Context(), opts)

		require.ErrorContains(t, err, "cannot advance search cursor")
		require.Len(t, fake.queries, 2)
	})
}
