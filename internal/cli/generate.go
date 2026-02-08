package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/shantoislamdev/kothaset/internal/generator"
	"github.com/shantoislamdev/kothaset/internal/output"
	"github.com/shantoislamdev/kothaset/internal/provider"
	"github.com/shantoislamdev/kothaset/internal/schema"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a dataset using LLM",
	Long: `Generate a high-quality dataset using a large language model as the teacher.

The generate command creates samples according to the specified schema and
writes them to the output file in the chosen format.

Examples:
  # Generate 100 instruction-response pairs
  kothaset generate -n 100 -s instruction --seed 42 -o dataset.jsonl

  # Generate with custom provider and seed file
  kothaset generate -n 1000 -p openai -s chat --seed 123 --seeds topics.txt -o chat_data.jsonl

  # Resume interrupted generation
  kothaset generate --resume checkpoint.json`,
	RunE: runGenerate,
}

var (
	// Generate command flags
	genSchema       string
	genProvider     string
	genOutput       string
	genFormat       string
	genCount        int
	genWorkers      int
	genSeed         int64
	genSeedFile     string
	genResume       string
	genDryRun       bool
	genModel        string
	genTemp         float64
	genMaxTokens    int
	genSystemPrompt string
)

func init() {
	// Required flags
	generateCmd.Flags().IntVarP(&genCount, "count", "n", 100, "number of samples to generate")
	generateCmd.Flags().StringVarP(&genOutput, "output", "o", "", "output file path (required)")
	generateCmd.MarkFlagRequired("output")

	// Schema and provider
	generateCmd.Flags().StringVarP(&genSchema, "schema", "s", "instruction", "dataset schema (instruction, chat, preference, classification)")
	generateCmd.Flags().StringVarP(&genProvider, "provider", "p", "", "LLM provider (default: from config)")
	generateCmd.Flags().StringVarP(&genModel, "model", "m", "", "model to use (default: from config)")

	// Output format
	generateCmd.Flags().StringVarP(&genFormat, "format", "f", "jsonl", "output format (jsonl, parquet, hf)")

	// Generation parameters
	generateCmd.Flags().Float64Var(&genTemp, "temperature", 0.7, "sampling temperature")
	generateCmd.Flags().IntVar(&genMaxTokens, "max-tokens", 2048, "maximum tokens per response")
	generateCmd.Flags().StringVar(&genSystemPrompt, "system-prompt", "", "custom system prompt")

	// Concurrency and workers
	generateCmd.Flags().IntVarP(&genWorkers, "workers", "w", 4, "number of concurrent workers")

	// Reproducibility
	generateCmd.Flags().Int64Var(&genSeed, "seed", 0, "random seed for reproducibility (required)")
	generateCmd.MarkFlagRequired("seed")
	generateCmd.Flags().StringVar(&genSeedFile, "seeds", "", "path to seed/topic file for diversity")

	// Resumability
	generateCmd.Flags().StringVar(&genResume, "resume", "", "resume from checkpoint file")

	// Dry run
	generateCmd.Flags().BoolVar(&genDryRun, "dry-run", false, "validate configuration without generating")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Get provider config
	providerName := getProviderName()
	providerCfg, err := cfg.GetProvider(providerName)
	if err != nil {
		return fmt.Errorf("provider %q not configured: %w", providerName, err)
	}

	// Model to use (override if specified)
	model := providerCfg.Model
	if genModel != "" {
		model = genModel
	}

	// Get schema
	sch, err := schema.Get(genSchema)
	if err != nil {
		return fmt.Errorf("schema %q not found: %w", genSchema, err)
	}

	// Dry run - just validate config without creating provider
	if genDryRun {
		fmt.Println("✓ Configuration valid")
		fmt.Printf("  Schema:      %s\n", genSchema)
		fmt.Printf("  Provider:    %s\n", providerName)
		fmt.Printf("  Model:       %s\n", model)
		fmt.Printf("  Count:       %d\n", genCount)
		fmt.Printf("  Output:      %s\n", genOutput)
		fmt.Printf("  Format:      %s\n", genFormat)
		fmt.Printf("  Workers:     %d\n", genWorkers)
		fmt.Printf("  Temperature: %.2f\n", genTemp)
		if genSeedFile != "" {
			fmt.Printf("  Seed file:   %s\n", genSeedFile)
		}
		return nil
	}

	// Override model in config for provider creation
	if genModel != "" {
		providerCfg.Model = genModel
	}

	// Create provider
	prov, err := provider.GetOrCreate(providerCfg)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Build generator config
	genCfg := generator.Config{
		NumSamples:      genCount,
		Schema:          genSchema,
		OutputPath:      genOutput,
		OutputFormat:    genFormat,
		Provider:        providerName,
		Model:           model,
		SystemPrompt:    genSystemPrompt,
		Temperature:     genTemp,
		MaxTokens:       genMaxTokens,
		Seed:            genSeed,
		Workers:         genWorkers,
		MaxRetries:      3,
		RetryDelay:      time.Second * 2,
		CheckpointEvery: 50,
		ResumeFrom:      genResume,
		SeedFile:        genSeedFile,
	}

	// Create generator
	gen := generator.New(genCfg, prov, sch)

	// Create and set output writer
	writer, err := output.NewWriter(genFormat, sch)
	if err != nil {
		return fmt.Errorf("failed to create output writer: %w", err)
	}
	gen.SetWriter(writer)

	// Setup sampler if seed file provided
	if genSeedFile != "" {
		sampler, err := generator.NewFileSampler(genSeedFile)
		if err != nil {
			return fmt.Errorf("failed to load seed file: %w", err)
		}
		gen.SetSampler(sampler)
		fmt.Printf("Loaded %d topics from seed file\n", sampler.Count())
	} else {
		// Use random sampler for variety
		gen.SetSampler(generator.NewRandomSampler(genSeed))
	}

	// Setup progress callback
	startTime := time.Now()
	gen.SetProgressCallback(func(p generator.Progress) {
		elapsed := time.Since(startTime)
		_ = elapsed // suppress unused warning
		fmt.Printf("\r[%3.0f%%] %d/%d samples | %d tokens | %.1f/min | ETA: %s    ",
			p.Percentage,
			p.Completed,
			p.Total,
			p.TokensUsed,
			p.SamplesPS*60,
			formatDuration(p.ETA),
		)
	})

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n⚠ Interrupted - saving checkpoint...")
		cancel()
	}()

	// Print header
	fmt.Printf("Generating %d samples using %s (%s)\n", genCount, providerName, providerCfg.Model)
	fmt.Printf("Schema: %s | Output: %s\n\n", genSchema, genOutput)

	// Run generation
	result, err := gen.Run(ctx)
	fmt.Println() // New line after progress

	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	// Print results
	fmt.Println("\n✓ Generation complete!")
	fmt.Printf("  Samples:      %d successful, %d failed\n", result.SuccessCount, result.FailedCount)
	fmt.Printf("  Tokens:       %d\n", result.TotalTokens)
	fmt.Printf("  Duration:     %s\n", formatDuration(result.Duration))
	fmt.Printf("  Output:       %s\n", result.OutputPath)

	return nil
}

func getProviderName() string {
	if genProvider != "" {
		return genProvider
	}
	if cfg != nil && cfg.Global.DefaultProvider != "" {
		return cfg.Global.DefaultProvider
	}
	return "openai"
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}
