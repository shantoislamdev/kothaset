package generator

import (
	"bufio"
	"context"

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

// NewFileSampler creates a sampler from a file
func NewFileSampler(path string) (*FileSampler, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var topics []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			topics = append(topics, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
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
	return s.topics
}

// Count returns the number of topics
func (s *FileSampler) Count() int {
	return len(s.topics)
}
