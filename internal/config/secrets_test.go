package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveAPIKey(t *testing.T) {
	// Set up env var
	os.Setenv("TEST_API_KEY", "secret-value")
	defer os.Unsetenv("TEST_API_KEY")

	tests := []struct {
		name       string
		apiKey     string
		provType   string
		want       string
		wantErr    bool
		errContain string
	}{
		{
			name:    "raw key",
			apiKey:  "sk-123456",
			want:    "sk-123456",
			wantErr: false,
		},
		{
			name:    "env var ref",
			apiKey:  "env.TEST_API_KEY",
			want:    "secret-value",
			wantErr: false,
		},
		{
			name:       "missing env var",
			apiKey:     "env.MISSING_KEY",
			want:       "",
			wantErr:    true,
			errContain: "environment variable not set",
		},
		{
			name:    "legacy secret ref env",
			apiKey:  "${env:TEST_API_KEY}",
			want:    "secret-value",
			wantErr: false,
		},
		{
			name:       "legacy secret ref missing env",
			apiKey:     "${env:MISSING_KEY}",
			want:       "",
			wantErr:    true,
			errContain: "environment variable not set",
		},
		{
			name:       "invalid secret ref format",
			apiKey:     "${invalid}",
			want:       "",
			wantErr:    true,
			errContain: "invalid secret reference format",
		},
		{
			name:     "empty api key uses default env var",
			apiKey:   "",
			provType: "openai",
			want:     "",
			wantErr:  true, // OPENAI_API_KEY not set in test
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ProviderConfig{
				Name:   "test-provider",
				Type:   tt.provType,
				APIKey: tt.apiKey,
			}

			got, err := resolveAPIKey(cfg)

			if tt.wantErr {
				if err == nil {
					t.Errorf("resolveAPIKey() expected error but got nil")
					return
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("resolveAPIKey() error = %v, want error containing %q", err, tt.errContain)
				}
				return
			}

			if err != nil {
				t.Errorf("resolveAPIKey() unexpected error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("resolveAPIKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadSecretsConfig_WithEnv(t *testing.T) {
	os.Setenv("TEST_KEY", "my-secret-key")
	defer os.Unsetenv("TEST_KEY")

	tmpDir := t.TempDir()
	secretsPath := filepath.Join(tmpDir, ".secrets.yaml")
	content := []byte(`
providers:
  - name: "test-provider"
    type: "openai"
    api_key: "env.TEST_KEY"
`)
	if err := os.WriteFile(secretsPath, content, 0644); err != nil {
		t.Fatalf("Failed to write temp secrets: %v", err)
	}

	secrets, err := LoadSecretsConfig(secretsPath)
	if err != nil {
		t.Fatalf("LoadSecretsConfig failed: %v", err)
	}

	if len(secrets.Providers) != 1 {
		t.Fatalf("Expected 1 provider, got %d", len(secrets.Providers))
	}

	if secrets.Providers[0].APIKey != "my-secret-key" {
		t.Errorf("Expected api_key 'my-secret-key', got '%s'", secrets.Providers[0].APIKey)
	}
}

func TestMaskSecret(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1234567890", "1234...7890"},
		{"short", "********"},
		{"12345678", "********"},
	}

	for _, tt := range tests {
		if got := MaskSecret(tt.input); got != tt.want {
			t.Errorf("MaskSecret(%s) = %s, want %s", tt.input, got, tt.want)
		}
	}
}
