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

This command creates:
  - kothaset.yaml  (public config - commit to git)
  - .secrets.yaml  (private config - gitignored)

Example:
  kothaset init
  kothaset init --force  # Overwrite existing files`,
	RunE: runInit,
}

var initForce bool

func init() {
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing files")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Create kothaset.yaml (public config)
	publicPath := "kothaset.yaml"
	if _, err := os.Stat(publicPath); err == nil && !initForce {
		return fmt.Errorf("config file already exists: %s (use --force to overwrite)", publicPath)
	}

	publicContent := `# KothaSet Configuration
# This file is PUBLIC - safe to commit to git
# Available schemas: instruction, chat, preference, classification

version: "1.0"

global:
  provider: openai
  schema: instruction  # Available: instruction, chat, preference, classification
  model: gpt-5.2
  # output_dir: ./output  # Defaults to current directory
  concurrency: 4
  # timeout: 2m  # Default: 2m


# Context: Background info or persona injected into every prompt
context: |
  Generate high-quality training data for an AI assistant.
  The data should be helpful, accurate, and well-formatted.

# Instructions: Specific rules and guidelines for generation
instructions:
  - Be creative and diverse in topics and approaches
  - Vary the style and complexity of responses
  - Use clear and concise language

logging:
  level: info
  format: text
`

	if err := os.WriteFile(publicPath, []byte(publicContent), 0644); err != nil {
		return fmt.Errorf("failed to write kothaset.yaml: %w", err)
	}

	// Create .secrets.yaml (private config)
	secretsPath := ".secrets.yaml"
	secretsContent := `# KothaSet Secrets
# This file is PRIVATE - add to .gitignore!

providers:
  - name: openai
    type: openai
    api_key: env.OPENAI_API_KEY  # Reads from environment variable
    # api_key: sk-...            # Or hardcode key directly
    max_retries: 3
    timeout: 1m
    rate_limit:
      requests_per_minute: 60

  # Example: DeepSeek
  # - name: deepseek
  #   type: openai
  #   base_url: https://api.deepseek.com/v1
  #   api_key: env.DEEPSEEK_API_KEY
  #   max_retries: 3
`

	if _, err := os.Stat(secretsPath); os.IsNotExist(err) || initForce {
		if err := os.WriteFile(secretsPath, []byte(secretsContent), 0644); err != nil {
			return fmt.Errorf("failed to write .secrets.yaml: %w", err)
		}
	}

	// Create .kothaset cache directory
	if err := os.MkdirAll(".kothaset", 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll("./output", 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Update .gitignore if exists
	gitignorePath := ".gitignore"
	gitignoreEntries := "\n# KothaSet\n.secrets.yaml\n.kothaset/\n"
	if data, err := os.ReadFile(gitignorePath); err == nil {
		if !contains(string(data), ".secrets.yaml") {
			f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
			if err == nil {
				f.WriteString(gitignoreEntries)
				f.Close()
				fmt.Println("✓ Updated .gitignore")
			}
		}
	}

	absPath, _ := filepath.Abs(publicPath)
	fmt.Printf("✓ Created %s (public config)\n", absPath)
	fmt.Println("✓ Created .secrets.yaml (private - add your API key)")
	fmt.Println("✓ Created .kothaset/ (cache directory)")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Add your API key to .secrets.yaml or set OPENAI_API_KEY")
	fmt.Println("  2. Edit kothaset.yaml to define your dataset context")
	fmt.Println("  3. Generate: kothaset generate -n 10 -i topics.txt -o dataset.jsonl --seed 42")

	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
