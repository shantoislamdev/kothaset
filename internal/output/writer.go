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

	// OpenAppend opens the writer in append mode for resuming
	// This preserves existing data instead of truncating
	OpenAppend(path string) error

	// Write writes a single sample to the output
	Write(sample *schema.Sample) error

	// Flush flushes any buffered data to the OS
	Flush() error

	// Sync flushes buffered data and fsyncs to physical storage.
	// Use at checkpoint boundaries for crash-safe durability.
	Sync() error

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
	default:
		return nil, fmt.Errorf("unsupported output format: %s (supported: jsonl)", format)
	}
}

// SupportedFormats returns a list of supported output formats
func SupportedFormats() []string {
	return []string{"jsonl"}
}
