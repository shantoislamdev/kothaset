package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", cfg.Version)
	}
	if cfg.Global.Provider != "openai" {
		t.Errorf("Expected default provider openai, got %s", cfg.Global.Provider)
	}
	if cfg.Global.CheckpointEvery != 10 {
		t.Errorf("Expected default checkpoint_every 10, got %d", cfg.Global.CheckpointEvery)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Version: "1.0",
				Global: GlobalConfig{
					Provider: "openai",
					Schema:   "instruction",
					Model:    "gpt-5.2",
				},
			},
			wantErr: false,
		},
		{
			name: "missing version",
			config: &Config{
				Global: GlobalConfig{
					Provider: "openai",
				},
			},
			wantErr: true,
		},
		{
			name: "missing provider",
			config: &Config{
				Version: "1.0",
				Global: GlobalConfig{
					Schema: "instruction",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadPublicConfig(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "kothaset.yaml")
	content := []byte(`
version: "1.0"
global:
  provider: "openai"
  schema: "chat"
  timeout: 30s
`)
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	// Test loading
	cfg, err := LoadPublicConfig(configPath)
	if err != nil {
		t.Fatalf("LoadPublicConfig failed: %v", err)
	}

	if cfg.Global.Schema != "chat" {
		t.Errorf("Expected schema chat, got %s", cfg.Global.Schema)
	}
	if cfg.Global.Timeout.Duration != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", cfg.Global.Timeout.Duration)
	}
}

func TestDuration_UnmarshalYAML(t *testing.T) {
	// Implicitly tested in LoadPublicConfig via yaml decoding
}
