package config

import (
	"fmt"
	"os"
	"strings"
)

// SecretType defines how a secret is stored/retrieved
type SecretType string

const (
	// SecretTypeEnv retrieves secret from environment variable
	SecretTypeEnv SecretType = "env"
	// SecretTypeFile retrieves secret from a file
	SecretTypeFile SecretType = "file"
	// SecretTypePlain is a plain text secret (not recommended)
	SecretTypePlain SecretType = "plain"
)

// SecretRef represents a reference to a secret value
type SecretRef struct {
	Type  SecretType `yaml:"$type,omitempty" json:"$type,omitempty"`
	Value string     `yaml:"$value,omitempty" json:"$value,omitempty"`
}

// resolveSecrets resolves all secret references in the configuration
func resolveSecrets(cfg *Config) error {
	for i := range cfg.Providers {
		p := &cfg.Providers[i]

		// Resolve API key
		apiKey, err := resolveAPIKey(p)
		if err != nil {
			// Don't fail on missing API keys during config loading
			// They will be validated when the provider is used
			continue
		}
		p.APIKey = apiKey
	}
	return nil
}

// resolveAPIKey resolves the API key for a provider
func resolveAPIKey(p *ProviderConfig) (string, error) {
	// Priority 1: Direct API key value
	if p.APIKey != "" && !isSecretRef(p.APIKey) {
		return p.APIKey, nil
	}

	// Priority 2: Environment variable reference
	if p.APIKeyEnv != "" {
		if value := os.Getenv(p.APIKeyEnv); value != "" {
			return value, nil
		}
		// Try common naming conventions
		alternatives := []string{
			strings.ToUpper(p.APIKeyEnv),
			strings.ToUpper(p.Name) + "_API_KEY",
		}
		for _, alt := range alternatives {
			if value := os.Getenv(alt); value != "" {
				return value, nil
			}
		}
	}

	// Priority 3: Parse secret reference in APIKey field
	if p.APIKey != "" && isSecretRef(p.APIKey) {
		return resolveSecretRef(p.APIKey)
	}

	// Priority 4: Default environment variable based on provider type
	defaultEnvVars := map[string]string{
		"openai":    "OPENAI_API_KEY",
		"anthropic": "ANTHROPIC_API_KEY",
		"deepseek":  "DEEPSEEK_API_KEY",
	}
	if envVar, ok := defaultEnvVars[p.Type]; ok {
		if value := os.Getenv(envVar); value != "" {
			return value, nil
		}
	}

	return "", fmt.Errorf("no API key found for provider %s", p.Name)
}

// isSecretRef checks if a string looks like a secret reference
func isSecretRef(s string) bool {
	return strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}")
}

// resolveSecretRef resolves a secret reference string
// Format: ${type:value}
// Examples:
//   - ${env:MY_API_KEY}
//   - ${file:/path/to/secret}
func resolveSecretRef(ref string) (string, error) {
	// Remove ${ and }
	inner := ref[2 : len(ref)-1]

	parts := strings.SplitN(inner, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid secret reference format: %s", ref)
	}

	secretType := SecretType(parts[0])
	value := parts[1]

	switch secretType {
	case SecretTypeEnv:
		envValue := os.Getenv(value)
		if envValue == "" {
			return "", fmt.Errorf("environment variable not set: %s", value)
		}
		return envValue, nil

	case SecretTypeFile:
		data, err := os.ReadFile(value)
		if err != nil {
			return "", fmt.Errorf("failed to read secret file: %w", err)
		}
		return strings.TrimSpace(string(data)), nil

	case SecretTypePlain:
		return value, nil

	default:
		return "", fmt.Errorf("unknown secret type: %s", secretType)
	}
}

// ResolveSecret is a public helper to resolve a single secret reference
func ResolveSecret(ref string) (string, error) {
	if !isSecretRef(ref) {
		return ref, nil
	}
	return resolveSecretRef(ref)
}

// MaskSecret returns a masked version of a secret for display
func MaskSecret(secret string) string {
	if len(secret) <= 8 {
		return "********"
	}
	return secret[:4] + "..." + secret[len(secret)-4:]
}
