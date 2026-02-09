package cli

import (
	"testing"
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
	if genSeed != 123 {
		t.Errorf("Expected seed 123, got %d", genSeed)
	}
	if genInputFile != "topics.txt" {
		t.Errorf("Expected input topics.txt, got %s", genInputFile)
	}
}
