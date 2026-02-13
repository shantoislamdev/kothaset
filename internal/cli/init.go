package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

version: "1.0"

global:
  provider: openai
  schema: instruction  # Available: instruction, chat, preference, classification
  model: gpt-5.2
  # output_dir: ./output  # Defaults to current directory
  concurrency: 4


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
		if err := os.WriteFile(secretsPath, []byte(secretsContent), 0600); err != nil {
			return fmt.Errorf("failed to write .secrets.yaml: %w", err)
		}
	}

	// Create .kothaset cache directory
	if err := os.MkdirAll(".kothaset", 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Handle .gitignore
	if err := handleGitignore(); err != nil {
		// Non-fatal error, just log it
		fmt.Printf("Warning: could not update .gitignore: %v\n", err)
	}

	absPath, _ := filepath.Abs(publicPath)
	fmt.Printf("✓ Created %s (public config)\n", absPath)
	fmt.Println("✓ Created .secrets.yaml (private - add your API key)")
	fmt.Println("✓ Created .kothaset/ (cache directory)")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Add your API key to .secrets.yaml or set OPENAI_API_KEY")
	fmt.Println("  2. Edit kothaset.yaml to define your dataset context")
	fmt.Println("  3. Generate: kothaset generate -n 10 -i topics.txt -o output/dataset.jsonl")
	fmt.Println("     (or use --seed random for different random seeds per request)")

	return nil
}

// handleGitignore checks for existing .gitignore and manages KothaSet entries
func handleGitignore() error {
	gitignorePath := ".gitignore"
	entries := []string{".secrets.yaml", ".kothaset/"}

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		// File doesn't exist, create new .gitignore
		if os.IsNotExist(err) {
			content := "# KothaSet\n"
			for _, entry := range entries {
				content += entry + "\n"
			}
			if err := os.WriteFile(gitignorePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to create .gitignore: %w", err)
			}
			fmt.Println("✓ Created .gitignore")
			return nil
		}
		return fmt.Errorf("failed to read .gitignore: %w", err)
	}

	// File exists, check for missing entries
	content := string(data)
	missingEntries := []string{}

	for _, entry := range entries {
		if !gitignoreContains(content, entry) {
			missingEntries = append(missingEntries, entry)
		}
	}

	// All entries already present, do nothing
	if len(missingEntries) == 0 {
		return nil
	}

	// Append missing entries
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open .gitignore for appending: %w", err)
	}
	defer f.Close()

	// Add newline if file doesn't end with one
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		f.WriteString("\n")
	}

	// Check if we need to add a header
	if !strings.Contains(content, "# KothaSet") {
		f.WriteString("\n# KothaSet\n")
	}

	for _, entry := range missingEntries {
		f.WriteString(entry + "\n")
	}

	fmt.Println("✓ Updated .gitignore")
	return nil
}

// gitignoreChecks if content contains the given entry, handling variations
func gitignoreContains(content, entry string) bool {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Check for exact match or pattern match
		if trimmed == entry {
			return true
		}
		// Handle variations: .kothaset/ vs .kothaset vs .kothaset/*
		entryTrimmed := strings.TrimSuffix(entry, "/")
		trimmedNoSuffix := strings.TrimSuffix(trimmed, "/")
		if trimmedNoSuffix == entryTrimmed {
			return true
		}
	}
	return false
}
