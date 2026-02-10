package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/shantoislamdev/kothaset/internal/schema"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/writer"
)

// ParquetWriter writes samples in Parquet columnar format
type ParquetWriter struct {
	schema    schema.Schema
	path      string
	samples   []*schema.Sample
	batchSize int
	mu        sync.Mutex
	useNative bool
}

// NewParquetWriter creates a new Parquet writer
func NewParquetWriter(sch schema.Schema) *ParquetWriter {
	return &ParquetWriter{
		schema:    sch,
		samples:   make([]*schema.Sample, 0, 1000),
		batchSize: 1000,
		useNative: true, // Enable native Parquet by default
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

func (w *ParquetWriter) OpenAppend(path string) error {
	// Parquet doesn't support true append - we store samples in memory
	// and rewrite on Close. Just set the path and continue from where we left off.
	// The checkpoint system tracks how many samples were already completed.
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

	// Use native Parquet if enabled
	if w.useNative {
		return w.writeParquetNative()
	}

	// Fallback to JSON placeholder
	return w.writeJSONPlaceholder()
}

// SetBatchSize sets the batch size for writes
func (w *ParquetWriter) SetBatchSize(size int) {
	w.batchSize = size
}

// SetUseNative controls whether to use native Parquet or JSON fallback
func (w *ParquetWriter) SetUseNative(native bool) {
	w.useNative = native
}

// ParquetRecord is a generic struct for Parquet writing
// Since parquet-go requires struct tags, we use a map-based approach
type ParquetRecord struct {
	ID          string `parquet:"name=id, type=BYTE_ARRAY, convertedtype=UTF8"`
	Instruction string `parquet:"name=instruction, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Input       string `parquet:"name=input, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Output      string `parquet:"name=output, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
}

// writeParquetNative writes using parquet-go library
func (w *ParquetWriter) writeParquetNative() error {
	// Create local file writer
	fw, err := local.NewLocalFileWriter(w.path)
	if err != nil {
		return fmt.Errorf("failed to create parquet file: %w", err)
	}
	defer fw.Close()

	// Create Parquet writer with the record schema
	pw, err := writer.NewParquetWriter(fw, new(ParquetRecord), 4)
	if err != nil {
		return fmt.Errorf("failed to create parquet writer: %w", err)
	}

	// Set compression
	pw.CompressionType = parquet.CompressionCodec_SNAPPY

	// Write each sample
	for _, sample := range w.samples {
		record := ParquetRecord{
			ID:          sample.ID,
			Instruction: sample.GetString("instruction"),
			Input:       sample.GetString("input"),
			Output:      sample.GetString("output"),
		}
		if err := pw.Write(record); err != nil {
			pw.WriteStop()
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	// Finalize writing
	if err := pw.WriteStop(); err != nil {
		return fmt.Errorf("failed to finalize parquet: %w", err)
	}

	return nil
}

// writeJSONPlaceholder writes JSON-based placeholder format
func (w *ParquetWriter) writeJSONPlaceholder() error {
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
		"_note":    "Use SetUseNative(true) for native Parquet support",
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(metadata)
}
