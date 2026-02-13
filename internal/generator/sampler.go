package generator

import (
	"bufio"
	"context"
	"fmt"

	"os"
	"strings"
	"sync"
)

// Sampler provides topics/seeds for sample generation
type Sampler interface {
	// Sample returns a topic for the given index
	Sample(ctx context.Context, index int) (string, error)
}

// FileSampler reads topics from a file (one per line)
type FileSampler struct {
	topics []string
	mu     sync.Mutex
}

// NewSampler creates a sampler from a file path or inline string
func NewSampler(input string) (Sampler, error) {
	// Check if input is a file
	info, err := os.Stat(input)
	if err == nil && !info.IsDir() {
		return NewFileSampler(input)
	}

	// If error is other than NotExist, return it (e.g. permission error)
	if err != nil && !os.IsNotExist(err) {
		// If it looks like a file path, treat the error as real.
		if strings.ContainsAny(input, `/\`) || strings.Contains(input, ".") {
			return nil, fmt.Errorf("cannot access input file '%s': %w", input, err)
		}
		// Otherwise, fall through to treat as inline topic.
	}

	// Treat as inline string (single topic)
	// User requested that inline input should be a single topic only.
	// For multiple topics, a file is required.
	var topics []string
	trimmed := strings.TrimSpace(input)
	if trimmed != "" {
		topics = append(topics, trimmed)
	}

	if len(topics) == 0 {
		return nil, fmt.Errorf("input provided but contains no valid topics")
	}

	return &FileSampler{
		topics: topics,
	}, nil
}

// NewFileSampler creates a sampler from a file
func NewFileSampler(path string) (*FileSampler, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var topics []string
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // up to 1MB per line
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			topics = append(topics, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(topics) == 0 {
		return nil, fmt.Errorf("input file is empty")
	}

	return &FileSampler{
		topics: topics,
	}, nil
}

// Sample returns a topic for the given index
func (s *FileSampler) Sample(ctx context.Context, index int) (string, error) {
	if len(s.topics) == 0 {
		return "", nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Use modulo for sequential access, or random for variety
	// Here we use index modulo for predictable coverage
	idx := index % len(s.topics)
	return s.topics[idx], nil
}

// Topics returns all loaded topics
func (s *FileSampler) Topics() []string {
	out := make([]string, len(s.topics))
	copy(out, s.topics)
	return out
}

// Count returns the number of topics
func (s *FileSampler) Count() int {
	return len(s.topics)
}
