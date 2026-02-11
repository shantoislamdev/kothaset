# Configuration

KothaSet uses a **two-file configuration system** to separate public settings from private credentials.

## Files

| File | Visibility | Purpose |
|------|------------|---------|
| **`kothaset.yaml`** | PUBLIC | Shared settings, context, instructions. **Commit to git.** |
| **`.secrets.yaml`** | PRIVATE | Provider credentials and API keys. **Add to `.gitignore`.** |

---

## 1. `kothaset.yaml` (Public)

This file controls *what* you generate.

```yaml
version: "1.0"

global:
  provider: openai      # Default provider to use
  schema: instruction   # Default schema (instruction, chat, preference, classification)
  model: gpt-5.2        # Model to use (moved from provider config)
  concurrency: 4        # Number of concurrent workers
  output_dir: ./output  # Default output directory (optional)
  cache_dir: .kothaset  # Cache directory (optional)
  timeout: 2m         # Request timeout (optional)
  max_tokens: 2048    # Max tokens per response (optional)
  output_format: jsonl # Default output format (optional)
  checkpoint_every: 10  # Save checkpoint every N samples (default: 10, 0 to disable)

# Context: Background info or persona injected into every prompt
context: |
  Generate high-quality training data for an AI assistant.
  The data should be helpful, accurate, and well-formatted.

# Instructions: Specific rules and guidelines for generation
instructions:
  - Be creative and diverse in topics and approaches
  - Vary the style and complexity of responses
  - Use clear and concise language

logging:
  level: info           # debug, info, warn, error
  format: text          # text, json
  file: kothaset.log    # Optional log file path



```

---

## 2. `.secrets.yaml` (Private)

This file controls *how* you access LLMs.

```yaml
providers:
  - name: openai
    type: openai
    api_key: env.OPENAI_API_KEY  # Reads from environment variable
    # api_key: sk-...            # Or hardcode key directly
    timeout: 1m
    rate_limit:
      requests_per_minute: 60

  # Custom endpoint example (DeepSeek, vLLM, Ollama)
  - name: local
    type: openai
    base_url: http://localhost:8000/v1
    api_key: not-needed          # Use 'api_key' for non-sensitive values
```

### API Key Resolution Logic

KothaSet resolves API keys in the following order:

1.  **`api_key: env.VAR_NAME`**: If `api_key` starts with `env.`, the value is read from the specified environment variable (e.g., `env.OPENAI_API_KEY`).
2.  **`api_key` (Raw Value)**: Otherwise, the string is used directly (e.g., `sk-...`).

**Recommendation:** Use `env.VAR_NAME` for security.

---

## Environment Variables

You can also use environment variables for API keys, which is recommended for CI/CD environments.

- `OPENAI_API_KEY`
- `ANTHROPIC_API_KEY`
- `DEEPSEEK_API_KEY`

77- In `.secrets.yaml`, reference them using the `env.` prefix:
78- 
79- ```yaml
80- api_key: env.OPENAI_API_KEY
81- ```
