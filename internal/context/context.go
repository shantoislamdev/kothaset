// Package context handles context.yaml loading for KothaSet.
package context

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds user-defined context and instructions loaded from context.yaml.
// Both fields are free-form paragraphs - users write whatever they want.
type Config struct {
	// Context describes the purpose of the dataset (free-form paragraph)
	Context string `yaml:"context"`

	// Instruction provides additional generation instructions (free-form paragraph)
	Instruction string `yaml:"instruction"`
}

// searchPaths defines where to look for context.yaml
var searchPaths = []string{
	"context.yaml",
	"context.yml",
}

// Load auto-loads context.yaml from current directory.
// Returns empty Config if no context file found (backward compatible).
func Load() (*Config, error) {
	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return LoadFromFile(path)
		}
	}

	// No context file found - return empty config (backward compatible)
	return &Config{}, nil
}

// LoadFromFile loads context from a specific file path.
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Trim whitespace from both fields
	cfg.Context = strings.TrimSpace(cfg.Context)
	cfg.Instruction = strings.TrimSpace(cfg.Instruction)

	return cfg, nil
}

// IsEmpty returns true if no context or instruction is set.
func (c *Config) IsEmpty() bool {
	return c.Context == "" && c.Instruction == ""
}
