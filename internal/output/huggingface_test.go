package output

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shantoislamdev/kothaset/internal/schema"
)

func TestHuggingFaceWriter(t *testing.T) {
	tmpDir := t.TempDir()
	// HF writer expects a directory path
	outPath := filepath.Join(tmpDir, "dataset_hf")

	s := schema.NewInstructionSchema()
	w := NewHuggingFaceWriter(s)

	if err := w.Open(outPath); err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	w.Write(&schema.Sample{Fields: map[string]any{"a": 1}})

	if err := w.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify structure
	if _, err := os.Stat(filepath.Join(outPath, "dataset_info.json")); err != nil {
		t.Error("dataset_info.json missing")
	}
	if _, err := os.Stat(filepath.Join(outPath, "train", "data-00000-of-00001.jsonl")); err != nil {
		t.Error("train data missing")
	}
}
