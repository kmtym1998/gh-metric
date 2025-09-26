package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/kmtym1998/gh-metric/internal/github"
)

// Formatter interface for different output formats
type Formatter interface {
	Format(metrics []github.PRMetrics, writer io.Writer) error
}

// JSONFormatter formats output as JSON
type JSONFormatter struct{}

// Format implements Formatter for JSON output
func (f *JSONFormatter) Format(metrics []github.PRMetrics, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(metrics)
}

// CSVFormatter formats output as CSV
type CSVFormatter struct{}

// Format implements Formatter for CSV output
func (f *CSVFormatter) Format(metrics []github.PRMetrics, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{
		"number",
		"title",
		"author",
		"created_at",
		"url",
		"until_first_review",
		"until_first_approve",
		"until_merge",
		"comment_count",
		"changed_files",
		"additions",
		"deletions",
		"reviewers",
	}

	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, metric := range metrics {
		row := []string{
			strconv.Itoa(metric.Number),
			metric.Title,
			metric.Author,
			metric.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			metric.URL,
			formatFloat(metric.UntilFirstReview),
			formatFloat(metric.UntilFirstApprove),
			formatFloat(metric.UntilMerge),
			strconv.Itoa(metric.CommentCount),
			strconv.Itoa(metric.ChangedFiles),
			strconv.Itoa(metric.Additions),
			strconv.Itoa(metric.Deletions),
			metric.Reviewers,
		}

		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// formatFloat formats a float64 to string with reasonable precision
func formatFloat(f float64) string {
	if f == 0 {
		return "0"
	}
	return fmt.Sprintf("%.2f", f)
}

// GetFormatter returns the appropriate formatter for the given format
func GetFormatter(format string) (Formatter, error) {
	switch format {
	case "json":
		return &JSONFormatter{}, nil
	case "csv":
		return &CSVFormatter{}, nil
	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}
}
