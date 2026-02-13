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
	if strings.HasPrefix(apiKey, "${") && strings.HasSuffix(apiKey, "}") {
		return resolveLegacyEnvRef(apiKey)
	}

	// Raw API key (used as-is)
	if apiKey != "" {
		return apiKey, nil
	}

	// Fallback: Default environment variable based on provider type
	defaultEnvVars := map[string]string{
		"openai": "OPENAI_API_KEY",
	}
	if envVar, ok := defaultEnvVars[p.Type]; ok {
		if value := os.Getenv(envVar); value != "" {
			return value, nil
		}
	}

	return "", fmt.Errorf("no API key found for provider %s", p.Name)
}

// resolveLegacyEnvRef resolves legacy ${env:VAR_NAME} references.
func resolveLegacyEnvRef(ref string) (string, error) {
	// Remove ${ and }
	inner := ref[2 : len(ref)-1]

	parts := strings.SplitN(inner, ":", 2)
	if len(parts) != 2 || parts[0] != "env" {
		return "", fmt.Errorf("unsupported secret reference: %s (only env vars are supported)", ref)
	}

	envVar := parts[1]
	envValue := os.Getenv(envVar)
	if envValue == "" {
		return "", fmt.Errorf("environment variable not set: %s", envVar)
	}
	return envValue, nil
}
