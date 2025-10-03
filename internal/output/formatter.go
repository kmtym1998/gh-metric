package output

import (
	"fmt"
	"io"

	"github.com/kmtym1998/gh-metric/internal/usecase"
)

// Formatter interface for different output formats
type Formatter interface {
	Format(metrics []usecase.PRMetric, writer io.Writer) error
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
		return &PRMetricCSVFormatter{}, nil
	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}
}
