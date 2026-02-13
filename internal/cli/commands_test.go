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
