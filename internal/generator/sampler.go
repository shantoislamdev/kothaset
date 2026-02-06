package generator

import (
	"bufio"
	"context"
	"math/rand"
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
	rand   *rand.Rand
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
		rand:   rand.New(rand.NewSource(rand.Int63())),
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

// RandomSampler generates random topics from a predefined set
type RandomSampler struct {
	categories []string
	rand       *rand.Rand
	mu         sync.Mutex
}

// NewRandomSampler creates a sampler with default categories
func NewRandomSampler(seed int64) *RandomSampler {
	return &RandomSampler{
		categories: defaultCategories,
		rand:       rand.New(rand.NewSource(seed)),
	}
}

// Sample returns a random category
func (s *RandomSampler) Sample(ctx context.Context, index int) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.categories[s.rand.Intn(len(s.categories))], nil
}

// SetCategories sets custom categories
func (s *RandomSampler) SetCategories(cats []string) {
	s.categories = cats
}

var defaultCategories = []string{
	"science and technology",
	"history and culture",
	"mathematics and logic",
	"programming and software",
	"creative writing",
	"business and finance",
	"health and medicine",
	"education and learning",
	"arts and entertainment",
	"philosophy and ethics",
	"environment and nature",
	"cooking and recipes",
	"travel and geography",
	"sports and fitness",
	"music and audio",
	"language and linguistics",
	"psychology and behavior",
	"law and regulations",
	"data analysis",
	"general knowledge",
}

// CompositeSampler combines multiple samplers
type CompositeSampler struct {
	samplers []Sampler
	weights  []float64
	rand     *rand.Rand
	mu       sync.Mutex
}

// NewCompositeSampler creates a sampler that randomly picks from multiple sources
func NewCompositeSampler(samplers []Sampler, weights []float64, seed int64) *CompositeSampler {
	if len(weights) != len(samplers) {
		// Equal weights
		weights = make([]float64, len(samplers))
		for i := range weights {
			weights[i] = 1.0 / float64(len(samplers))
		}
	}
	return &CompositeSampler{
		samplers: samplers,
		weights:  weights,
		rand:     rand.New(rand.NewSource(seed)),
	}
}

// Sample picks a sampler and returns its sample
func (s *CompositeSampler) Sample(ctx context.Context, index int) (string, error) {
	s.mu.Lock()
	r := s.rand.Float64()
	s.mu.Unlock()

	var cumulative float64
	for i, w := range s.weights {
		cumulative += w
		if r < cumulative {
			return s.samplers[i].Sample(ctx, index)
		}
	}
	return s.samplers[len(s.samplers)-1].Sample(ctx, index)
}
