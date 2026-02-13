package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shantoislamdev/kothaset/internal/schema"
	"gopkg.in/yaml.v3"
)

// Config file names
const (
	PublicConfigFile  = "kothaset.yaml"
	SecretsConfigFile = ".secrets.yaml"
)

// LoadPublicConfig loads the public kothaset.yaml configuration
func LoadPublicConfig(configPath string) (*Config, error) {
	// Start with defaults
	cfg := DefaultConfig()

	// Determine which config file to load
	var configFilePath string

	if configPath != "" {
		configFilePath = configPath
	} else {
		// Search for config files
		searchPaths := []string{
			PublicConfigFile,
			"kothaset.yml",
		}

		// Add user config directory
		if home, err := os.UserHomeDir(); err == nil {
			searchPaths = append(searchPaths,
				filepath.Join(home, ".config", "kothaset", "kothaset.yaml"),
				filepath.Join(home, ".config", "kothaset", "kothaset.yml"),
			)
		}

		for _, p := range searchPaths {
			if _, err := os.Stat(p); err == nil {
				configFilePath = p
				break
			}
		}
	}

	// If no config file found, use defaults
	if configFilePath == "" {
		return cfg, nil
	}

	// Read and parse the config file
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Apply defaults for cache_dir if not set
	if cfg.Global.CacheDir == "" {
		cfg.Global.CacheDir = ".kothaset"
	}

	// Trim context whitespace
	cfg.Context = strings.TrimSpace(cfg.Context)

	// Use a runtime default when concurrency is explicitly set to 0.
	if cfg.Global.Concurrency == 0 {
		cfg.Global.Concurrency = runtime.NumCPU()
		if cfg.Global.Concurrency <= 0 {
			cfg.Global.Concurrency = 4
		}
	}

	// Enforce a concrete default output format.
	if cfg.Global.OutputFormat == "" {
		cfg.Global.OutputFormat = "jsonl"
	}

	return cfg, nil
}

// LoadSecretsConfig loads the private .secrets.yaml configuration
func LoadSecretsConfig(secretsPath string) (*SecretsConfig, error) {
	secrets := DefaultSecretsConfig()

	// Determine which secrets file to load
	var secretsFilePath string

	if secretsPath != "" {
		secretsFilePath = secretsPath
	} else {
		// Search for secrets files
		searchPaths := []string{
			SecretsConfigFile,
			".secrets.yml",
		}

		for _, p := range searchPaths {
			if _, err := os.Stat(p); err == nil {
				secretsFilePath = p
				break
			}
		}
	}

	// If no secrets file found, use defaults
	if secretsFilePath == "" {
		return secrets, nil
	}

	// Read and parse the secrets file
	data, err := os.ReadFile(secretsFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets file: %w", err)
	}

	if err := yaml.Unmarshal(data, secrets); err != nil {
		return nil, fmt.Errorf("failed to parse secrets: %w", err)
	}

	for _, p := range secrets.Providers {
		if p.Name == "" {
			return nil, fmt.Errorf("provider name is required")
		}
		if p.Type == "" {
			return nil, fmt.Errorf("provider %s type is required", p.Name)
		}
		if p.Type != "openai" {
			return nil, fmt.Errorf("unsupported provider type: %s", p.Type)
		}
	}

	// Resolve any secret references (env vars)
	if err := resolveSecrets(secrets); err != nil {
		return nil, fmt.Errorf("failed to resolve secrets: %w", err)
	}

	return secrets, nil
}

// GetProvider returns the provider configuration by name
func (s *SecretsConfig) GetProvider(name string) (*ProviderConfig, error) {
	for i := range s.Providers {
		if s.Providers[i].Name == name {
			return &s.Providers[i], nil
		}
	}
	return nil, fmt.Errorf("provider not found: %s", name)
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	if c.Version == "" {
		return fmt.Errorf("config version is required")
	}

	if c.Global.Provider == "" {
		return fmt.Errorf("global.provider is required")
	}

	if c.Global.Schema == "" {
		return fmt.Errorf("global.schema is required")
	}

	// Verify schema is registered.
	if _, err := schema.Get(c.Global.Schema); err != nil {
		return fmt.Errorf("global.schema %q is not a valid schema: %w (available: %v)",
			c.Global.Schema, err, schema.List())
	}

	if c.Global.Model == "" {
		return fmt.Errorf("global.model is required")
	}

	if c.Global.Concurrency < 0 {
		return fmt.Errorf("global.concurrency must be non-negative (0 = use default)")
	}

	if c.Global.OutputFormat == "" {
		return fmt.Errorf("global.output_format is required (supported: jsonl)")
	}
	if c.Global.OutputFormat != "jsonl" {
		return fmt.Errorf("unsupported output_format: %s (supported: jsonl)", c.Global.OutputFormat)
	}

	return nil
}
