package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
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
  kothaset generate -n 100 --seed 42 -o dataset.jsonl

  # Generate with custom provider and input file
  kothaset generate -n 1000 -p openai -s chat --seed 123 -i topics.txt -o chat_data.jsonl

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
	genInputFile    string
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
	generateCmd.Flags().StringVarP(&genSchema, "schema", "s", "", "dataset schema (default: from config)")
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
	generateCmd.Flags().Int64Var(&genSeed, "seed", 0, "random seed for reproducibility")
	// generateCmd.MarkFlagRequired("seed") // Optional now
	generateCmd.Flags().StringVarP(&genInputFile, "input", "i", "", "path to input file for topics/seeds (required)")
	generateCmd.MarkFlagRequired("input")

	// Resumability
	generateCmd.Flags().StringVar(&genResume, "resume", "", "resume from checkpoint file")

	// Dry run
	generateCmd.Flags().BoolVar(&genDryRun, "dry-run", false, "validate configuration without generating")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Ensure input file is provided (handle empty string case)
	if genInputFile == "" {
		return fmt.Errorf("input file is required (use -i or --input)")
	}

	// Auto-generate seed if not provided
	if !cmd.Flags().Changed("seed") {
		genSeed = time.Now().UnixNano()
		fmt.Printf("ℹ No seed provided, using random seed: %d\n", genSeed)
	}

	// Get provider name from flag or config
	providerName := genProvider
	if providerName == "" {
		providerName = cfg.Global.Provider
	}

	// Get provider config from secrets
	providerCfg, err := secrets.GetProvider(providerName)
	if err != nil {
		return fmt.Errorf("provider %q not configured in .secrets.yaml: %w", providerName, err)
	}

	// Get schema name from flag or config
	schemaName := genSchema
	if schemaName == "" {
		schemaName = cfg.Global.Schema
	}

	// Get model from flag or config
	model := genModel
	if model == "" {
		model = cfg.Global.Model
	}

	// Get schema
	sch, err := schema.Get(schemaName)
	if err != nil {
		return fmt.Errorf("schema %q not found: %w", schemaName, err)
	}

	// Dry run - just validate config without creating provider
	if genDryRun {
		fmt.Println("✓ Configuration valid")
		fmt.Printf("  Schema:      %s\n", schemaName)
		fmt.Printf("  Provider:    %s\n", providerName)
		fmt.Printf("  Model:       %s\n", model)
		fmt.Printf("  Count:       %d\n", genCount)
		fmt.Printf("  Output:      %s\n", genOutput)
		fmt.Printf("  Format:      %s\n", genFormat)
		fmt.Printf("  Workers:     %d\n", genWorkers)
		fmt.Printf("  Temperature: %.2f\n", genTemp)
		if genInputFile != "" {
			fmt.Printf("  Input file:  %s\n", genInputFile)
		}
		if cfg.Context != "" {
			fmt.Println("  Context:     ✓ (from kothaset.yaml)")
		}
		if len(cfg.Instructions) > 0 {
			fmt.Println("  Instructions: ✓ (from kothaset.yaml)")
		}
		return nil
	}

	// Create provider config for generation
	provCfg := &provider.Config{
		Name:       providerCfg.Name,
		Type:       providerCfg.Type,
		BaseURL:    providerCfg.BaseURL,
		APIKey:     providerCfg.APIKey,
		Model:      model,
		MaxRetries: providerCfg.MaxRetries,
		Timeout:    providerCfg.Timeout.Duration,
		RateLimit:  providerCfg.RateLimit.RequestsPerMinute,
	}

	// Create provider
	prov, err := provider.GetOrCreate(provCfg)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Context and instructions from kothaset.yaml
	userContext := cfg.Context
	userInstruction := strings.Join(cfg.Instructions, "\n")

	if userContext != "" || len(cfg.Instructions) > 0 {
		fmt.Println("✓ Loaded context from kothaset.yaml")
	}

	// Build generator config
	genCfg := generator.Config{
		NumSamples:      genCount,
		Schema:          schemaName,
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
		InputFile:       genInputFile,
		UserContext:     userContext,
		UserInstruction: userInstruction,
	}

	// Create generator
	gen := generator.New(genCfg, prov, sch)

	// Create and set output writer
	writer, err := output.NewWriter(genFormat, sch)
	if err != nil {
		return fmt.Errorf("failed to create output writer: %w", err)
	}
	gen.SetWriter(writer)

	// Setup sampler from input file (mandatory)
	sampler, err := generator.NewFileSampler(genInputFile)
	if err != nil {
		return fmt.Errorf("failed to load input file %s: %w", genInputFile, err)
	}
	gen.SetSampler(sampler)

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n⚠ Received interrupt, saving checkpoint...")
		cancel()
	}()

	// Print generation info
	fmt.Printf("🚀 Generating %d samples using %s (%s)\n", genCount, providerName, model)
	fmt.Printf("   Schema: %s | Output: %s\n\n", schemaName, genOutput)

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

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}
