package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/shantoislamdev/kothaset/internal/provider"
	"github.com/shantoislamdev/kothaset/internal/schema"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration, schemas, or datasets",
	Long: `Validate various components of your KothaSet setup.

Subcommands:
  config   - Validate the configuration file
  schema   - Validate a schema definition
  dataset  - Validate an existing dataset`,
}

var validateConfigCmd = &cobra.Command{
	Use:   "config [path]",
	Short: "Validate configuration file",
	Long: `Validate a KothaSet configuration file for correctness.

If no path is provided, validates the default config resolution order.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("✓ Configuration is valid")
		if cfg != nil {
			fmt.Printf("  Version:   %s\n", cfg.Version)
			fmt.Printf("  Provider:  %s\n", cfg.Global.Provider)
			fmt.Printf("  Schema:    %s\n", cfg.Global.Schema)
			fmt.Printf("  Model:     %s\n", cfg.Global.Model)
		}
		if secrets != nil {
			fmt.Printf("  Configured providers: %d\n", len(secrets.Providers))
		}
		return nil
	},
}

var validateSchemaCmd = &cobra.Command{
	Use:   "schema <name|path>",
	Short: "Validate a schema definition",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		schemaName := args[0]

		// Get schema from registry
		sch, err := schema.Get(schemaName)
		if err != nil {
			return fmt.Errorf("schema not found: %s\nAvailable schemas: %v", schemaName, schema.List())
		}

		// Validate schema has required components
		var issues []string

		// Check name
		if sch.Name() == "" {
			issues = append(issues, "schema has no name")
		}

		// Check style
		if sch.Style() == "" {
			issues = append(issues, "schema has no style defined")
		}

		// Check fields
		fields := sch.Fields()
		if len(fields) == 0 {
			issues = append(issues, "schema has no fields defined")
		}

		// Check required fields
		requiredFields := sch.RequiredFields()
		for _, reqField := range requiredFields {
			found := false
			for _, f := range fields {
				if f.Name == reqField {
					found = true
					break
				}
			}
			if !found {
				issues = append(issues, fmt.Sprintf("required field '%s' not in field definitions", reqField))
			}
		}

		// Report results
		if len(issues) > 0 {
			fmt.Printf("✗ Schema '%s' has %d issue(s):\n", schemaName, len(issues))
			for _, issue := range issues {
				fmt.Printf("  - %s\n", issue)
			}
			return fmt.Errorf("schema validation failed")
		}

		fmt.Printf("✓ Schema '%s' is valid\n", schemaName)
		fmt.Printf("  Style:  %s\n", sch.Style())
		fmt.Printf("  Fields: %d (%d required)\n", len(fields), len(requiredFields))
		return nil
	},
}

var validateDatasetCmd = &cobra.Command{
	Use:   "dataset <path>",
	Short: "Validate an existing dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		// Check file exists
		info, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("cannot access file: %w", err)
		}

		// Detect format from extension
		format := detectFormat(filePath)
		if format != "jsonl" {
			return fmt.Errorf("unsupported format: %s (only .jsonl is currently supported)", filePath)
		}

		fmt.Printf("Validating dataset: %s\n", filePath)
		fmt.Printf("  Format: %s\n", format)
		fmt.Printf("  Size:   %d bytes\n", info.Size())

		// Read and validate based on format
		var rowCount int
		var parseErr error

		switch format {
		case "jsonl":
			rowCount, parseErr = validateJSONL(filePath)
		}

		if parseErr != nil {
			fmt.Printf("✗ Validation failed: %v\n", parseErr)
			return parseErr
		}

		fmt.Printf("✓ Valid dataset\n")
		fmt.Printf("  Rows: %d\n", rowCount)
		return nil
	},
}

// detectFormat returns the format string based on file extension
func detectFormat(path string) string {
	if hasExtension(path, ".jsonl") {
		return "jsonl"
	}
	return ""
}

// hasExtension checks if path ends with the given extension (case-insensitive)
func hasExtension(path, ext string) bool {
	return len(path) > len(ext) && strings.EqualFold(path[len(path)-len(ext):], ext)
}

// validateJSONL validates a JSONL file and returns row count
func validateJSONL(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Allow long lines (up to 10MB per line for large JSON objects)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)
	count := 0
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			return count, fmt.Errorf("line %d: invalid JSON: %w", lineNum, err)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		return count, fmt.Errorf("read error: %w", err)
	}
	return count, nil
}

func init() {
	validateCmd.AddCommand(validateConfigCmd)
	validateCmd.AddCommand(validateSchemaCmd)
	validateCmd.AddCommand(validateDatasetCmd)
}

// Schema command
var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Manage dataset schemas",
	Long: `List, show, create, and export dataset schemas.

Schemas define the structure and generation prompts for your datasets.`,
}

var schemaListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available schemas",
	RunE: func(cmd *cobra.Command, args []string) error {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSTYLE\tDESCRIPTION")
		fmt.Fprintln(w, "instruction\tinstruction\tAlpaca-style instruction-response pairs")
		fmt.Fprintln(w, "chat\tchat\tShareGPT-style multi-turn conversations")
		fmt.Fprintln(w, "preference\tpreference\tDPO/RLHF preference pairs (chosen/rejected)")
		fmt.Fprintln(w, "classification\tclassification\tText classification with labels")
		w.Flush()
		return nil
	},
}

var schemaShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show schema details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		schemaName := args[0]

		// Get schema from registry
		sch, err := schema.Get(schemaName)
		if err != nil {
			return fmt.Errorf("schema not found: %s\nRun 'kothaset schema list' to see available schemas", schemaName)
		}

		// Display schema metadata
		fmt.Printf("Name:        %s\n", sch.Name())
		fmt.Printf("Style:       %s\n", sch.Style())
		fmt.Printf("Description: %s\n", sch.Description())
		if sch.Version() != "" {
			fmt.Printf("Version:     %s\n", sch.Version())
		}

		// Display fields
		fields := sch.Fields()
		if len(fields) > 0 {
			fmt.Println("\nFields:")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "  NAME\tTYPE\tREQUIRED")
			for _, f := range fields {
				required := "no"
				if f.Required {
					required = "yes"
				}
				fmt.Fprintf(w, "  %s\t%s\t%s\n", f.Name, f.Type, required)
			}
			w.Flush()
		}

		// Display required fields summary
		requiredFields := sch.RequiredFields()
		if len(requiredFields) > 0 {
			fmt.Printf("\nRequired: %v\n", requiredFields)
		}

		return nil
	},
}

func init() {
	schemaCmd.AddCommand(schemaListCmd)
	schemaCmd.AddCommand(schemaShowCmd)
}

// Provider command
var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Manage LLM providers",
	Long:  `List and test LLM provider configurations.`,
}

var providerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tSTATUS")

		if secrets != nil && len(secrets.Providers) > 0 {
			for _, p := range secrets.Providers {
				status := "configured"
				if p.APIKey == "" {
					status = "no api key"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", p.Name, p.Type, status)
			}
		} else {
			fmt.Fprintln(w, "openai\topenai\tdefault")
		}
		w.Flush()
		return nil
	},
}

var providerTestCmd = &cobra.Command{
	Use:   "test <name>",
	Short: "Test provider connection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		providerName := args[0]

		// Find provider config by name
		var providerCfg *provider.Config
		if secrets != nil {
			for _, p := range secrets.Providers {
				if p.Name == providerName {
					timeout := 30 * time.Second
					if p.Timeout.Duration > 0 {
						timeout = p.Timeout.Duration
					}
					providerCfg = &provider.Config{
						Name:       p.Name,
						Type:       p.Type,
						BaseURL:    p.BaseURL,
						APIKey:     p.APIKey,
						MaxRetries: p.MaxRetries,
						Timeout:    timeout,
						Headers:    p.Headers,
					}
					break
				}
			}
		}

		if providerCfg == nil {
			return fmt.Errorf("provider not found: %s\nRun 'kothaset provider list' to see available providers", providerName)
		}

		if providerCfg.APIKey == "" {
			return fmt.Errorf("provider %s has no API key configured", providerName)
		}

		fmt.Printf("Testing provider %s (%s)...\n", providerName, providerCfg.Type)

		// Create provider instance
		p, err := provider.GetOrCreate(providerCfg)
		if err != nil {
			fmt.Printf("✗ Failed to create provider: %v\n", err)
			return err
		}

		// Test connection with health check
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		start := time.Now()
		if err := p.HealthCheck(ctx); err != nil {
			fmt.Printf("✗ Connection failed: %v\n", err)
			return err
		}
		elapsed := time.Since(start)

		// Success output
		fmt.Printf("✓ Provider %s: connected\n", providerName)
		fmt.Printf("  Type:     %s\n", p.Type())
		fmt.Printf("  Model:    %s\n", p.Model())
		fmt.Printf("  Latency:  %v\n", elapsed.Round(time.Millisecond))
		if p.SupportsStreaming() {
			fmt.Println("  Streaming: supported")
		}

		return nil
	},
}

func init() {
	providerCmd.AddCommand(providerListCmd)
	providerCmd.AddCommand(providerTestCmd)
}
