package output

import (
	"encoding/json"
	"io"

	"github.com/kmtym1998/gh-metric/internal/usecase"
)

// JSONFormatter formats output as JSON
type JSONFormatter struct{}

// Format implements Formatter for JSON output
func (f *JSONFormatter) Format(metrics []usecase.PRMetric, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(metrics)
}
