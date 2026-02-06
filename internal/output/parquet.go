package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/shantoislamdev/kothaset/internal/schema"
)

// ParquetWriter writes samples in Parquet columnar format
// Note: This is a simplified JSON-based implementation.
// For production Parquet support, integrate github.com/xitongsys/parquet-go
type ParquetWriter struct {
	schema    schema.Schema
	path      string
	samples   []*schema.Sample
	batchSize int
	mu        sync.Mutex
}

// NewParquetWriter creates a new Parquet writer
func NewParquetWriter(sch schema.Schema) *ParquetWriter {
	return &ParquetWriter{
		schema:    sch,
		samples:   make([]*schema.Sample, 0, 1000),
		batchSize: 1000,
	}
}

func (w *ParquetWriter) Format() string { return "parquet" }

func (w *ParquetWriter) Open(path string) error {
	w.path = path
	// Ensure directory exists
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func (w *ParquetWriter) Write(sample *schema.Sample) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.samples = append(w.samples, sample)
	return nil
}

func (w *ParquetWriter) Flush() error {
	// No-op for batch writer, actual write happens on Close
	return nil
}

func (w *ParquetWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.samples) == 0 {
		return nil
	}

	// For now, write as JSON array (Parquet requires external library)
	// This serves as a placeholder that can be replaced with real Parquet
	file, err := os.Create(w.path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create columnar structure
	columns := make(map[string][]any)
	for _, sample := range w.samples {
		for key, value := range sample.Fields {
			columns[key] = append(columns[key], value)
		}
	}

	// Write metadata
	metadata := map[string]any{
		"format":   "parquet-placeholder",
		"schema":   w.schema.Name(),
		"num_rows": len(w.samples),
		"columns":  columns,
		"_note":    "Install parquet-go for native Parquet support",
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(metadata)
}

// SetBatchSize sets the batch size for writes
func (w *ParquetWriter) SetBatchSize(size int) {
	w.batchSize = size
}

// writeParquetNative would be the native implementation
// Requires: github.com/xitongsys/parquet-go
func (w *ParquetWriter) writeParquetNative() error {
	// TODO: Implement with parquet-go when dependency is added
	// Example structure:
	// fw, err := local.NewLocalFileWriter(w.path)
	// pw, err := writer.NewParquetWriter(fw, schema, 4)
	// for _, sample := range w.samples {
	//     pw.Write(sample)
	// }
	// pw.WriteStop()
	// fw.Close()
	return fmt.Errorf("native Parquet not yet implemented - install parquet-go dependency")
}
