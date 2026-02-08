package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config file names
const (
	PublicConfigFile  = "kothaset.yaml"
	SecretsConfigFile = ".secrets.yaml"
)

// Load loads configuration from kothaset.yaml and .secrets.yaml
func Load() (*Config, *SecretsConfig, error) {
	cfg, err := LoadPublicConfig("")
	if err != nil {
		return nil, nil, err
	}

	secrets, err := LoadSecretsConfig("")
	if err != nil {
		return nil, nil, err
	}

	return cfg, secrets, nil
}

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

	// Resolve any secret references (env vars)
	if err := resolveProviderSecrets(secrets); err != nil {
		return nil, fmt.Errorf("failed to resolve secrets: %w", err)
	}

	return secrets, nil
}

// resolveProviderSecrets resolves environment variable references in provider configs
func resolveProviderSecrets(secrets *SecretsConfig) error {
	for i := range secrets.Providers {
		p := &secrets.Providers[i]

		// Resolve API key from env if specified
		if p.APIKeyEnv != "" && p.APIKey == "" {
			p.APIKey = os.Getenv(p.APIKeyEnv)
		}

		// Handle env.VAR_NAME syntax in api_key field
		if strings.HasPrefix(p.APIKey, "env.") {
			envVar := strings.TrimPrefix(p.APIKey, "env.")
			p.APIKey = os.Getenv(envVar)
		}
	}
	return nil
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

	return nil
}

// SavePublicConfig saves the public configuration to kothaset.yaml
func SavePublicConfig(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
