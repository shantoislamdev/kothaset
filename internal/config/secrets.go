package config

import (
	"fmt"
	"os"
	"strings"
)

// resolveSecrets resolves all secret references in the secrets configuration
func resolveSecrets(cfg *SecretsConfig) error {
	for i := range cfg.Providers {
		p := &cfg.Providers[i]

		// Resolve API key
		apiKey, err := resolveAPIKey(p)
		if err != nil {
			// Don't fail on missing API keys during config loading
			// They will be validated when the provider is used
			fmt.Fprintf(os.Stderr, "⚠ Provider '%s': %v (will validate later)\n", p.Name, err)
			continue
		}
		p.APIKey = apiKey
	}
	return nil
}

// resolveAPIKey resolves the API key for a provider
// Supports formats:
//   - Raw API key: "sk-..." (used as-is)
//   - Environment variable: "env.OPENAI_API_KEY" (reads from env)
//   - Legacy secret ref: "${env:MY_KEY}" (backwards compatible)
func resolveAPIKey(p *ProviderConfig) (string, error) {
	apiKey := p.APIKey

	// Check for env.VAR_NAME format
	if strings.HasPrefix(apiKey, "env.") {
		envVar := strings.TrimPrefix(apiKey, "env.")
		if value := os.Getenv(envVar); value != "" {
			return value, nil
		}
		return "", fmt.Errorf("environment variable not set: %s", envVar)
	}

	// Check for legacy ${env:VAR_NAME} format
	if isSecretRef(apiKey) {
		return resolveSecretRef(apiKey)
	}

	// Raw API key (used as-is)
	if apiKey != "" {
		return apiKey, nil
	}

	// Fallback: Default environment variable based on provider type
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

	secretType := parts[0]
	value := parts[1]

	switch secretType {
	case "env":
		envValue := os.Getenv(value)
		if envValue == "" {
			return "", fmt.Errorf("environment variable not set: %s", value)
		}
		return envValue, nil

	case "file":
		data, err := os.ReadFile(value)
		if err != nil {
			return "", fmt.Errorf("failed to read secret file: %w", err)
		}
		return strings.TrimSpace(string(data)), nil

	case "plain":
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
