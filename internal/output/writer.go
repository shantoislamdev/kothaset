// Package output provides writers for different output formats.
package output

import (
	"fmt"

	"github.com/shantoislamdev/kothaset/internal/schema"
)

// Writer defines the interface for dataset output writers
type Writer interface {
	// Open initializes the writer for the given path
	Open(path string) error

	// Write writes a single sample to the output
	Write(sample *schema.Sample) error

	// Flush flushes any buffered data
	Flush() error

	// Close closes the writer and releases resources
	Close() error

	// Format returns the output format name
	Format() string
}

// NewWriter creates a new writer for the given format
func NewWriter(format string, sch schema.Schema) (Writer, error) {
	switch format {
	case "jsonl", "":
		return NewJSONLWriter(sch), nil
	case "parquet":
		return NewParquetWriter(sch), nil
	case "hf", "huggingface":
		return NewHuggingFaceWriter(sch), nil
	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}
}

// SupportedFormats returns a list of supported output formats
func SupportedFormats() []string {
	return []string{"jsonl", "parquet", "hf"}
}
