package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/shantoislamdev/kothaset/internal/config"
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

	// Create default configuration
	defaultConfig := config.DefaultConfig()

	// Marshal to YAML
	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add header comment
	header := `# KothaSet Configuration
# Documentation: https://github.com/shantoislamdev/kothaset

`
	content := header + string(data)

	// Write config file
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Create output directory if specified
	if defaultConfig.Global.OutputDir != "" {
		if err := os.MkdirAll(defaultConfig.Global.OutputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	absPath, _ := filepath.Abs(configPath)
	fmt.Printf("✓ Created configuration file: %s\n", absPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Add your API key to the config or set OPENAI_API_KEY environment variable")
	fmt.Println("  2. Run 'kothaset generate --help' to see generation options")
	fmt.Println("  3. Generate your first dataset: kothaset generate -n 10 -o dataset.jsonl")

	return nil
}
