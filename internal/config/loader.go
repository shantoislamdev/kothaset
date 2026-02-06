package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load loads configuration from multiple sources with proper precedence
func Load(configPath string) (*Config, error) {
	// Start with defaults
	cfg := DefaultConfig()

	// Determine which config file to load
	var configFilePath string

	if configPath != "" {
		// Explicit path provided
		configFilePath = configPath
	} else {
		// Search for config files
		searchPaths := []string{
			".kothaset.yaml",
			".kothaset.yml",
		}

		// Add user config directory
		if home, err := os.UserHomeDir(); err == nil {
			searchPaths = append(searchPaths,
				filepath.Join(home, ".config", "kothaset", "config.yaml"),
				filepath.Join(home, ".config", "kothaset", "config.yml"),
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
		if err := resolveSecrets(cfg); err != nil {
			return nil, fmt.Errorf("failed to resolve secrets: %w", err)
		}
		return cfg, nil
	}

	// Read and parse the config file using yaml
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Resolve any secret references
	if err := resolveSecrets(cfg); err != nil {
		return nil, fmt.Errorf("failed to resolve secrets: %w", err)
	}

	return cfg, nil
}

// LoadFromFile loads configuration from a specific file path
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := resolveSecrets(cfg); err != nil {
		return nil, fmt.Errorf("failed to resolve secrets: %w", err)
	}

	return cfg, nil
}

// SaveToFile saves configuration to a file
func SaveToFile(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GetProvider returns the provider configuration by name
func (c *Config) GetProvider(name string) (*ProviderConfig, error) {
	for i := range c.Providers {
		if c.Providers[i].Name == name {
			return &c.Providers[i], nil
		}
	}
	return nil, fmt.Errorf("provider not found: %s", name)
}

// GetDefaultProvider returns the default provider configuration
func (c *Config) GetDefaultProvider() (*ProviderConfig, error) {
	if c.Global.DefaultProvider == "" {
		if len(c.Providers) > 0 {
			return &c.Providers[0], nil
		}
		return nil, fmt.Errorf("no providers configured")
	}
	return c.GetProvider(c.Global.DefaultProvider)
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	if c.Version == "" {
		return fmt.Errorf("config version is required")
	}

	// Validate providers
	for i, p := range c.Providers {
		if p.Name == "" {
			return fmt.Errorf("provider %d: name is required", i)
		}
		if p.Type == "" {
			return fmt.Errorf("provider %s: type is required", p.Name)
		}
		if p.Model == "" {
			return fmt.Errorf("provider %s: model is required", p.Name)
		}
	}

	// Validate default provider exists
	if c.Global.DefaultProvider != "" {
		if _, err := c.GetProvider(c.Global.DefaultProvider); err != nil {
			return fmt.Errorf("default provider not found: %s", c.Global.DefaultProvider)
		}
	}

	return nil
}

// Merge merges another config into this one (other takes precedence)
func (c *Config) Merge(other *Config) {
	if other.Version != "" {
		c.Version = other.Version
	}

	// Merge global settings
	if other.Global.DefaultProvider != "" {
		c.Global.DefaultProvider = other.Global.DefaultProvider
	}
	if other.Global.DefaultSchema != "" {
		c.Global.DefaultSchema = other.Global.DefaultSchema
	}
	if other.Global.OutputDir != "" {
		c.Global.OutputDir = other.Global.OutputDir
	}
	if other.Global.Concurrency > 0 {
		c.Global.Concurrency = other.Global.Concurrency
	}

	// Merge providers (add or replace by name)
	for _, op := range other.Providers {
		found := false
		for i := range c.Providers {
			if c.Providers[i].Name == op.Name {
				c.Providers[i] = op
				found = true
				break
			}
		}
		if !found {
			c.Providers = append(c.Providers, op)
		}
	}

	// Merge profiles
	if c.Profiles == nil {
		c.Profiles = make(map[string]Profile)
	}
	for k, v := range other.Profiles {
		c.Profiles[k] = v
	}
}
