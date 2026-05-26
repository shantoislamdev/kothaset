package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveSecrets_EnvNotSet(t *testing.T) {
	tmpDir := t.TempDir()
	secretsPath := filepath.Join(tmpDir, ".secrets.yaml")
	content := []byte(`
providers:
  - name: "missing-env"
    type: "openai"
    api_key: "env.DOES_NOT_EXIST_FOR_TEST"
`)

	if err := os.WriteFile(secretsPath, content, 0o644); err != nil {
		t.Fatalf("failed to write secrets file: %v", err)
	}

	cfg, loadErr := LoadSecretsConfig(secretsPath)
	if loadErr != nil {
		t.Fatalf("LoadSecretsConfig should not fail before resolution: %v", loadErr)
	}

	resolveErr := ResolveSecrets(cfg)
	if resolveErr == nil {
		t.Fatal("ResolveSecrets should return error for missing env var")
	}

	errText := resolveErr.Error()
	if !strings.Contains(errText, "missing-env") {
		t.Errorf("expected provider name in error, got: %q", errText)
	}
	if !strings.Contains(errText, "environment variable not set") {
		t.Errorf("expected missing env in error, got: %q", errText)
	}
}

func TestValidate_Expanded(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name:    "missing model",
			cfg:     Config{Version: "1.0", Global: GlobalConfig{Provider: "openai", Schema: "instruction"}},
			wantErr: "global.model is required",
		},
		{
			name:    "negative concurrency",
			cfg:     Config{Version: "1.0", Global: GlobalConfig{Provider: "openai", Schema: "instruction", Model: "gpt-5.2", Concurrency: -1}},
			wantErr: "global.concurrency must be non-negative (0 = use default)",
		},
		{
			name:    "unsupported output format",
			cfg:     Config{Version: "1.0", Global: GlobalConfig{Provider: "openai", Schema: "instruction", Model: "gpt-5.2", OutputFormat: "json"}},
			wantErr: "unsupported output_format",
		},
		{
			name: "valid expanded config",
			cfg:  Config{Version: "1.0", Global: GlobalConfig{Provider: "openai", Schema: "instruction", Model: "gpt-5.2", Concurrency: 0, OutputFormat: "jsonl"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() error = %q, want contains %q", err.Error(), tt.wantErr)
			}
		})
	}
}
