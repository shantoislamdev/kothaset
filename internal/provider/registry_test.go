package provider

import (
	"context"
	"testing"
)

// MockProvider for testing registry
type MockProvider struct {
	name string
}

func (m *MockProvider) Generate(ctx context.Context, req GenerationRequest) (*GenerationResponse, error) {
	return nil, nil
}
func (m *MockProvider) GenerateStream(ctx context.Context, req GenerationRequest) (<-chan StreamChunk, error) {
	return nil, nil
}
func (m *MockProvider) Name() string                          { return m.name }
func (m *MockProvider) Type() string                          { return "mock" }
func (m *MockProvider) Model() string                         { return "mock-model" }
func (m *MockProvider) SupportedModels() []string             { return []string{"mock-model"} }
func (m *MockProvider) SupportsStreaming() bool               { return false }
func (m *MockProvider) SupportsBatching() bool                { return false }
func (m *MockProvider) Validate() error                       { return nil }
func (m *MockProvider) HealthCheck(ctx context.Context) error { return nil }
func (m *MockProvider) Close() error                          { return nil }

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	// Test registering a factory
	r.RegisterFactory("mock", func(cfg *Config) (Provider, error) {
		return &MockProvider{name: cfg.Name}, nil
	})

	// Test GetOrCreate with new provider
	cfg := &Config{
		Name: "test-provider",
		Type: "mock",
	}
	p, err := r.GetOrCreate(cfg)
	if err != nil {
		t.Fatalf("GetOrCreate failed: %v", err)
	}
	if p.Name() != "test-provider" {
		t.Errorf("Expected provider name test-provider, got %s", p.Name())
	}

	// Test GetOrCreate existing
	p2, err := r.GetOrCreate(cfg)
	if err != nil {
		t.Fatalf("GetOrCreate existing failed: %v", err)
	}
	if p != p2 {
		t.Error("Expected same provider instance")
	}

	// Test Get
	p3, err := r.Get("test-provider")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if p != p3 {
		t.Error("Expected same provider instance from Get")
	}

	// Test List
	list := r.List()
	if len(list) != 1 || list[0] != "test-provider" {
		t.Errorf("List failed, expected [test-provider], got %v", list)
	}

	// Test specific error for unknown factory
	unknownCfg := &Config{Name: "unknown", Type: "unknown"}
	_, err = r.GetOrCreate(unknownCfg)
	if err == nil {
		t.Error("Expected error for unknown factory, got nil")
	}

	// Test Close
	if err := r.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if len(r.List()) != 0 {
		t.Error("Registry should be empty after Close")
	}
}

func TestGlobalRegistry(t *testing.T) {
	// Just verify global functions map to a registry instance
	// We can't easily reset global state so we just check basic functionality
	// ensuring no panic
	list := List()
	_ = list
}
