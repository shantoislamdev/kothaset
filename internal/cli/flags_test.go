package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/shantoislamdev/kothaset/internal/config"
)

func TestGenerateFlags(t *testing.T) {
	// Reset flags to defaults for test
	genCount = 100
	genSchema = ""

	// Execute command with flags
	cmd := generateCmd
	cmd.SetArgs([]string{
		"--count", "50",
		"--schema", "chat",
		"--output", "out.jsonl",
		"--seed", "123",
		"--input", "topics.txt",
	})

	// We don't want to actually RUN the command because it needs config/secrets/files
	// So we just parse flags.
	// Cobra parses flags when Execute() is called or ParseFlags()

	if err := cmd.ParseFlags([]string{
		"--count", "50",
		"--schema", "chat",
		"--output", "out.jsonl",
		"--seed", "123",
		"--input", "topics.txt",
	}); err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}

	if genCount != 50 {
		t.Errorf("Expected count 50, got %d", genCount)
	}
	if genSchema != "chat" {
		t.Errorf("Expected schema chat, got %s", genSchema)
	}
	if genOutput != "out.jsonl" {
		t.Errorf("Expected output out.jsonl, got %s", genOutput)
	}
	if genSeed != "123" {
		t.Errorf("Expected seed '123', got %s", genSeed)
	}
	if genInputFile != "topics.txt" {
		t.Errorf("Expected input topics.txt, got %s", genInputFile)
	}
}

func TestGenerate_BoundsChecking(t *testing.T) {
	// Preserve globals used by runGenerate
	origCount, origTemp, origMaxTokens, origWorkers, origInput := genCount, genTemp, genMaxTokens, genWorkers, genInputFile
	defer func() {
		genCount, genTemp, genMaxTokens, genWorkers, genInputFile = origCount, origTemp, origMaxTokens, origWorkers, origInput
	}()

	tests := []struct {
		name    string
		setup   func()
		wantErr string
	}{
		{
			name: "negative count",
			setup: func() {
				genInputFile = "topic"
				genCount = -1
				genTemp = 0.7
				genMaxTokens = 0
				genWorkers = 1
			},
			wantErr: "--count must be >= 1",
		},
		{
			name: "temperature too high",
			setup: func() {
				genInputFile = "topic"
				genCount = 1
				genTemp = 2.1
				genMaxTokens = 0
				genWorkers = 1
			},
			wantErr: "--temperature must be between 0 and 2.0",
		},
		{
			name: "negative max tokens",
			setup: func() {
				genInputFile = "topic"
				genCount = 1
				genTemp = 0.7
				genMaxTokens = -5
				genWorkers = 1
			},
			wantErr: "--max-tokens must be >= 0",
		},
		{
			name: "invalid workers",
			setup: func() {
				genInputFile = "topic"
				genCount = 1
				genTemp = 0.7
				genMaxTokens = 0
				genWorkers = 0
			},
			wantErr: "--workers must be >= 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := runGenerate(generateCmd, nil)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want contains %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestGenerate_SeedValidationStrict(t *testing.T) {
	origCfg, origSecrets := cfg, secrets
	origCount, origTemp, origMaxTokens := genCount, genTemp, genMaxTokens
	origWorkers, origInput, origOutput := genWorkers, genInputFile, genOutput
	origSeed, origDryRun := genSeed, genDryRun
	defer func() {
		cfg, secrets = origCfg, origSecrets
		genCount, genTemp, genMaxTokens = origCount, origTemp, origMaxTokens
		genWorkers, genInputFile, genOutput = origWorkers, origInput, origOutput
		genSeed, genDryRun = origSeed, origDryRun
	}()

	cfg = &config.Config{
		Version: "1.0",
		Global: config.GlobalConfig{
			Provider:     "openai",
			Schema:       "instruction",
			Model:        "gpt-5.2",
			OutputFormat: "jsonl",
		},
	}
	secrets = &config.SecretsConfig{
		Providers: []config.ProviderConfig{
			{
				Name:       "openai",
				Type:       "openai",
				APIKey:     "test-key",
				MaxRetries: 3,
				Timeout:    config.Duration{Duration: time.Second},
			},
		},
	}

	genCount = 1
	genTemp = 0.7
	genMaxTokens = 0
	genWorkers = 1
	genInputFile = "topic"
	genOutput = "out.jsonl"
	genSeed = "123abc"
	genDryRun = true

	err := runGenerate(generateCmd, nil)
	if err == nil {
		t.Fatalf("expected seed validation error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid seed value") {
		t.Fatalf("unexpected error: %v", err)
	}
}
