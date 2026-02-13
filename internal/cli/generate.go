package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/schollz/progressbar/v3"
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
  # Generate 100 instruction-response pairs with fixed seed
  kothaset generate -n 100 --seed 42 -o dataset.jsonl

  # Generate with random seed per request (maximizes diversity)
  kothaset generate -n 100 --seed random -i topics.txt -o dataset.jsonl

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
	genSeed         string
	genInputFile    string
	genResume       string
	genDryRun       bool
	genModel        string
	genTemp         float64
	genMaxTokens    int
	genSystemPrompt string
	genTimeout      string
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
	generateCmd.Flags().StringVarP(&genFormat, "format", "f", "", "output format (jsonl)")

	// Generation parameters
	generateCmd.Flags().Float64Var(&genTemp, "temperature", 0.7, "sampling temperature")
	generateCmd.Flags().IntVar(&genMaxTokens, "max-tokens", 0, "maximum tokens per response (0 = default)")
	generateCmd.Flags().StringVar(&genSystemPrompt, "system-prompt", "", "custom system prompt")
	generateCmd.Flags().StringVar(&genTimeout, "timeout", "", "maximum total generation time (e.g. '30m', '2h')")

	// Concurrency and workers
	generateCmd.Flags().IntVarP(&genWorkers, "workers", "w", 4, "number of concurrent workers")

	// Reproducibility
	generateCmd.Flags().StringVar(&genSeed, "seed", "", "random seed for reproducibility (use 'random' for client-side random seeds per request)")
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
	if hasParentPathTraversal(genOutput) {
		return fmt.Errorf("output path must not contain '..': %s", genOutput)
	}

	// Validate generation parameters
	if genCount <= 0 {
		return fmt.Errorf("--count must be >= 1, got %d", genCount)
	}
	if genTemp < 0 || genTemp > 2.0 {
		return fmt.Errorf("--temperature must be between 0 and 2.0, got %.2f", genTemp)
	}
	if genMaxTokens < 0 {
		return fmt.Errorf("--max-tokens must be >= 0, got %d", genMaxTokens)
	}
	if genWorkers <= 0 {
		return fmt.Errorf("--workers must be >= 1, got %d", genWorkers)
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

	// Resolve output format
	if genFormat == "" {
		genFormat = cfg.Global.OutputFormat
		if genFormat == "" {
			genFormat = "jsonl" // Hard default
		}
	}

	// Resolve max tokens
	if genMaxTokens == 0 {
		// Use global config if set
		if cfg.Global.MaxTokens > 0 {
			genMaxTokens = cfg.Global.MaxTokens
		}
		// Else remains 0 (unlimited/model default)
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
	defer provider.CloseAll()

	// Context and instructions from kothaset.yaml
	userContext := cfg.Context
	userInstruction := strings.Join(cfg.Instructions, "\n")

	if userContext != "" || len(cfg.Instructions) > 0 {
		fmt.Println("✓ Loaded context from kothaset.yaml")
	}

	// Handle seed (optional)
	// Supports: empty (no seed), "random" (different random per request), or a specific number (fixed seed)
	var seedPtr *int64
	var randomSeed bool
	if cmd.Flags().Changed("seed") || genSeed != "" {
		if genSeed == "random" {
			randomSeed = true
		} else if genSeed != "" {
			// Parse as int64 - fixed seed to be sent to AI
			var parsedSeed int64
			_, err := fmt.Sscanf(genSeed, "%d", &parsedSeed)
			if err != nil {
				return fmt.Errorf("invalid seed value %q: must be 'random' or a number", genSeed)
			}
			seedPtr = &parsedSeed
		}
	}

	// Determine checkpoint interval (flag > global config > default)
	checkpointEvery := cfg.Global.CheckpointEvery
	if checkpointEvery <= 0 {
		checkpointEvery = 10 // fallback default
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
		Seed:            seedPtr,    // Fixed seed sent to AI (nil if not specified)
		RandomSeed:      randomSeed, // When true, generates new random seed per request
		Workers:         genWorkers,
		RateLimit:       providerCfg.RateLimit.RequestsPerMinute,
		MaxRetries:      3,
		RetryDelay:      time.Second * 2,
		CheckpointEvery: checkpointEvery,
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
	// Now supports inline strings too
	sampler, err := generator.NewSampler(genInputFile)
	if err != nil {
		return fmt.Errorf("failed to load input (file or inline) %q: %w", genInputFile, err)
	}
	gen.SetSampler(sampler)

	// Create context with optional overall timeout
	var ctx context.Context
	var cancel context.CancelFunc
	if genTimeout != "" {
		dur, err := time.ParseDuration(genTimeout)
		if err != nil {
			return fmt.Errorf("invalid --timeout value %q: %w", genTimeout, err)
		}
		ctx, cancel = context.WithTimeout(context.Background(), dur)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	// Handle signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	go func() {
		select {
		case <-sigCh:
			fmt.Println("\n⚠ Received interrupt, saving checkpoint...")
			cancel()
		case <-ctx.Done():
			// Generation finished normally or was cancelled.
		}
	}()

	// Create progress bar
	bar := progressbar.NewOptions(genCount,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(40),
		progressbar.OptionSetDescription("Generating samples..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]█[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: "[dark_gray]░[reset]",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionOnCompletion(func() {
			fmt.Println()
		}),
	)

	// Set progress callback to update the bar
	gen.SetProgressCallback(func(p generator.Progress) {
		bar.Set(p.Completed)
	})

	// Print generation info
	fmt.Printf("🚀 Generating %d samples using %s (%s)\n", genCount, providerName, model)
	fmt.Printf("   Schema: %s | Output: %s\n\n", schemaName, genOutput)

	// Run generation
	result, err := gen.Run(ctx)
	bar.Finish()
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

func hasParentPathTraversal(path string) bool {
	parts := strings.FieldsFunc(path, func(r rune) bool {
		return r == '/' || r == '\\'
	})
	for _, part := range parts {
		if part == ".." {
			return true
		}
	}
	return false
}
