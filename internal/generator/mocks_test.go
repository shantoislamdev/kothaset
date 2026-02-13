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
func (m *MockProvider) SupportsStreaming() bool               { return false }
func (m *MockProvider) Validate() error                       { return nil }
func (m *MockProvider) HealthCheck(ctx context.Context) error { return nil }
func (m *MockProvider) Close() error                          { return nil }

// MockWriter implements output.Writer for testing.
type MockWriter struct {
	Samples     []*schema.Sample
	FailOnWrite bool
	WriteCount  int
	FailAfter   int
	SyncCalls   int
	CloseCalls  int
	FlushCalls  int
	OpenCalls   int
	OpenAppends int
	mu          sync.Mutex
}

func (w *MockWriter) Open(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.OpenCalls++
	return nil
}

func (w *MockWriter) OpenAppend(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.OpenAppends++
	return nil
}

func (w *MockWriter) Write(sample *schema.Sample) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.WriteCount++
	if w.FailOnWrite {
		if w.FailAfter <= 0 || w.WriteCount > w.FailAfter {
			return fmt.Errorf("mock writer error")
		}
	}
	w.Samples = append(w.Samples, sample)
	return nil
}

func (w *MockWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.FlushCalls++
	return nil
}
func (w *MockWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.SyncCalls++
	return nil
}
func (w *MockWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.CloseCalls++
	return nil
}
func (w *MockWriter) Format() string { return "mock" }

// MockSampler implements Sampler for testing
type MockSampler struct {
	Topic string
}

func (s *MockSampler) Sample(ctx context.Context, index int) (string, error) {
	return s.Topic, nil
}
