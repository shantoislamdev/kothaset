// Package generator implements the core dataset generation engine.
package generator

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/shantoislamdev/kothaset/internal/output"
	"github.com/shantoislamdev/kothaset/internal/provider"
	"github.com/shantoislamdev/kothaset/internal/schema"
)

const cacheDir = ".kothaset"

// getCheckpointPath returns the path for the checkpoint file in the cache directory
func getCheckpointPath(outputPath string) string {
	// Create a safe filename from the output path by replacing path separators
	baseName := filepath.Base(outputPath)
	checkpointFile := baseName + ".checkpoint"
	return filepath.Join(cacheDir, checkpointFile)
}

// Config contains all settings for dataset generation
type Config struct {
	// Target
	NumSamples   int    `yaml:"num_samples" json:"num_samples"`
	Schema       string `yaml:"schema" json:"schema"`
	OutputPath   string `yaml:"output_path" json:"output_path"`
	OutputFormat string `yaml:"output_format" json:"output_format"` // jsonl, json

	// Provider
	Provider string `yaml:"provider" json:"provider"`
	Model    string `yaml:"model" json:"model"`

	// Generation parameters
	SystemPrompt string  `yaml:"system_prompt,omitempty" json:"system_prompt,omitempty"`
	Temperature  float64 `yaml:"temperature" json:"temperature"`
	MaxTokens    int     `yaml:"max_tokens" json:"max_tokens"`
	TopP         float64 `yaml:"top_p,omitempty" json:"top_p,omitempty"`

	// Reproducibility
	Seed          *int64 `yaml:"seed,omitempty" json:"seed,omitempty"`
	RandomSeed    bool   `yaml:"random_seed,omitempty" json:"random_seed,omitempty"` // Generate new random seed per request
	Deterministic bool   `yaml:"deterministic" json:"deterministic"`

	// Concurrency
	Workers   int `yaml:"workers" json:"workers"`
	BatchSize int `yaml:"batch_size" json:"batch_size"`

	// Resilience
	MaxRetries      int           `yaml:"max_retries" json:"max_retries"`
	RetryDelay      time.Duration `yaml:"retry_delay" json:"retry_delay"`
	CheckpointEvery int           `yaml:"checkpoint_every" json:"checkpoint_every"`
	ResumeFrom      string        `yaml:"resume_from,omitempty" json:"resume_from,omitempty"`

	// Input file for topics/seeds (required)
	InputFile string `yaml:"input_file" json:"input_file"`

	// Variables for prompt templates
	Variables map[string]any `yaml:"variables,omitempty" json:"variables,omitempty"`

	// Context from context.yaml (free-form paragraphs)
	UserContext     string `yaml:"user_context,omitempty" json:"user_context,omitempty"`
	UserInstruction string `yaml:"user_instruction,omitempty" json:"user_instruction,omitempty"`
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		NumSamples:   100,
		Schema:       "instruction",
		OutputFormat: "jsonl",
		Temperature:  0.7,
		// MaxTokens:       2048, // Removed default
		Workers:         4,
		MaxRetries:      3,
		RetryDelay:      time.Second * 2,
		CheckpointEvery: 50,
	}
}

// Result contains the outcome of a generation run
type Result struct {
	TotalSamples    int `json:"total_samples"`
	SuccessCount    int `json:"success_count"`
	FailedCount     int `json:"failed_count"`
	DuplicatesFound int `json:"duplicates_found"`
	TotalTokens     int `json:"total_tokens"`

	Duration       time.Duration `json:"duration"`
	OutputPath     string        `json:"output_path"`
	CheckpointPath string        `json:"checkpoint_path,omitempty"`
}

// Progress represents the current generation progress
type Progress struct {
	Total      int     `json:"total"`
	Completed  int     `json:"completed"`
	Failed     int     `json:"failed"`
	InProgress int     `json:"in_progress"`
	Percentage float64 `json:"percentage"`
	TokensUsed int     `json:"tokens_used"`

	ETA       time.Duration `json:"eta"`
	SamplesPS float64       `json:"samples_per_second"`
}

// ProgressCallback is called with progress updates
type ProgressCallback func(Progress)

// Generator orchestrates dataset generation
type Generator struct {
	config   Config
	provider provider.Provider
	schema   schema.Schema
	sampler  Sampler

	// State - only store counts, not samples (to prevent memory leaks)
	completed  int32
	failed     int32
	tokensUsed int64

	// Callbacks
	onProgress ProgressCallback

	// Output
	writer output.Writer
}

// New creates a new generator
func New(cfg Config, prov provider.Provider, sch schema.Schema) *Generator {
	return &Generator{
		config:   cfg,
		provider: prov,
		schema:   sch,
	}
}

// SetProgressCallback sets the progress callback
func (g *Generator) SetProgressCallback(cb ProgressCallback) {
	g.onProgress = cb
}

// SetSampler sets the seed sampler for topics
func (g *Generator) SetSampler(s Sampler) {
	g.sampler = s
}

// SetWriter sets the output writer
func (g *Generator) SetWriter(w output.Writer) {
	g.writer = w
}

// Run executes the generation process
func (g *Generator) Run(ctx context.Context) (*Result, error) {
	startTime := time.Now()

	// Load checkpoint if resuming
	if g.config.ResumeFrom != "" {
		checkpoint, err := LoadCheckpoint(g.config.ResumeFrom)
		if err != nil {
			return nil, fmt.Errorf("failed to load checkpoint: %w", err)
		}
		// Resume from checkpoint count - samples already written to output file
		atomic.StoreInt32(&g.completed, int32(checkpoint.Completed))
		atomic.StoreInt32(&g.failed, int32(checkpoint.Failed))
		atomic.StoreInt64(&g.tokensUsed, int64(checkpoint.TokensUsed))
	}

	// Open output writer if not already set
	if g.writer == nil {
		return nil, fmt.Errorf("output writer not set - call SetWriter() first")
	}

	// Ensure sampler is set
	if g.sampler == nil {
		return nil, fmt.Errorf("sampler not set: input file is mandatory")
	}
	if err := g.writer.Open(g.config.OutputPath); err != nil {
		return nil, fmt.Errorf("failed to open output: %w", err)
	}
	defer g.writer.Close()

	// Calculate remaining samples
	remaining := g.config.NumSamples - int(atomic.LoadInt32(&g.completed))

	// Create worker pool
	pool := NewWorkerPool(g.config.Workers)

	// Submit work
	results := make(chan *workerResult, remaining)
	var wg sync.WaitGroup
	checkpointCounter := 0

	for i := 0; i < remaining; i++ {
		select {
		case <-ctx.Done():
			goto cleanup
		default:
		}

		// Acquire a worker slot *before* spawning the goroutine
		// This provides backpressure and prevents spawning millions of goroutines
		pool.Acquire()

		wg.Add(1)
		sampleIndex := int(atomic.LoadInt32(&g.completed)) + i

		go func(idx int) {
			defer wg.Done()
			defer pool.Release()

			result := g.generateSample(ctx, idx)
			results <- result
		}(sampleIndex)
	}

	// Collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		if result.err != nil {
			atomic.AddInt32(&g.failed, 1)
		} else {
			// Write to output immediately - don't store in memory to prevent memory leaks
			if err := g.writer.Write(result.sample); err != nil {
				return nil, fmt.Errorf("failed to write sample: %w", err)
			}

			atomic.AddInt32(&g.completed, 1)
			atomic.AddInt64(&g.tokensUsed, int64(result.tokens))
		}

		// Update progress
		g.reportProgress(startTime)

		// Checkpoint
		checkpointCounter++
		if g.config.CheckpointEvery > 0 && checkpointCounter >= g.config.CheckpointEvery {
			if err := g.saveCheckpoint(); err != nil {
				// Log but don't fail
				fmt.Fprintf(os.Stderr, "Warning: failed to save checkpoint: %v\n", err)
			}
			checkpointCounter = 0
		}
	}

cleanup:
	// Final checkpoint
	if err := g.saveCheckpoint(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save final checkpoint: %v\n", err)
	}

	duration := time.Since(startTime)
	tokens := int(atomic.LoadInt64(&g.tokensUsed))

	return &Result{
		TotalSamples: g.config.NumSamples,
		SuccessCount: int(atomic.LoadInt32(&g.completed)),
		FailedCount:  int(atomic.LoadInt32(&g.failed)),
		TotalTokens:  tokens,

		Duration:   duration,
		OutputPath: g.config.OutputPath,
	}, nil
}

type workerResult struct {
	sample *schema.Sample
	tokens int
	err    error
}

// generateRandomSeed creates a cryptographically secure random seed
func generateRandomSeed() int64 {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		// Fallback to time-based seed if crypto/rand fails
		return time.Now().UnixNano()
	}
	return int64(binary.BigEndian.Uint64(b[:]))
}

func (g *Generator) generateSample(ctx context.Context, index int) *workerResult {
	// Build prompt options
	opts := schema.PromptOptions{
		Variables:       g.config.Variables,
		UserContext:     g.config.UserContext,
		UserInstruction: g.config.UserInstruction,
	}

	// Get topic from sampler if available
	if g.sampler != nil {
		topic, err := g.sampler.Sample(ctx, index)
		if err == nil {
			opts.Topic = topic
		}
	}

	// Generate prompt
	prompt, err := g.schema.GeneratePrompt(ctx, opts)
	if err != nil {
		return &workerResult{err: fmt.Errorf("failed to generate prompt: %w", err)}
	}

	// Determine seed for this request
	var requestSeed *int64
	if g.config.RandomSeed {
		// Generate a new random seed for each request
		seed := generateRandomSeed()
		requestSeed = &seed
	} else {
		// Use the fixed seed (may be nil)
		requestSeed = g.config.Seed
	}

	// Build request
	req := provider.GenerationRequest{
		Messages: []provider.Message{
			{Role: "user", Content: prompt},
		},
		SystemPrompt: g.config.SystemPrompt,
		Temperature:  g.config.Temperature,
		MaxTokens:    g.config.MaxTokens,
		TopP:         g.config.TopP,
		Seed:         requestSeed,
	}

	// Execute with retries
	var resp *provider.GenerationResponse
	var lastErr error
	for attempt := 0; attempt <= g.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return &workerResult{err: ctx.Err()}
			case <-time.After(g.config.RetryDelay):
			}
		}

		resp, lastErr = g.provider.Generate(ctx, req)
		if lastErr == nil {
			break
		}

		if !provider.IsRetryableError(lastErr) {
			return &workerResult{err: lastErr}
		}
	}

	if lastErr != nil {
		return &workerResult{err: lastErr}
	}

	// Parse response
	sample, err := g.schema.ParseResponse(resp.Content)
	if err != nil {
		return &workerResult{err: fmt.Errorf("failed to parse response: %w", err)}
	}

	// Set metadata
	sample.ID = uuid.New().String()
	sample.Metadata = schema.SampleMetadata{
		GeneratedAt: time.Now(),
		Provider:    g.provider.Name(),
		Model:       resp.Model,
		Temperature: g.config.Temperature,
		TokensUsed:  resp.Usage.TotalTokens,
		Latency:     resp.Latency,
		Topic:       opts.Topic,
	}

	// Validate
	if err := g.schema.ValidateSample(sample); err != nil {
		return &workerResult{err: fmt.Errorf("sample validation failed: %w", err)}
	}

	return &workerResult{
		sample: sample,
		tokens: resp.Usage.TotalTokens,
	}
}

func (g *Generator) reportProgress(startTime time.Time) {
	if g.onProgress == nil {
		return
	}

	completed := int(atomic.LoadInt32(&g.completed))
	failed := int(atomic.LoadInt32(&g.failed))
	tokens := int(atomic.LoadInt64(&g.tokensUsed))

	elapsed := time.Since(startTime)
	samplesPS := float64(completed) / elapsed.Seconds()

	remaining := g.config.NumSamples - completed - failed
	var eta time.Duration
	if samplesPS > 0 {
		eta = time.Duration(float64(remaining)/samplesPS) * time.Second
	}

	g.onProgress(Progress{
		Total:      g.config.NumSamples,
		Completed:  completed,
		Failed:     failed,
		Percentage: float64(completed) / float64(g.config.NumSamples) * 100,
		TokensUsed: tokens,

		ETA:       eta,
		SamplesPS: samplesPS,
	})
}

func (g *Generator) saveCheckpoint() error {
	cp := &Checkpoint{
		Timestamp:  time.Now(),
		Config:     g.config,
		Completed:  int(atomic.LoadInt32(&g.completed)),
		Failed:     int(atomic.LoadInt32(&g.failed)),
		TokensUsed: int(atomic.LoadInt64(&g.tokensUsed)),
	}

	return SaveCheckpoint(cp, getCheckpointPath(g.config.OutputPath))
}

// Checkpoint represents saved generation state
type Checkpoint struct {
	Timestamp  time.Time `json:"timestamp"`
	Config     Config    `json:"config"`
	Completed  int       `json:"completed"`
	Failed     int       `json:"failed"`
	TokensUsed int       `json:"tokens_used"`
}

// SaveCheckpoint saves a checkpoint to disk
func SaveCheckpoint(cp *Checkpoint, path string) error {
	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(cp, "", "  ")
	if err != nil {
		return err
	}

	// Write atomically
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

// LoadCheckpoint loads a checkpoint from disk
func LoadCheckpoint(path string) (*Checkpoint, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cp Checkpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return nil, err
	}
	return &cp, nil
}
