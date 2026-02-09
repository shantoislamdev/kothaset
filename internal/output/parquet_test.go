package output

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shantoislamdev/kothaset/internal/schema"
)

func TestParquetWriter_Native(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "test.parquet")

	s := schema.NewInstructionSchema()
	w := NewParquetWriter(s)

	if err := w.Open(outPath); err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	sample := &schema.Sample{
		ID: "test-1",
		Fields: map[string]any{
			"instruction": "Write a hello world program",
			"input":       "",
			"output":      "print('Hello, World!')",
		},
	}
	w.Write(sample)

	if err := w.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify file exists and has content
	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Expected non-empty parquet file")
	}

	// Parquet files start with "PAR1" magic bytes
	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if len(content) < 4 || string(content[:4]) != "PAR1" {
		t.Errorf("Expected Parquet magic bytes, got: %v", content[:4])
	}
}

func TestParquetWriter_Fallback(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "test_fallback.parquet")

	s := schema.NewInstructionSchema()
	w := NewParquetWriter(s)
	w.SetUseNative(false) // Use JSON fallback

	if err := w.Open(outPath); err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	sample := &schema.Sample{
		ID: "test-1",
		Fields: map[string]any{
			"instruction": "Test instruction",
		},
	}
	w.Write(sample)

	if err := w.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify file exists
	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	// Should contain JSON placeholder marker
	if len(content) == 0 {
		t.Error("Expected non-empty file")
	}
}

func TestParquetWriter_EmptySamples(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "empty.parquet")

	s := schema.NewInstructionSchema()
	w := NewParquetWriter(s)

	if err := w.Open(outPath); err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	// Don't write any samples

	if err := w.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// File should not exist or be empty
	_, err := os.Stat(outPath)
	if err == nil {
		t.Error("Expected no file for empty samples")
	}
}
