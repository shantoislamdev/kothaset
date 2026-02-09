package generator

import (
	"context"
	"os"
	"path/filepath"
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
