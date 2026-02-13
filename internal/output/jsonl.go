package output

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"

	"github.com/shantoislamdev/kothaset/internal/schema"
)

// JSONLWriter writes samples as JSON Lines format
type JSONLWriter struct {
	schema schema.Schema
	file   *os.File
	writer *bufio.Writer
	mu     sync.Mutex
}

// NewJSONLWriter creates a new JSONL writer
func NewJSONLWriter(sch schema.Schema) *JSONLWriter {
	return &JSONLWriter{
		schema: sch,
	}
}

func (w *JSONLWriter) Format() string { return "jsonl" }

func (w *JSONLWriter) Open(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	w.file = file
	w.writer = bufio.NewWriterSize(file, 64*1024) // 64KB buffer
	return nil
}

func (w *JSONLWriter) Write(sample *schema.Sample) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	data, err := json.Marshal(sample.Fields)
	if err != nil {
		return err
	}

	if _, err := w.writer.Write(data); err != nil {
		return err
	}
	if _, err = w.writer.WriteString("\n"); err != nil {
		return err
	}
	// Flush to OS immediately so data survives application crashes
	return w.writer.Flush()
}

func (w *JSONLWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.writer != nil {
		return w.writer.Flush()
	}
	return nil
}

// Sync flushes buffered data and fsyncs to physical storage.
// Use at checkpoint boundaries for crash-safe durability.
func (w *JSONLWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.writer != nil {
		if err := w.writer.Flush(); err != nil {
			return err
		}
	}
	if w.file != nil {
		return w.file.Sync()
	}
	return nil
}

func (w *JSONLWriter) Close() error {
	if err := w.Flush(); err != nil {
		return err
	}
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// OpenAppend opens the file in append mode for resuming
func (w *JSONLWriter) OpenAppend(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	w.file = file
	w.writer = bufio.NewWriterSize(file, 64*1024)
	return nil
}
