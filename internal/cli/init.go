package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new KothaSet project",
	Long: `Initialize a new KothaSet project in the current directory.

This command creates a .kothaset.yaml configuration file with sensible
defaults and example provider/schema configurations.

Example:
  kothaset init
  kothaset init --force  # Overwrite existing config`,
	RunE: runInit,
}

var initForce bool

func init() {
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing config file")
}

func runInit(cmd *cobra.Command, args []string) error {
	configPath := ".kothaset.yaml"

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil && !initForce {
		return fmt.Errorf("config file already exists: %s (use --force to overwrite)", configPath)
	}

	// Config template with all fields visible
	content := `# KothaSet Configuration
# Documentation: https://github.com/shantoislamdev/kothaset

version: "1.0"

global:
  default_provider: openai
  default_schema: instruction
  output_dir: ./output
  cache_dir: ./.kothaset
  concurrency: 4
  timeout: 2m0s

providers:
  - name: openai
    type: openai
    base_url: https://api.openai.com/v1  # Change for OpenAI-compatible APIs
    api_key: env.OPENAI_API_KEY  # Use env.VAR_NAME for environment variable or raw API key
    model: gpt-4
    max_retries: 3
    timeout: 1m0s
    rate_limit:
      requests_per_minute: 60

schemas:
  - name: instruction
    builtin: true
  - name: chat
    builtin: true
  - name: preference
    builtin: true
  - name: classification
    builtin: true

logging:
  level: info
  format: text
`

	// Write config file
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Create example context.yaml
	contextPath := "context.yaml"
	contextContent := `# KothaSet Context Configuration
# This file defines the purpose and instructions for your dataset.
# Both fields are free-form paragraphs - write whatever you need.

context: |
  Generate high-quality training data for an AI assistant.
  The data should be helpful, accurate, and well-formatted.

instruction: |
  Be creative and diverse in topics and approaches.
  Vary the style and complexity of responses.
`
	if _, err := os.Stat(contextPath); os.IsNotExist(err) {
		if err := os.WriteFile(contextPath, []byte(contextContent), 0644); err != nil {
			return fmt.Errorf("failed to write context.yaml: %w", err)
		}
	}

	// Create output directory
	if err := os.MkdirAll("./output", 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	absPath, _ := filepath.Abs(configPath)
	fmt.Printf("✓ Created configuration file: %s\n", absPath)
	fmt.Println("✓ Created context.yaml (customize for your dataset)")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Add your API key to .kothaset.yaml or set OPENAI_API_KEY")
	fmt.Println("  2. Edit context.yaml to define your dataset purpose and instructions")
	fmt.Println("  3. Generate: kothaset generate -n 10 -i topics.txt -o dataset.jsonl --seed 42")

	return nil
}
