// Package cli implements the command-line interface for KothaSet.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shantoislamdev/kothaset/internal/config"
)

var (
	// Version information set at build time
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"

	// Global config instance
	cfg *config.Config

	// Global flags
	cfgFile string
	verbose bool
	quiet   bool
)

// rootCmd represents the base command when called without subcommands
var rootCmd = &cobra.Command{
	Use:   "kothaset",
	Short: "High-quality dataset generation CLI for LLM training",
	Long: `KothaSet is a powerful CLI tool for generating high-quality datasets
using large language models as teacher models. These datasets are designed
for training and fine-tuning smaller models (0.6B-32B parameters).

Features:
  • Multiple LLM providers (OpenAI, Anthropic, custom endpoints)
  • Flexible dataset schemas (instruction, chat, preference, classification)
  • Resumable generation with checkpointing
  • Seed-based generation for diversity
  • Multiple output formats (JSONL, Parquet, HuggingFace)

Example:
  kothaset generate --schema instruction --count 1000 --output dataset.jsonl`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config loading for version command
		if cmd.Name() == "version" {
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
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: .kothaset.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-essential output")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))

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
	cfg, err = config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	return nil
}

// GetConfig returns the loaded configuration
func GetConfig() *config.Config {
	return cfg
}

// IsVerbose returns whether verbose mode is enabled
func IsVerbose() bool {
	return verbose
}

// IsQuiet returns whether quiet mode is enabled
func IsQuiet() bool {
	return quiet
}

// printError prints an error message to stderr
func printError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}
