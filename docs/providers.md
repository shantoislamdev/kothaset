# Provider Setup

KothaSet supports OpenAI and any OpenAI-compatible API (DeepSeek, vLLM, Ollama, etc.).

Provider credentials are stored in **`.secrets.yaml`** (PRIVATE).
The model selection is now global in **`kothaset.yaml`** (PUBLIC).

## Configuration Format

```yaml
# .secrets.yaml
providers:
  - name: <unique-identifier>
    type: openai  # Currently only 'openai' type supported
    api_key: env.<ENV_VAR_NAME>
    # api_key: <raw-key>  # Alternative: hardcoded
    base_url: <optional-url>
    timeout: <optional-duration>
    rate_limit:
      requests_per_minute: <int>
```

---

## 1. OpenAI

Standard setup for OpenAI API.

```yaml
# .secrets.yaml
providers:
  - name: openai
    type: openai
    api_key: env.OPENAI_API_KEY
    timeout: 2m
    max_retries: 3
```

## 2. DeepSeek

Use DeepSeek's OpenAI-compatible endpoint.

```yaml
# .secrets.yaml
providers:
  - name: deepseek
    type: openai
    base_url: https://api.deepseek.com/v1
    api_key: env.DEEPSEEK_API_KEY
```

Then in `kothaset.yaml`, set the model:

```yaml
global:
  provider: deepseek
  model: deepseek-chat
```

## 3. Local Models (Ollama / vLLM)

For local inference servers.

```yaml
# .secrets.yaml
providers:
  - name: local
    type: openai
    base_url: http://localhost:11434/v1  # Ollama default
    api_key: not-needed
```

In `kothaset.yaml`:

```yaml
global:
  provider: local
  model: llama2  # Model name in Ollama
```

---

## Environment Variables

We recommend using environment variables for API keys.

**Windows PowerShell:**
```powershell
$env:OPENAI_API_KEY = "sk-..."
```

**Linux/macOS:**
```bash
export OPENAI_API_KEY="sk-..."
```

---

## Provider Health Check Behavior

`kothaset provider test <name>` performs a minimal generation-style request to validate end-to-end readiness:

- authentication/API key
- model availability/validity
- network connectivity and endpoint behavior

This is stricter than metadata-only checks and better reflects real generation readiness.
