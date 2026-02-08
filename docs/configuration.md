# Configuration Reference

KothaSet uses a YAML configuration file (`.kothaset.yaml`) to manage providers, schemas, and generation settings.

## Quick Start

Initialize a configuration file:

```bash
kothaset init
```

This creates `.kothaset.yaml` in your current directory.

---

## Configuration Structure

```yaml
version: "1.0"

global:
  default_provider: openai
  default_schema: instruction
  output_dir: ./output
  cache_dir: ./.kothaset
  concurrency: 4
  timeout: 2m

providers:
  - name: openai
    type: openai
    api_key_env: OPENAI_API_KEY
    model: gpt-4
    max_retries: 3
    timeout: 1m
    rate_limit:
      requests_per_minute: 60

schemas:
  - name: instruction
    builtin: true

profiles:
  fast:
    provider: openai
    schema: instruction
    generation:
      temperature: 0.7
      max_tokens: 1024
      seed: 42
      workers: 8

logging:
  level: info
  format: text
```

---

## Global Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `default_provider` | string | `openai` | Default LLM provider to use |
| `default_schema` | string | `instruction` | Default dataset schema |
| `output_dir` | string | `./output` | Directory for generated datasets |
| `cache_dir` | string | `./.kothaset` | Directory for checkpoints and cache |
| `concurrency` | int | `4` | Number of concurrent workers |
| `timeout` | duration | `2m` | Default request timeout |

---

## Provider Configuration

Each provider entry supports:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | ✓ | Unique identifier for this provider |
| `type` | string | ✓ | Provider type (`openai`) |
| `base_url` | string | | Custom API endpoint URL |
| `api_key` | string | | API key (plain text) |
| `api_key_env` | string | | Environment variable for API key |
| `model` | string | ✓ | Model to use |
| `headers` | map | | Additional HTTP headers |
| `timeout` | duration | | Request timeout |
| `max_retries` | int | | Maximum retry attempts |
| `rate_limit` | object | | Rate limiting settings |

### Rate Limit Settings

| Field | Type | Description |
|-------|------|-------------|
| `requests_per_minute` | int | Maximum requests per minute |
| `tokens_per_minute` | int | Maximum tokens per minute |

### Example Providers

```yaml
providers:
  # OpenAI
  - name: openai
    type: openai
    api_key_env: OPENAI_API_KEY
    model: gpt-4o
    max_retries: 3
    rate_limit:
      requests_per_minute: 60

  # DeepSeek
  - name: deepseek
    type: openai
    base_url: https://api.deepseek.com/v1
    api_key_env: DEEPSEEK_API_KEY
    model: deepseek-chat

  # Local vLLM
  - name: local
    type: openai
    base_url: http://localhost:8000/v1
    api_key: not-needed
    model: meta-llama/Llama-2-7b-chat-hf
```

---

## Schema Configuration

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Schema identifier |
| `path` | string | Path to custom schema file |
| `builtin` | bool | Use built-in schema |

```yaml
schemas:
  - name: instruction
    builtin: true
  - name: custom-qa
    path: ./schemas/qa.yaml
```

---

## Profiles

Profiles are named presets for quick configuration switching:

```yaml
profiles:
  production:
    description: High-quality production generation
    provider: openai
    schema: instruction
    generation:
      temperature: 0.5
      max_tokens: 2048
      seed: 42
      workers: 4
      checkpoint_every: 100

  fast:
    description: Quick iteration for testing
    provider: local
    schema: instruction
    generation:
      temperature: 0.9
      max_tokens: 512
      seed: 123
      workers: 8
```

### Generation Settings

| Field | Type | Description |
|-------|------|-------------|
| `temperature` | float | Sampling temperature (0-2) |
| `max_tokens` | int | Maximum tokens per response |
| `top_p` | float | Nucleus sampling parameter |
| `seed` | int64 | **Required.** Random seed for reproducibility |
| `workers` | int | Concurrent generation workers |
| `batch_size` | int | Batch size (if supported) |
| `checkpoint_every` | int | Save checkpoint every N samples |
| `system_prompt` | string | Custom system prompt |

---

## Logging Configuration

| Field | Type | Options | Description |
|-------|------|---------|-------------|
| `level` | string | `debug`, `info`, `warn`, `error` | Log verbosity |
| `format` | string | `text`, `json` | Log output format |
| `file` | string | | Optional log file path |

```yaml
logging:
  level: debug
  format: json
  file: ./kothaset.log
```

---

## Environment Variables

KothaSet supports environment variable references:

```yaml
providers:
  - name: openai
    api_key_env: OPENAI_API_KEY  # Reads from $OPENAI_API_KEY
```

Set environment variables:

```bash
# Windows PowerShell
$env:OPENAI_API_KEY = "sk-..."

# Linux/macOS
export OPENAI_API_KEY="sk-..."
```

---

## Configuration Precedence

1. Command-line flags (highest priority)
2. Environment variables
3. Configuration file (`.kothaset.yaml`)
4. Default values (lowest priority)
