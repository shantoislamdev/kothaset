// Package config handles configuration loading and management for KothaSet.
package config

import (
	"time"
)

// Config is the root configuration structure for kothaset.yaml (public config)
type Config struct {
	// Version of the config schema for migrations
	Version string `yaml:"version" json:"version"`

	// Global settings
	Global GlobalConfig `yaml:"global" json:"global"`

	// Context is the free-form context paragraph for dataset generation
	Context string `yaml:"context,omitempty" json:"context,omitempty"`

	// Instructions is a list of generation instructions (one per line)
	Instructions []string `yaml:"instructions,omitempty" json:"instructions,omitempty"`

	// Logging configuration
	Logging LoggingConfig `yaml:"logging,omitempty" json:"logging,omitempty"`

	// Named profiles for quick switching (optional)
	Profiles map[string]Profile `yaml:"profiles,omitempty" json:"profiles,omitempty"`
}

// GlobalConfig contains global settings
type GlobalConfig struct {
	// Provider is the LLM provider to use
	Provider string `yaml:"provider" json:"provider"`

	// Schema is the dataset schema (instruction, chat, preference, classification)
	Schema string `yaml:"schema" json:"schema"`

	// Model is the LLM model to use
	Model string `yaml:"model" json:"model"`

	// OutputDir is the default directory for generated datasets
	OutputDir string `yaml:"output_dir" json:"output_dir"`

	// CacheDir is the directory for caching (optional, defaults to .kothaset/)
	CacheDir string `yaml:"cache_dir,omitempty" json:"cache_dir,omitempty"`

	// Concurrency is the default number of concurrent workers
	Concurrency int `yaml:"concurrency" json:"concurrency"`

	// Timeout is the default request timeout
	Timeout Duration `yaml:"timeout" json:"timeout"`

	// MaxTokens is the default max tokens per response (0 = unlimited/model default)
	MaxTokens int `yaml:"max_tokens,omitempty" json:"max_tokens,omitempty"`

	// OutputFormat is the default output format (jsonl, parquet, hf)
	OutputFormat string `yaml:"output_format,omitempty" json:"output_format,omitempty"`
}

// SecretsConfig is the root structure for .secrets.yaml (private config)
type SecretsConfig struct {
	// Providers contains provider configurations with credentials
	Providers []ProviderConfig `yaml:"providers" json:"providers"`
}

// ProviderConfig contains LLM provider settings (in .secrets.yaml)
type ProviderConfig struct {
	// Name is the unique identifier for this provider configuration
	Name string `yaml:"name" json:"name"`

	// Type is the provider type (openai, anthropic, custom)
	Type string `yaml:"type" json:"type"`

	// BaseURL is the API base URL (for custom endpoints)
	BaseURL string `yaml:"base_url,omitempty" json:"base_url,omitempty"`

	// APIKey is the API key (can be a secret reference)
	APIKey string `yaml:"api_key,omitempty" json:"api_key,omitempty"`

	// Headers are additional HTTP headers
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`

	// Timeout for requests to this provider
	Timeout Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`

	// MaxRetries is the maximum number of retries on failure
	MaxRetries int `yaml:"max_retries" json:"max_retries"`

	// RateLimit configuration
	RateLimit RateLimitConfig `yaml:"rate_limit,omitempty" json:"rate_limit,omitempty"`
}

// RateLimitConfig defines rate limiting settings
type RateLimitConfig struct {
	// RequestsPerMinute is the maximum requests per minute
	RequestsPerMinute int `yaml:"requests_per_minute" json:"requests_per_minute"`

	// TokensPerMinute is the maximum tokens per minute
	TokensPerMinute int `yaml:"tokens_per_minute,omitempty" json:"tokens_per_minute,omitempty"`
}

// Profile is a named preset for quick configuration
type Profile struct {
	// Description of what this profile is for
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Provider to use
	Provider string `yaml:"provider,omitempty" json:"provider,omitempty"`

	// Schema to use
	Schema string `yaml:"schema,omitempty" json:"schema,omitempty"`

	// Generation settings
	Generation GenerationConfig `yaml:"generation,omitempty" json:"generation,omitempty"`
}

// GenerationConfig contains settings for dataset generation
type GenerationConfig struct {
	// Temperature for sampling
	Temperature float64 `yaml:"temperature" json:"temperature"`

	// MaxTokens per response
	MaxTokens int `yaml:"max_tokens" json:"max_tokens"`

	// TopP nucleus sampling parameter
	TopP float64 `yaml:"top_p,omitempty" json:"top_p,omitempty"`

	// Seed for reproducibility
	Seed int64 `yaml:"seed" json:"seed"`

	// Workers for concurrent generation
	Workers int `yaml:"workers" json:"workers"`

	// BatchSize for batch requests (if supported)
	BatchSize int `yaml:"batch_size,omitempty" json:"batch_size,omitempty"`

	// CheckpointEvery samples to save progress
	CheckpointEvery int `yaml:"checkpoint_every" json:"checkpoint_every"`

	// SystemPrompt override
	SystemPrompt string `yaml:"system_prompt,omitempty" json:"system_prompt,omitempty"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	// Level is the log level (debug, info, warn, error)
	Level string `yaml:"level" json:"level"`

	// Format is the log format (text, json)
	Format string `yaml:"format" json:"format"`

	// File is an optional log file path
	File string `yaml:"file,omitempty" json:"file,omitempty"`
}

// Duration is a wrapper around time.Duration for YAML unmarshaling
type Duration struct {
	time.Duration
}

// UnmarshalYAML implements yaml.Unmarshaler
func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = duration
	return nil
}

// MarshalYAML implements yaml.Marshaler
func (d Duration) MarshalYAML() (interface{}, error) {
	return d.Duration.String(), nil
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Version: "1.0",
		Global: GlobalConfig{
			Provider:     "openai",
			Schema:       "instruction",
			Model:        "gpt-5.2",
			OutputDir:    ".",
			CacheDir:     ".kothaset",
			Concurrency:  4,
			Timeout:      Duration{time.Minute * 2},
			OutputFormat: "jsonl",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

// DefaultSecretsConfig returns default secrets configuration
func DefaultSecretsConfig() *SecretsConfig {
	return &SecretsConfig{
		Providers: []ProviderConfig{
			{
				Name:       "openai",
				Type:       "openai",
				APIKey:     "env.OPENAI_API_KEY",
				MaxRetries: 3,
				Timeout:    Duration{time.Minute},
				RateLimit: RateLimitConfig{
					RequestsPerMinute: 60,
				},
			},
		},
	}
}
