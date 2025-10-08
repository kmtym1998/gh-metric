package output

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/kmtym1998/gh-metric/internal/usecase"
)

// PRMetricCSVFormatter formats output as CSV
type PRMetricCSVFormatter struct{}

// Format implements Formatter for CSV output
func (f *PRMetricCSVFormatter) Format(metrics []usecase.PRMetric, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{
		`"repository"`,
		`"number"`,
		`"title"`,
		`"author"`,
		`"created_at"`,
		`"created_week"`,
		`"created_month"`,
		`"merged_at"`,
		`"merged_week"`,
		`"merged_month"`,
		`"url"`,
		`"until_first_review"`,
		`"until_first_approve"`,
		`"until_merge"`,
		`"comment_count"`,
		`"changed_files"`,
		`"additions"`,
		`"deletions"`,
		`"approved_by"`,
		`"first_reviewed_by"`,
		`"first_approved_by"`,
	}

	rows := []string{strings.Join(header, ",")}
	// Write data rows
	for _, metric := range metrics {
		row := []string{
			`"` + escapeDoubleQuotes(metric.Repository) + `"`,
			`"` + escapeDoubleQuotes(strconv.Itoa(metric.Number)) + `"`,
			`"` + escapeDoubleQuotes(metric.Title) + `"`,
			`"` + escapeDoubleQuotes(metric.Author) + `"`,
			`"` + escapeDoubleQuotes(metric.CreatedAt.Format("2006-01-02T15:04:05Z07:00")) + `"`,
			`"` + escapeDoubleQuotes(metric.CreatedWeek.Format("2006-01-02")) + `"`,
			`"` + escapeDoubleQuotes(metric.CreatedMonth.Format("2006-01-02")) + `"`,
			`"` + escapeDoubleQuotes(metric.MergedAt.Format("2006-01-02T15:04:05Z07:00")) + `"`,
			`"` + escapeDoubleQuotes(metric.MergedWeek.Format("2006-01-02")) + `"`,
			`"` + escapeDoubleQuotes(metric.MergedMonth.Format("2006-01-02")) + `"`,
			`"` + escapeDoubleQuotes(metric.URL) + `"`,
			`"` + escapeDoubleQuotes(formatFloat(metric.UntilFirstReview)) + `"`,
			`"` + escapeDoubleQuotes(formatFloat(metric.UntilFirstApprove)) + `"`,
			`"` + escapeDoubleQuotes(formatFloat(metric.UntilMerge)) + `"`,
			`"` + escapeDoubleQuotes(strconv.Itoa(metric.CommentCount)) + `"`,
			`"` + escapeDoubleQuotes(strconv.Itoa(metric.ChangedFiles)) + `"`,
			`"` + escapeDoubleQuotes(strconv.Itoa(metric.Additions)) + `"`,
			`"` + escapeDoubleQuotes(strconv.Itoa(metric.Deletions)) + `"`,
			`"` + escapeDoubleQuotes(strings.Join(metric.ApprovedBy, ",")) + `"`,
			`"` + escapeDoubleQuotes(metric.FirstReviewedBy) + `"`,
			`"` + escapeDoubleQuotes(metric.FirstApprovedBy) + `"`,
		}
		rows = append(rows, strings.Join(row, ","))
	}

	if _, err := writer.Write([]byte(strings.Join(rows, "\n") + "\n")); err != nil {
		return fmt.Errorf("failed to write CSV rows: %w", err)
	}

	return nil
}

func escapeDoubleQuotes(s string) string {
	return strings.ReplaceAll(s, `"`, `""`)
}
