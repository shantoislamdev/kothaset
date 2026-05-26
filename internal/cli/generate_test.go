package cli

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"30 seconds", 30 * time.Second, "30s"},
		{"59 seconds", 59 * time.Second, "59s"},
		{"1 minute", 60 * time.Second, "1m0s"},
		{"90 seconds", 90 * time.Second, "1m30s"},
		{"1 hour", 3600 * time.Second, "1h0m"},
		{"1h1m1s", 3661 * time.Second, "1h1m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.d)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

func TestHasParentPathTraversal(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"data/file.jsonl", false},
		{"../etc/passwd", true},
		{"data/../out.jsonl", true},
		{"data/..\\out.jsonl", true},
		{"data/sub/dir", false},
		{"..", true},
		{"data/..", true},
		{"", false},
		{"file.jsonl", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := hasParentPathTraversal(tt.path)
			if got != tt.want {
				t.Errorf("hasParentPathTraversal(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
