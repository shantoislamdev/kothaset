# Troubleshooting

## Quick Fixes

| Error | Fix |
|-------|-----|
| API key required | `export OPENAI_API_KEY="sk-..."` |
| Rate limit exceeded | Reduce workers: `-w 2` |
| Connection refused | Check if server is running |
| Timeout | Add `timeout: 5m` in config |

---

## Authentication

### "API key is required"
```bash
# Windows
$env:OPENAI_API_KEY = "sk-..."

# Linux/macOS
export OPENAI_API_KEY="sk-..."
```

### "Invalid API key"
Check at [platform.openai.com/api-keys](https://platform.openai.com/api-keys)

---

## Rate Limiting

### "Rate limit exceeded"

1. Reduce workers: `kothaset generate -w 2 ...`
2. Add to config:
   ```yaml
   providers:
     - name: openai
       rate_limit:
         requests_per_minute: 30
   ```

---

## Connection

### "Connection refused"
For local providers, ensure server is running:
```bash
# vLLM
python -m vllm.entrypoints.openai.api_server --model llama2 --port 8000

# Ollama
ollama serve
```

### "Request timed out"
Increase timeout in config:
```yaml
providers:
  - name: openai
    timeout: 5m
```

---

## Generation

### "Schema not found"
Valid schemas: `instruction`, `chat`, `preference`, `classification`

### "Provider not configured"
Ensure provider name in config matches `-p` flag.

### Slow generation

| Cause | Fix |
|-------|-----|
| High latency | Use faster model |
| Low concurrency | Increase: `-w 8` |
| Large responses | Reduce `--max-tokens` |

---

## Checkpoints

### Can't resume
Checkpoints saved as `<output>.checkpoint`:
```bash
kothaset generate --resume dataset.jsonl.checkpoint
```

---

## Config

### "Config file not found"
```bash
kothaset init
```

---

## Getting Help

**Debug mode:**
```yaml
logging:
  level: debug
```

**Dry run:**
```bash
kothaset generate --dry-run -n 100 --seed 42
```

**Report bugs:** [GitHub Issues](https://github.com/shantoislamdev/kothaset/issues)
