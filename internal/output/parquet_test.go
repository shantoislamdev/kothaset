package output

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/shantoislamdev/kothaset/internal/schema"
)

func TestParquetWriter(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "test.parquet")

	s := schema.NewInstructionSchema()
	w := NewParquetWriter(s)

	if err := w.Open(outPath); err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	sample := &schema.Sample{
		Fields: map[string]any{
			"col1": "val1",
		},
	}
	w.Write(sample)

	if err := w.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify it wrote the placeholder JSON
	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal(content, &data); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if data["format"] != "parquet-placeholder" {
		t.Error("Expected placeholder format")
	}
}
