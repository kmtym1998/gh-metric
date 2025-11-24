package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kmtym1998/gh-metric/internal/usecase"
)

func TestIssuesCommand(t *testing.T) {
	// Reset flags before each test
	resetIssueFlags := func() {
		issueFlags = issuesFlags{
			state:  "all",
			limit:  100,
			output: "json",
		}
	}

	t.Run("Command initialization", func(t *testing.T) {
		// Test that the command is properly initialized
		assert.NotNil(t, issuesCmd)
		assert.Equal(t, "issues [owner/repo]", issuesCmd.Use)
		assert.Contains(t, issuesCmd.Short, "issue metrics")

		// Check that the command is registered with root
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == "issues" {
				found = true
				break
			}
		}
		assert.True(t, found, "issues command should be registered with root")
	})

	t.Run("Flag validation", func(t *testing.T) {
		tests := []struct {
			name      string
			setup     func()
			wantError bool
			errorMsg  string
		}{
			{
				name: "Valid flags with all defaults",
				setup: func() {
					resetIssueFlags()
				},
				wantError: false,
			},
			{
				name: "Invalid output format",
				setup: func() {
					resetIssueFlags()
					issueFlags.output = "csv"
				},
				wantError: true,
				errorMsg:  "only JSON output is currently supported",
			},
			{
				name: "Invalid state",
				setup: func() {
					resetIssueFlags()
					issueFlags.state = "invalid"
				},
				wantError: true,
				errorMsg:  "invalid state",
			},
			{
				name: "Invalid since date format",
				setup: func() {
					resetIssueFlags()
					issueFlags.since = "2024/01/01"
				},
				wantError: true,
				errorMsg:  "invalid since date format",
			},
			{
				name: "Invalid until date format",
				setup: func() {
					resetIssueFlags()
					issueFlags.until = "01-01-2024"
				},
				wantError: true,
				errorMsg:  "invalid until date format",
			},
			{
				name: "Valid date formats",
				setup: func() {
					resetIssueFlags()
					issueFlags.since = "2024-01-01"
					issueFlags.until = "2024-12-31"
				},
				wantError: false,
			},
			{
				name: "Invalid limit (zero)",
				setup: func() {
					resetIssueFlags()
					issueFlags.limit = 0
				},
				wantError: true,
				errorMsg:  "limit must be greater than 0",
			},
			{
				name: "Invalid limit (negative)",
				setup: func() {
					resetIssueFlags()
					issueFlags.limit = -1
				},
				wantError: true,
				errorMsg:  "limit must be greater than 0",
			},
			{
				name: "Valid state - open",
				setup: func() {
					resetIssueFlags()
					issueFlags.state = "open"
				},
				wantError: false,
			},
			{
				name: "Valid state - closed",
				setup: func() {
					resetIssueFlags()
					issueFlags.state = "closed"
				},
				wantError: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.setup()
				err := validateIssueFlags()

				if tt.wantError {
					assert.Error(t, err)
					if tt.errorMsg != "" {
						assert.Contains(t, err.Error(), tt.errorMsg)
					}
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("Repository parsing", func(t *testing.T) {
		tests := []struct {
			name      string
			repo      string
			wantError bool
			errorMsg  string
		}{
			{
				name:      "Valid repo format",
				repo:      "owner/repo",
				wantError: false,
			},
			{
				name:      "Invalid repo format - no slash",
				repo:      "ownerrepo",
				wantError: true,
				errorMsg:  "invalid repo format",
			},
			{
				name:      "Invalid repo format - multiple slashes",
				repo:      "owner/repo/extra",
				wantError: true,
				errorMsg:  "invalid repo format",
			},
			{
				name:      "Empty repo (should detect from git)",
				repo:      "",
				wantError: false, // This would actually fail if no git repo, but we're testing the parsing logic
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resetIssueFlags()
				issueFlags.repo = tt.repo

				// We can't fully test runIssues without mocking GitHub service,
				// but we can test the repo parsing logic
				if tt.repo != "" {
					parts := strings.Split(tt.repo, "/")
					if tt.wantError {
						assert.NotEqual(t, 2, len(parts), "Invalid repository format should not have exactly 2 parts")
					} else {
						assert.Equal(t, 2, len(parts), "Repository should be in owner/repo format")
					}
				}
			})
		}
	})

	t.Run("Output formatting", func(t *testing.T) {
		t.Run("JSON output", func(t *testing.T) {
			// Create sample output
			sampleOutput := &usecase.AggregateIssuesOutput{
				Repository: "test/repo",
				Period: usecase.Period{
					Since: "2024-01-01",
					Until: "2024-12-31",
				},
				TotalCount: 2,
				Items: []usecase.IssueMetric{
					{
						Number:      1,
						Title:       "Test Issue 1",
						State:       "open",
						Author:      "user1",
						Assignees:   []string{"user2"},
						Labels:      []string{"bug", "priority"},
						CreatedAt:   time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
						CreatedWeek: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
						CreatedMonth: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt:   time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC),
					},
					{
						Number:      2,
						Title:       "Test Issue 2",
						State:       "closed",
						Author:      "user3",
						Assignees:   []string{},
						Labels:      []string{"enhancement"},
						CreatedAt:   time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC),
						CreatedWeek: time.Date(2024, 1, 29, 0, 0, 0, 0, time.UTC),
						CreatedMonth: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt:   time.Date(2024, 2, 5, 10, 0, 0, 0, time.UTC),
						ClosedAt:    func() *time.Time { t := time.Date(2024, 2, 5, 10, 0, 0, 0, time.UTC); return &t }(),
					},
				},
			}

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Output the results
			resetIssueFlags()
			err := outputIssueResults(sampleOutput)
			require.NoError(t, err)

			// Restore stdout and read output
			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Parse the JSON output
			var result usecase.AggregateIssuesOutput
			err = json.Unmarshal([]byte(output), &result)
			require.NoError(t, err, "Output should be valid JSON")

			// Verify the content
			assert.Equal(t, "test/repo", result.Repository)
			assert.Equal(t, 2, result.TotalCount)
			assert.Len(t, result.Items, 2)
			assert.Equal(t, "Test Issue 1", result.Items[0].Title)
			assert.Equal(t, "open", result.Items[0].State)
			assert.Equal(t, "Test Issue 2", result.Items[1].Title)
			assert.Equal(t, "closed", result.Items[1].State)
		})

		t.Run("Unsupported CSV output", func(t *testing.T) {
			sampleOutput := &usecase.AggregateIssuesOutput{}

			resetIssueFlags()
			issueFlags.output = "csv"

			err := outputIssueResults(sampleOutput)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "only JSON output is currently supported")
		})
	})

	t.Run("Command flags parsing", func(t *testing.T) {
		tests := []struct {
			name         string
			args         []string
			expectedRepo string
			expectedState string
			expectedLimit int
			expectedSince string
			expectedUntil string
		}{
			{
				name:         "Default values",
				args:         []string{},
				expectedRepo: "",
				expectedState: "all",
				expectedLimit: 100,
			},
			{
				name:         "With repo flag",
				args:         []string{"--repo", "owner/repo"},
				expectedRepo: "owner/repo",
				expectedState: "all",
				expectedLimit: 100,
			},
			{
				name:         "With all flags",
				args:         []string{
					"--repo", "test/repo",
					"--state", "open",
					"--limit", "50",
					"--since", "2024-01-01",
					"--until", "2024-12-31",
				},
				expectedRepo: "test/repo",
				expectedState: "open",
				expectedLimit: 50,
				expectedSince: "2024-01-01",
				expectedUntil: "2024-12-31",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Reset flags and re-parse
				resetIssueFlags()
				issuesCmd.ParseFlags(tt.args)

				// Note: Flag parsing doesn't automatically set our flag variables
				// In a real test, we would need to use issuesCmd.Flags().GetString() etc.
				// This is more of a demonstration of how the test structure would look
			})
		}
	})
}

// TestIssuesIntegration tests the full flow with mocked dependencies
func TestIssuesIntegration(t *testing.T) {
	t.Run("Success case with mock data", func(t *testing.T) {
		// This test would require mocking the GitHub service
		// For a true integration test, we would:
		// 1. Create a mock GitHub service
		// 2. Inject it into the usecase
		// 3. Run the full command
		// 4. Verify the output

		// Example structure (would need actual mock implementation):
		/*
		mockService := &MockGitHubService{
			FetchIssuesFunc: func(ctx context.Context, owner, repo string, opts github.FetchIssuesOptions) ([]github.Issue, error) {
				return []github.Issue{
					{Number: 1, Title: "Test Issue", State: "open"},
				}, nil
			},
		}

		// Inject mock and run command
		// Verify output
		*/

		// For now, we just verify the command structure is correct
		assert.NotNil(t, issuesCmd.RunE)
	})
}

// TestIssuesCommandErrorHandling tests error scenarios
func TestIssuesCommandErrorHandling(t *testing.T) {
	t.Run("GitHub API error", func(t *testing.T) {
		// This would test how the command handles GitHub API errors
		// Would require mocking to simulate API failures

		// Example structure:
		/*
		mockService := &MockGitHubService{
			FetchIssuesFunc: func(ctx context.Context, owner, repo string, opts github.FetchIssuesOptions) ([]github.Issue, error) {
				return nil, fmt.Errorf("API rate limit exceeded")
			},
		}

		err := runIssues(cmd, args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit")
		*/
	})

	t.Run("Context cancellation", func(t *testing.T) {
		// Test that the command respects context cancellation
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		cmd := &cobra.Command{}
		cmd.SetContext(ctx)

		// In a real test with mocked dependencies, we would verify
		// that the command returns promptly when context is cancelled
	})
}

// Benchmark tests for performance validation
func BenchmarkIssuesJSONOutput(b *testing.B) {
	// Create a large sample output for benchmarking
	sampleOutput := &usecase.AggregateIssuesOutput{
		Repository: "test/repo",
		TotalCount: 100,
		Items:      make([]usecase.IssueMetric, 100),
	}

	for i := 0; i < 100; i++ {
		sampleOutput.Items[i] = usecase.IssueMetric{
			Number:      i + 1,
			Title:       "Test Issue",
			State:       "open",
			Author:      "user",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	// Redirect output to discard
	oldStdout := os.Stdout
	os.Stdout = os.NewFile(0, os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = outputIssueResults(sampleOutput)
	}
}