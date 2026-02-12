package generator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shantoislamdev/kothaset/internal/provider"
	"github.com/shantoislamdev/kothaset/internal/schema"
)

// MockProvider implements provider.Provider for testing
type MockProvider struct {
	ShouldFail bool
	Delay      time.Duration
	Response   string
	Calls      int
	mu         sync.Mutex
}

func (m *MockProvider) Generate(ctx context.Context, req provider.GenerationRequest) (*provider.GenerationResponse, error) {
	m.mu.Lock()
	m.Calls++
	m.mu.Unlock()

	if m.Delay > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(m.Delay):
		}
	}

	if m.ShouldFail {
		return nil, fmt.Errorf("mock provider error")
	}

	return &provider.GenerationResponse{
		Content: m.Response,
		Model:   "mock-model",
		Usage: provider.TokenUsage{
			TotalTokens: 10,
		},
		Latency: time.Millisecond * 10,
	}, nil
}

func (m *MockProvider) Name() string                          { return "mock" }
func (m *MockProvider) Type() string                          { return "mock" }
func (m *MockProvider) Model() string                         { return "mock-model" }
func (m *MockProvider) SupportedModels() []string             { return []string{"mock-model"} }
func (m *MockProvider) SupportsStreaming() bool               { return false }
func (m *MockProvider) SupportsBatching() bool                { return false }
func (m *MockProvider) Validate() error                       { return nil }
func (m *MockProvider) HealthCheck(ctx context.Context) error { return nil }
func (m *MockProvider) Close() error                          { return nil }

// MockWriter implements output.Writer for testing.
type MockWriter struct {
	Samples []*schema.Sample
	mu      sync.Mutex
}

func (w *MockWriter) Open(path string) error { return nil }

func (w *MockWriter) OpenAppend(path string) error { return nil }

func (w *MockWriter) Write(sample *schema.Sample) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.Samples = append(w.Samples, sample)
	return nil
}

func (w *MockWriter) Flush() error   { return nil }
func (w *MockWriter) Close() error   { return nil }
func (w *MockWriter) Format() string { return "mock" }

// MockSampler implements Sampler for testing
type MockSampler struct {
	Topic string
}

func (s *MockSampler) Sample(ctx context.Context, index int) (string, error) {
	return s.Topic, nil
}
