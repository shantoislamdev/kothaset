package output

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/shantoislamdev/kothaset/internal/schema"
)

func TestJSONLWriter(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "test.jsonl")

	s := schema.NewInstructionSchema()
	w := NewJSONLWriter(s)

	if err := w.Open(outPath); err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	// Write sample
	sample := &schema.Sample{
		Fields: map[string]any{
			"instruction": "Q",
			"output":      "A",
		},
	}
	if err := w.Write(sample); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify content
	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal(content, &data); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if data["instruction"] != "Q" {
		t.Errorf("Expected instruction Q, got %v", data["instruction"])
	}
}

func TestJSONLWriter_OpenAppend(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "append.jsonl")

	// Create initial file
	if err := os.WriteFile(outPath, []byte(`{"a":1}`+"\n"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	s := schema.NewInstructionSchema()
	w := NewJSONLWriter(s)
	if err := w.OpenAppend(outPath); err != nil {
		t.Fatalf("OpenAppend failed: %v", err)
	}

	if err := w.Write(&schema.Sample{Fields: map[string]any{"b": 2}}); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	w.Close()

	// Verify line count
	file, _ := os.Open(outPath)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	lines := 0
	for scanner.Scan() {
		lines++
	}
	if lines != 2 {
		t.Errorf("Expected 2 lines, got %d", lines)
	}
}

func TestJSONLWriter_Open_CreatesParentDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "deep", "nested", "dataset.jsonl")

	w := NewJSONLWriter(schema.NewInstructionSchema())
	if err := w.Open(outPath); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected output file to exist at %q: %v", outPath, err)
	}
}

func TestJSONLWriter_OpenAppend_CreatesParentDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "deep", "append", "dataset.jsonl")

	w := NewJSONLWriter(schema.NewInstructionSchema())
	if err := w.OpenAppend(outPath); err != nil {
		t.Fatalf("OpenAppend failed: %v", err)
	}
	if err := w.Write(&schema.Sample{Fields: map[string]any{"a": 1}}); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected output file to exist at %q: %v", outPath, err)
	}
}
