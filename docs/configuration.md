# Configuration Reference

## Quick Start

```bash
kothaset init
```

---

## Full Config Example

```yaml
version: "1.0"

global:
  default_provider: openai
  default_schema: instruction
  concurrency: 4
  timeout: 2m

providers:
  - name: openai
    type: openai
    api_key_env: OPENAI_API_KEY
    model: gpt-4o
    max_retries: 3
    rate_limit:
      requests_per_minute: 60

schemas:
  - name: instruction
    builtin: true

logging:
  level: info
  format: text
```

---

## Global Settings

| Field | Default | Description |
|-------|---------|-------------|
| `default_provider` | `openai` | Default LLM provider |
| `default_schema` | `instruction` | Default schema |
| `concurrency` | `4` | Concurrent workers |
| `timeout` | `2m` | Request timeout |

---

## Provider Config

| Field | Required | Description |
|-------|----------|-------------|
| `name` | ✓ | Provider identifier |
| `type` | ✓ | Provider type (`openai`) |
| `base_url` | | Custom API endpoint |
| `api_key_env` | | Env var for API key |
| `model` | ✓ | Model to use |
| `max_retries` | | Retry attempts |
| `rate_limit` | | Rate limiting |

### Example Providers

```yaml
providers:
  # OpenAI
  - name: openai
    type: openai
    api_key_env: OPENAI_API_KEY
    model: gpt-4o

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
    model: llama2
```

### Rate Limit Settings

| Field | Description |
|-------|-------------|
| `requests_per_minute` | Max requests per minute |
| `tokens_per_minute` | Max tokens per minute |

---

## Schema Config

```yaml
schemas:
  - name: instruction
    builtin: true
  - name: custom-qa
    path: ./schemas/qa.yaml
```

---

## Generation Settings

| Field | Description |
|-------|-------------|
| `temperature` | Sampling temperature (0-2) |
| `max_tokens` | Max tokens per response |
| `seed` | **Required.** Reproducibility seed |
| `workers` | Concurrent workers |
| `checkpoint_every` | Checkpoint frequency |

---

## Logging

```yaml
logging:
  level: info    # debug, info, warn, error
  format: text   # text, json
  file: ./log    # optional
```

---

## Environment Variables

```bash
# Windows
$env:OPENAI_API_KEY = "sk-..."

# Linux/macOS
export OPENAI_API_KEY="sk-..."
```

---

## Precedence

1. CLI flags (highest)
2. Environment variables
3. Config file
4. Defaults (lowest)
