// Package cli implements the command-line interface for KothaSet.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/shantoislamdev/kothaset/internal/config"
)

var (
	// Version information set at build time
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"

	// Global config instances
	cfg     *config.Config
	secrets *config.SecretsConfig

	// Global flags
	cfgFile string
)

// rootCmd represents the base command when called without subcommands
var rootCmd = &cobra.Command{
	Use:   "kothaset",
	Short: "High-quality dataset generation CLI for LLM training",
	Long: `KothaSet is a powerful CLI tool for generating high-quality datasets
using large language models as teacher models. These datasets are designed
for training and fine-tuning smaller models (0.6B-32B parameters).

Features:
  • Multiple LLM providers (OpenAI and OpenAI-compatible endpoints)
  • Flexible dataset schemas (instruction, chat, preference, classification)
  • Resumable generation with checkpointing
  • Seed-based generation for diversity
  • JSONL output format with streaming writes

Example:
  kothaset generate --schema instruction --count 1000 --output dataset.jsonl`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config loading for version and init commands
		if cmd.Name() == "version" || cmd.Name() == "init" {
			return nil
		}
		return initConfig()
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: kothaset.yaml)")

	// Register subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(schemaCmd)
	rootCmd.AddCommand(providerCmd)
}

// initConfig reads in config file and ENV variables if set
func initConfig() error {
	var err error

	// Load public config (kothaset.yaml)
	cfg, err = config.LoadPublicConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load kothaset.yaml: %w", err)
	}

	// Load secrets (.secrets.yaml)
	secrets, err = config.LoadSecretsConfig("")
	if err != nil {
		return fmt.Errorf("failed to load .secrets.yaml: %w", err)
	}

	return nil
}
