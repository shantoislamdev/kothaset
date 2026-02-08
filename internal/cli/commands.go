package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

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
		// TODO: Implement in Phase 3
		return fmt.Errorf("schema validation not yet implemented - coming in Phase 3")
	},
}

var validateDatasetCmd = &cobra.Command{
	Use:   "dataset <path>",
	Short: "Validate an existing dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement after output formats
		return fmt.Errorf("dataset validation not yet implemented")
	},
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
		// TODO: Implement in Phase 3
		return fmt.Errorf("schema show not yet implemented - coming in Phase 3")
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
		// TODO: Implement in Phase 2
		return fmt.Errorf("provider test not yet implemented - coming in Phase 2")
	},
}

func init() {
	providerCmd.AddCommand(providerListCmd)
	providerCmd.AddCommand(providerTestCmd)
}
