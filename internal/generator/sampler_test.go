package generator

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFileSampler(t *testing.T) {
	// Create temp input file
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "topics.txt")
	content := []byte("Topic A\nTopic B\n# Comment\nTopic C\n")
	if err := os.WriteFile(inputPath, content, 0644); err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	// Test NewFileSampler
	sampler, err := NewFileSampler(inputPath)
	if err != nil {
		t.Fatalf("NewFileSampler failed: %v", err)
	}

	if sampler.Count() != 3 {
		t.Errorf("Expected 3 topics, got %d", sampler.Count())
	}

	ctx := context.Background()

	// Test Sample (sequential)
	topic, _ := sampler.Sample(ctx, 0)
	if topic != "Topic A" {
		t.Errorf("Expected Topic A, got %s", topic)
	}

	topic, _ = sampler.Sample(ctx, 1)
	if topic != "Topic B" {
		t.Errorf("Expected Topic B, got %s", topic)
	}

	topic, _ = sampler.Sample(ctx, 2)
	if topic != "Topic C" {
		t.Errorf("Expected Topic C, got %s", topic)
	}

	// Test wrapping
	topic, _ = sampler.Sample(ctx, 3)
	if topic != "Topic A" {
		t.Errorf("Expected Topic A (wrapped), got %s", topic)
	}
}

func TestNewSampler_PermissionError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission mode behavior is unreliable on Windows")
	}

	tmpDir := t.TempDir()
	restrictedDir := filepath.Join(tmpDir, "restricted")
	if err := os.Mkdir(restrictedDir, 0o755); err != nil {
		t.Fatalf("failed to create restricted dir: %v", err)
	}
	path := filepath.Join(restrictedDir, "restricted.txt")
	if err := os.WriteFile(path, []byte("secret"), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.Chmod(restrictedDir, 0o000); err != nil {
		t.Fatalf("failed to chmod dir: %v", err)
	}
	defer os.Chmod(restrictedDir, 0o755)

	_, err := NewSampler(path)
	if err == nil {
		t.Fatalf("expected permission/access error, got nil")
	}
	if !strings.Contains(err.Error(), "cannot access input file") && !errors.Is(err, os.ErrPermission) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTopics_DefensiveCopy(t *testing.T) {
	s := &FileSampler{topics: []string{"A", "B"}}

	out := s.Topics()
	out[0] = "MUTATED"

	if got := s.topics[0]; got != "A" {
		t.Fatalf("internal topics mutated by caller: got %q", got)
	}
}
