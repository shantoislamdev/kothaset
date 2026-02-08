# Provider Setup

KothaSet supports OpenAI and OpenAI-compatible APIs. This guide covers setting up various providers.

---

## Supported Providers

| Provider | Type | Description |
|----------|------|-------------|
| OpenAI | `openai` | Official OpenAI API |
| DeepSeek | `openai` | DeepSeek AI (OpenAI-compatible) |
| vLLM | `openai` | Self-hosted vLLM server |
| Ollama | `openai` | Local Ollama server |
| Any OpenAI-compatible | `openai` | Any API following OpenAI spec |

---

## OpenAI

### Setup

1. Get an API key from [OpenAI Platform](https://platform.openai.com/api-keys)

2. Set the environment variable:
   ```bash
   # Windows PowerShell
   $env:OPENAI_API_KEY = "sk-..."
   
   # Linux/macOS
   export OPENAI_API_KEY="sk-..."
   ```

3. Configure in `.kothaset.yaml`:
   ```yaml
   providers:
     - name: openai
       type: openai
       api_key_env: OPENAI_API_KEY
       model: gpt-5.2
       max_retries: 3
       rate_limit:
         requests_per_minute: 60
   ```

### Usage

```bash
kothaset generate -n 1000 -p openai --seed 42 -o dataset.jsonl
```

---

## DeepSeek

DeepSeek provides cost-effective models via an OpenAI-compatible API.

### Setup

1. Get an API key from [DeepSeek Platform](https://platform.deepseek.com/)

2. Set the environment variable:
   ```bash
   export DEEPSEEK_API_KEY="sk-..."
   ```

3. Configure in `.kothaset.yaml`:
   ```yaml
   providers:
     - name: deepseek
       type: openai
       base_url: https://api.deepseek.com/v1
       api_key_env: DEEPSEEK_API_KEY
       model: deepseek-chat-3.2
       max_retries: 3
   ```

### Available Models

| Model | Notes |
|-------|-------|
| `deepseek-chat-3.2` | General purpose |
| `deepseek-reasoner-3.2` | Advanced reasoning |

### Usage

```bash
kothaset generate -n 1000 -p deepseek --seed 42 -o dataset.jsonl
```

---

## vLLM (Self-Hosted)

vLLM provides high-throughput inference for self-hosted models.

### Setup

1. Start vLLM server:
   ```bash
   python -m vllm.entrypoints.openai.api_server \
     --model meta-llama/Llama-2-7b-chat-hf \
     --port 8000
   ```

2. Configure in `.kothaset.yaml`:
   ```yaml
   providers:
     - name: vllm
       type: openai
       base_url: http://localhost:8000/v1
       api_key: not-needed
       model: meta-llama/Llama-2-7b-chat-hf
   ```

### Usage

```bash
kothaset generate -n 1000 -p vllm --seed 42 -o dataset.jsonl
```

---

## Ollama (Local)

Ollama provides an easy way to run models locally.

### Setup

1. Install and start Ollama:
   ```bash
   ollama serve
   ollama pull llama2
   ```

2. Configure in `.kothaset.yaml`:
   ```yaml
   providers:
     - name: ollama
       type: openai
       base_url: http://localhost:11434/v1
       api_key: ollama
       model: llama2
   ```

### Usage

```bash
kothaset generate -n 100 -p ollama --seed 42 -o dataset.jsonl
```

---

## Provider Configuration Options

### Full Configuration Example

```yaml
providers:
  - name: my-provider
    type: openai
    
    # API Endpoint
    base_url: https://api.example.com/v1
    
    # Authentication (choose one)
    api_key: sk-...              # Direct key
    api_key_env: MY_API_KEY      # Environment variable
    
    # Model
    model: gpt-5.2
    
    # Additional headers (optional)
    headers:
      X-Custom-Header: value
    
    # Timeouts and retries
    timeout: 2m
    max_retries: 3
    
    # Rate limiting
    rate_limit:
      requests_per_minute: 60
      tokens_per_minute: 100000
```

### Configuration Options

| Option | Type | Description |
|--------|------|-------------|
| `name` | string | Unique identifier for this provider |
| `type` | string | Provider type (currently `openai`) |
| `base_url` | string | Custom API endpoint |
| `api_key` | string | API key (plain text) |
| `api_key_env` | string | Environment variable name |
| `model` | string | Model identifier |
| `headers` | map | Additional HTTP headers |
| `timeout` | duration | Request timeout (e.g., `30s`, `2m`) |
| `max_retries` | int | Maximum retry attempts |
| `rate_limit.requests_per_minute` | int | Request rate limit |
| `rate_limit.tokens_per_minute` | int | Token rate limit |

---

## Multiple Providers

Configure multiple providers for different use cases:

```yaml
providers:
  # High quality generation
  - name: quality
    type: openai
    api_key_env: OPENAI_API_KEY
    model: gpt-4o

  # Fast, cheap generation
  - name: fast
    type: openai
    api_key_env: OPENAI_API_KEY
    model: gpt-4o-mini

  # Local development
  - name: local
    type: openai
    base_url: http://localhost:8000/v1
    api_key: not-needed
    model: llama2
```

Switch between providers:

```bash
kothaset generate -p quality --seed 42 -n 100 -o high_quality.jsonl
kothaset generate -p fast --seed 42 -n 1000 -o bulk.jsonl
kothaset generate -p local --seed 42 -n 50 -o test.jsonl
```

---





## Troubleshooting

### Authentication Errors

```
Error: provider "openai" not configured: API key is required
```

**Solution:** Ensure your API key is set:
```bash
echo $OPENAI_API_KEY  # Should show your key
```

### Rate Limiting

```
Error: rate limit exceeded
```

**Solution:** Reduce workers or add rate limiting:
```yaml
providers:
  - name: openai
    rate_limit:
      requests_per_minute: 30
```

### Connection Errors

```
Error: network error: connection refused
```

**Solution:** Verify the `base_url` is correct and the server is running.

### Timeout Errors

```
Error: request timed out
```

**Solution:** Increase the timeout:
```yaml
providers:
  - name: openai
    timeout: 5m
```
