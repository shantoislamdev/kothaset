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

Notes:
- Checkpoints are saved under `.kothaset/` with filenames derived from the absolute output path.
- Retryable provider errors use exponential backoff with jitter and respect provider retry-after hints when available.

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

1. **`api_key: env.VAR_NAME`**: If `api_key` starts with `env.`, read from that environment variable (e.g., `env.OPENAI_API_KEY`).
2. **Legacy env reference `${env:VAR_NAME}`**: Backward-compatible env reference format.
3. **File reference `${file:/path/to/secret}`**: Reads the secret value from a file.
4. **Raw `api_key` value**: Otherwise, the value is used directly (e.g., `sk-...`).
5. **Default provider env fallback** (when `api_key` is empty):
   - `openai` → `OPENAI_API_KEY`
   - `anthropic` → `ANTHROPIC_API_KEY`
   - `deepseek` → `DEEPSEEK_API_KEY`

If a provider key cannot be resolved during load, KothaSet logs a warning to stderr and continues loading. Validation still happens when the provider is used.

**Recommendation:** Use `env.VAR_NAME` for security.

---

## Environment Variables

You can also use environment variables for API keys, which is recommended for CI/CD environments.

- `OPENAI_API_KEY`
- `ANTHROPIC_API_KEY`
- `DEEPSEEK_API_KEY`

In `.secrets.yaml`, reference them using the `env.` prefix:

```yaml
api_key: env.OPENAI_API_KEY
```
