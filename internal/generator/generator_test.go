package generator

import (
	"context"
	"testing"
	"time"

	"github.com/shantoislamdev/kothaset/internal/schema"
)

func TestGenerator_Run_Success(t *testing.T) {
	// Setup
	cfg := DefaultConfig()
	cfg.NumSamples = 5
	cfg.Workers = 2

	prov := &MockProvider{Response: `{"instruction": "this is a long enough instruction", "output": "this is a long enough output"}`}
	s := schema.NewInstructionSchema()
	gen := New(cfg, prov, s)

	// Set required components
	gen.SetSampler(&MockSampler{Topic: "test-topic"})
	writer := &MockWriter{}
	gen.SetWriter(writer)

	// Run
	ctx := context.Background()
	res, err := gen.Run(ctx)
	if err != nil {
		t.Fatalf("Generator.Run failed: %v", err)
	}

	// Verify
	if res.SuccessCount != 5 {
		t.Errorf("Expected 5 successes, got %d", res.SuccessCount)
	}
	if res.FailedCount != 0 {
		t.Errorf("Expected 0 failures, got %d", res.FailedCount)
	}
	if prov.Calls != 5 { // Or more if retries happened (shouldn't here)
		t.Errorf("Expected 5 provider calls, got %d", prov.Calls)
	}
	if len(writer.Samples) != 5 {
		t.Errorf("Expected 5 written samples, got %d", len(writer.Samples))
	}
}

func TestGenerator_Run_ProviderError(t *testing.T) {
	// Setup
	cfg := DefaultConfig()
	cfg.NumSamples = 2
	cfg.MaxRetries = 1
	cfg.RetryDelay = time.Millisecond // Fast retries for test

	prov := &MockProvider{ShouldFail: true}
	s := schema.NewInstructionSchema()
	gen := New(cfg, prov, s)

	gen.SetSampler(&MockSampler{Topic: "test"})
	gen.SetWriter(&MockWriter{})

	// Run
	ctx := context.Background()
	res, err := gen.Run(ctx)
	if err != nil {
		t.Fatalf("Run expected to complete (with failures), got error: %v", err)
	}

	if res.FailedCount != 2 {
		t.Errorf("Expected 2 failures, got %d", res.FailedCount)
	}
	if res.SuccessCount != 0 {
		t.Errorf("Expected 0 successes, got %d", res.SuccessCount)
	}
}

func TestGenerator_ProgressCallback(t *testing.T) {
	cfg := DefaultConfig()
	cfg.NumSamples = 2

	prov := &MockProvider{Response: `{"instruction":"this is long enough","output":"this is also long enough"}`}
	gen := New(cfg, prov, schema.NewInstructionSchema())
	gen.SetSampler(&MockSampler{Topic: "t"})
	gen.SetWriter(&MockWriter{})

	called := false
	gen.SetProgressCallback(func(p Progress) {
		called = true
		if p.Total != 2 {
			t.Errorf("Progress total expected 2, got %d", p.Total)
		}
	})

	gen.Run(context.Background())

	if !called {
		t.Error("Progress callback not called")
	}
}
