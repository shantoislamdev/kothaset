// Package config handles configuration loading and management for KothaSet.
package config

import (
	"time"
)

// Config is the root configuration structure
type Config struct {
	// Version of the config schema for migrations
	Version string `yaml:"version" json:"version"`

	// Global settings
	Global GlobalConfig `yaml:"global" json:"global"`

	// Provider configurations
	Providers []ProviderConfig `yaml:"providers" json:"providers"`

	// Schema configurations
	Schemas []SchemaConfig `yaml:"schemas" json:"schemas"`

	// Named profiles for quick switching
	Profiles map[string]Profile `yaml:"profiles,omitempty" json:"profiles,omitempty"`

	// Logging configuration
	Logging LoggingConfig `yaml:"logging" json:"logging"`
}

// GlobalConfig contains global settings
type GlobalConfig struct {
	// DefaultProvider is the default LLM provider to use
	DefaultProvider string `yaml:"default_provider" json:"default_provider"`

	// DefaultSchema is the default dataset schema
	DefaultSchema string `yaml:"default_schema" json:"default_schema"`

	// OutputDir is the default directory for generated datasets
	OutputDir string `yaml:"output_dir" json:"output_dir"`

	// CacheDir is the directory for caching (e.g., checkpoints)
	CacheDir string `yaml:"cache_dir" json:"cache_dir"`

	// Concurrency is the default number of concurrent workers
	Concurrency int `yaml:"concurrency" json:"concurrency"`

	// Timeout is the default request timeout
	Timeout Duration `yaml:"timeout" json:"timeout"`
}

// ProviderConfig contains LLM provider settings
type ProviderConfig struct {
	// Name is the unique identifier for this provider configuration
	Name string `yaml:"name" json:"name"`

	// Type is the provider type (openai, anthropic, custom)
	Type string `yaml:"type" json:"type"`

	// BaseURL is the API base URL (for custom endpoints)
	BaseURL string `yaml:"base_url,omitempty" json:"base_url,omitempty"`

	// APIKey is the API key (can be a secret reference)
	APIKey string `yaml:"api_key,omitempty" json:"api_key,omitempty"`

	// APIKeyEnv is the environment variable containing the API key
	APIKeyEnv string `yaml:"api_key_env,omitempty" json:"api_key_env,omitempty"`

	// Model is the default model to use
	Model string `yaml:"model" json:"model"`

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

// SchemaConfig contains dataset schema settings
type SchemaConfig struct {
	// Name is the unique identifier for this schema
	Name string `yaml:"name" json:"name"`

	// Path is the file path to a custom schema definition
	Path string `yaml:"path,omitempty" json:"path,omitempty"`

	// Builtin indicates this is a built-in schema name
	Builtin bool `yaml:"builtin,omitempty" json:"builtin,omitempty"`
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

	// Seed for reproducibility (0 = random)
	Seed int64 `yaml:"seed,omitempty" json:"seed,omitempty"`

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
			DefaultProvider: "openai",
			DefaultSchema:   "instruction",
			OutputDir:       "./output",
			CacheDir:        "./.kothaset",
			Concurrency:     4,
			Timeout:         Duration{time.Minute * 2},
		},
		Providers: []ProviderConfig{
			{
				Name:       "openai",
				Type:       "openai",
				BaseURL:    "", // Set custom base URL for OpenAI-compatible APIs
				APIKeyEnv:  "OPENAI_API_KEY",
				Model:      "gpt-4",
				MaxRetries: 3,
				Timeout:    Duration{time.Minute},
				RateLimit: RateLimitConfig{
					RequestsPerMinute: 60,
				},
			},
		},
		Schemas: []SchemaConfig{
			{Name: "instruction", Builtin: true},
			{Name: "chat", Builtin: true},
			{Name: "preference", Builtin: true},
			{Name: "classification", Builtin: true},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}
}
