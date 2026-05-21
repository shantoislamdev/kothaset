package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasExtension_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name string
		path string
		ext  string
		want bool
	}{
		{name: "lowercase", path: "dataset.jsonl", ext: ".jsonl", want: true},
		{name: "uppercase", path: "dataset.JSONL", ext: ".jsonl", want: true},
		{name: "mixedcase", path: "dataset.JsOn", ext: ".json", want: true},
		{name: "non-match", path: "dataset.txt", ext: ".jsonl", want: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := hasExtension(tc.path, tc.ext); got != tc.want {
				t.Fatalf("hasExtension(%q, %q) = %v, want %v", tc.path, tc.ext, got, tc.want)
			}
		})
	}
}

func TestDetectFormat_CaseInsensitive(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{path: "sample.JSONL", want: "jsonl"},
		{path: "sample.JsOn", want: ""},
		{path: "sample.CSV", want: ""},
		{path: "sample.txt", want: ""},
	}

	for _, tc := range tests {
		if got := detectFormat(tc.path); got != tc.want {
			t.Fatalf("detectFormat(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}

func TestValidateConfigCmd_LoadsPathArgument(t *testing.T) {
	origCfg := cfg
	defer func() { cfg = origCfg }()
	cfg = nil

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "kothaset.yaml")
	content := []byte(`version: "1.0"
global:
  provider: openai
  schema: instruction
  model: gpt-5.2
  output_format: jsonl
`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	if err := validateConfigCmd.RunE(validateConfigCmd, []string{configPath}); err != nil {
		t.Fatalf("expected config path validation to pass, got: %v", err)
	}
}

func BenchmarkValidateJSONL(b *testing.B) {
	// Create a temporary JSONL file with 1000 lines
	dir := b.TempDir()
	path := filepath.Join(dir, "test.jsonl")
	file, err := os.Create(path)
	if err != nil {
		b.Fatal(err)
	}

	jsonLine := `{"id": "12345", "name": "John Doe", "email": "john.doe@example.com", "isActive": true, "roles": ["admin", "user"], "metadata": {"created_at": "2023-01-01T00:00:00Z", "login_count": 42}}` + "\n"

	for i := 0; i < 1000; i++ {
		file.WriteString(jsonLine)
	}
	file.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count, err := validateJSONL(path)
		if err != nil {
			b.Fatal(err)
		}
		if count != 1000 {
			b.Fatalf("expected 1000 rows, got %d", count)
		}
	}
}
