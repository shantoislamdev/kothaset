package generator

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/shantoislamdev/kothaset/internal/provider"
	"github.com/shantoislamdev/kothaset/internal/schema"
)

type indexTrackingSampler struct {
	mu      sync.Mutex
	indices []int
}

func (s *indexTrackingSampler) Sample(ctx context.Context, index int) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.indices = append(s.indices, index)
	return "tracked-topic", nil
}

func (s *indexTrackingSampler) Indices() []int {
	s.mu.Lock()
	defer s.mu.Unlock()
	cloned := make([]int, len(s.indices))
	copy(cloned, s.indices)
	return cloned
}

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

func TestGenerator_Run_ResumeTopicIndex(t *testing.T) {
	tmpDir := t.TempDir()
	checkpointPath := filepath.Join(tmpDir, "resume.checkpoint")

	cp := &Checkpoint{
		Timestamp:  time.Now(),
		Config:     DefaultConfig(),
		Completed:  50,
		Failed:     0,
		TokensUsed: 0,
	}
	if err := SaveCheckpoint(cp, checkpointPath); err != nil {
		t.Fatalf("failed to save checkpoint: %v", err)
	}

	cfg := DefaultConfig()
	cfg.NumSamples = 55
	cfg.Workers = 4
	cfg.ResumeFrom = checkpointPath
	cfg.OutputPath = filepath.Join(tmpDir, "out.jsonl")

	prov := &MockProvider{Response: `{"instruction":"this is long enough","output":"this is long enough output"}`}
	gen := New(cfg, prov, schema.NewInstructionSchema())
	sampler := &indexTrackingSampler{}
	gen.SetSampler(sampler)
	gen.SetWriter(&MockWriter{})

	res, err := gen.Run(context.Background())
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if res.SuccessCount != 55 {
		t.Fatalf("expected success count 55, got %d", res.SuccessCount)
	}

	indices := sampler.Indices()
	if len(indices) != 5 {
		t.Fatalf("expected 5 sampled indices, got %d (%v)", len(indices), indices)
	}
	sort.Ints(indices)
	expected := []int{50, 51, 52, 53, 54}
	if !reflect.DeepEqual(indices, expected) {
		t.Fatalf("expected sampled indices %v, got %v", expected, indices)
	}
}

func TestGenerator_Run_ResumeSchemaMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	checkpointPath := filepath.Join(tmpDir, "resume.checkpoint")

	cp := &Checkpoint{
		SchemaVersion: checkpointVersion,
		Timestamp:     time.Now(),
		Config: Config{
			Schema: "chat",
		},
		Completed: 2,
	}
	if err := SaveCheckpoint(cp, checkpointPath); err != nil {
		t.Fatalf("failed to save checkpoint: %v", err)
	}

	cfg := DefaultConfig()
	cfg.NumSamples = 5
	cfg.Schema = "instruction"
	cfg.ResumeFrom = checkpointPath
	cfg.OutputPath = filepath.Join(tmpDir, "out.jsonl")

	gen := New(cfg, &MockProvider{Response: `{"instruction":"this is long enough","output":"this is long enough output"}`}, schema.NewInstructionSchema())
	gen.SetSampler(&MockSampler{Topic: "topic"})
	gen.SetWriter(&MockWriter{})

	_, err := gen.Run(context.Background())
	if err == nil {
		t.Fatalf("expected resume schema mismatch error")
	}
	if !strings.Contains(err.Error(), "resume schema mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerator_Run_ResumeOutputMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	checkpointPath := filepath.Join(tmpDir, "resume.checkpoint")

	cp := &Checkpoint{
		SchemaVersion: checkpointVersion,
		Timestamp:     time.Now(),
		Config: Config{
			Schema:     "instruction",
			OutputPath: filepath.Join(tmpDir, "one.jsonl"),
		},
		Completed: 2,
	}
	if err := SaveCheckpoint(cp, checkpointPath); err != nil {
		t.Fatalf("failed to save checkpoint: %v", err)
	}

	cfg := DefaultConfig()
	cfg.NumSamples = 5
	cfg.Schema = "instruction"
	cfg.ResumeFrom = checkpointPath
	cfg.OutputPath = filepath.Join(tmpDir, "two.jsonl")

	gen := New(cfg, &MockProvider{Response: `{"instruction":"this is long enough","output":"this is long enough output"}`}, schema.NewInstructionSchema())
	gen.SetSampler(&MockSampler{Topic: "topic"})
	gen.SetWriter(&MockWriter{})

	_, err := gen.Run(context.Background())
	if err == nil {
		t.Fatalf("expected resume output mismatch error")
	}
	if !strings.Contains(err.Error(), "resume output mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerator_Run_ResumeCompletedExceedsRequested(t *testing.T) {
	tmpDir := t.TempDir()
	checkpointPath := filepath.Join(tmpDir, "resume.checkpoint")

	cp := &Checkpoint{
		SchemaVersion: checkpointVersion,
		Timestamp:     time.Now(),
		Config: Config{
			Schema:     "instruction",
			OutputPath: filepath.Join(tmpDir, "out.jsonl"),
		},
		Completed: 10,
	}
	if err := SaveCheckpoint(cp, checkpointPath); err != nil {
		t.Fatalf("failed to save checkpoint: %v", err)
	}

	cfg := DefaultConfig()
	cfg.NumSamples = 5
	cfg.Schema = "instruction"
	cfg.ResumeFrom = checkpointPath
	cfg.OutputPath = filepath.Join(tmpDir, "out.jsonl")

	gen := New(cfg, &MockProvider{Response: `{"instruction":"this is long enough","output":"this is long enough output"}`}, schema.NewInstructionSchema())
	gen.SetSampler(&MockSampler{Topic: "topic"})
	gen.SetWriter(&MockWriter{})

	_, err := gen.Run(context.Background())
	if err == nil {
		t.Fatalf("expected resume count mismatch error")
	}
	if !strings.Contains(err.Error(), "resume count mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSaveCheckpoint_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "checkpoint.json")

	first := &Checkpoint{
		SchemaVersion: checkpointVersion,
		Timestamp:     time.Now(),
		Config:        DefaultConfig(),
		Completed:     1,
	}
	if err := SaveCheckpoint(first, path); err != nil {
		t.Fatalf("first save failed: %v", err)
	}

	second := &Checkpoint{
		SchemaVersion: checkpointVersion,
		Timestamp:     time.Now(),
		Config:        DefaultConfig(),
		Completed:     2,
	}
	if err := SaveCheckpoint(second, path); err != nil {
		t.Fatalf("second save failed: %v", err)
	}

	got, err := LoadCheckpoint(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if got.Completed != 2 {
		t.Fatalf("expected overwritten checkpoint completed=2, got %d", got.Completed)
	}
}

func TestGenerator_Run_WriteError_Graceful(t *testing.T) {
	cfg := DefaultConfig()
	cfg.NumSamples = 10
	cfg.Workers = 4

	prov := &MockProvider{Response: `{"instruction":"this is long enough","output":"this is long enough output"}`}
	writer := &MockWriter{FailOnWrite: true, FailAfter: 3}
	gen := New(cfg, prov, schema.NewInstructionSchema())
	gen.SetSampler(&MockSampler{Topic: "test"})
	gen.SetWriter(writer)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := gen.Run(ctx)
	if err == nil {
		t.Fatal("expected write error, got nil")
	}
	if res == nil {
		t.Fatal("expected partial result on write error, got nil")
	}
	if res.SuccessCount < 1 || res.SuccessCount > 3 {
		t.Fatalf("expected partial successes between 1 and 3, got %d", res.SuccessCount)
	}
	if writer.CloseCalls != 1 {
		t.Fatalf("expected writer close to be called once, got %d", writer.CloseCalls)
	}
}

func TestGenerator_Run_Cancellation_NoPanic(t *testing.T) {
	cfg := DefaultConfig()
	cfg.NumSamples = 100
	cfg.Workers = 8
	cfg.MaxRetries = 0

	prov := &MockProvider{
		Delay:    50 * time.Millisecond,
		Response: `{"instruction":"this is long enough","output":"this is long enough output"}`,
	}
	gen := New(cfg, prov, schema.NewInstructionSchema())
	gen.SetSampler(&MockSampler{Topic: "cancel-topic"})
	gen.SetWriter(&MockWriter{})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	done := make(chan struct{})
	var runErr error
	go func() {
		defer close(done)
		_, runErr = gen.Run(ctx)
	}()

	select {
	case <-done:
		if runErr != nil && !errors.Is(runErr, context.Canceled) {
			t.Fatalf("unexpected error on cancellation: %v", runErr)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("run did not terminate after cancellation (possible goroutine leak)")
	}
}

func TestGenerator_Run_ExponentialBackoff(t *testing.T) {
	gen := New(DefaultConfig(), &MockProvider{}, schema.NewInstructionSchema())
	gen.randFloat = func() float64 { return 0.5 } // no jitter (factor 1.0)
	gen.config.RetryDelay = 100 * time.Millisecond

	if d := gen.retryDelay(1, provider.NewProviderError(provider.ErrKindRateLimit, "retry", nil)); d != 100*time.Millisecond {
		t.Fatalf("attempt 1 delay mismatch: got %v", d)
	}
	if d := gen.retryDelay(2, provider.NewProviderError(provider.ErrKindRateLimit, "retry", nil)); d != 200*time.Millisecond {
		t.Fatalf("attempt 2 delay mismatch: got %v", d)
	}
	if d := gen.retryDelay(3, provider.NewProviderError(provider.ErrKindRateLimit, "retry", nil)); d != 400*time.Millisecond {
		t.Fatalf("attempt 3 delay mismatch: got %v", d)
	}

	if d := gen.retryDelay(20, provider.NewProviderError(provider.ErrKindRateLimit, "retry", nil)); d != 30*time.Second {
		t.Fatalf("delay should be capped at 30s, got %v", d)
	}

	retryAfterErr := provider.NewRateLimitError("retry later", 7)
	if d := gen.retryDelay(3, retryAfterErr); d != 7*time.Second {
		t.Fatalf("retry-after should override backoff, got %v", d)
	}
}

func TestGetCheckpointPath_UsesFullPath(t *testing.T) {
	p1 := getCheckpointPath(filepath.Join("one", "dataset.jsonl"), defaultCacheDir)
	p2 := getCheckpointPath(filepath.Join("two", "dataset.jsonl"), defaultCacheDir)
	if p1 == p2 {
		t.Fatalf("checkpoint path should differ for same basenames in different dirs: %s", p1)
	}
	if filepath.Base(p1) == "dataset.jsonl.checkpoint" || filepath.Base(p2) == "dataset.jsonl.checkpoint" {
		t.Fatalf("checkpoint filename should include transformed full path, got %q and %q", filepath.Base(p1), filepath.Base(p2))
	}
	if err := os.MkdirAll(filepath.Dir(p1), 0755); err != nil {
		t.Fatalf("failed to create checkpoint directory: %v", err)
	}
}
