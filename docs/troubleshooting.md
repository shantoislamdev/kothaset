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

### "Network error" / "DNS resolution failed"
1. Check internet connection
2. Try `ping api.openai.com`
3. Check firewall/proxy settings

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

### "Failed to parse response"
LLM returned unexpected format. Lower temperature for consistency:
```bash
kothaset generate --temperature 0.5 --seed 42 -o dataset.jsonl
```

### Slow generation

| Cause | Fix |
|-------|-----|
| High latency | Use faster model |
| Low concurrency | Increase: `-w 8` |
| Large responses | Reduce `--max-tokens` |

---

## Output Issues

### Empty output file
All samples failed. Check error messages, verify API key, try `-n 1` to debug.

### Incomplete samples
`max_tokens` too low:
```bash
kothaset generate --max-tokens 2048 --seed 42 -o dataset.jsonl
```

---

## Checkpoints

### "Failed to load checkpoint"
1. Verify file exists: `ls dataset.jsonl.checkpoint`
2. If corrupted, start fresh (existing output is preserved)

### Can't find checkpoint
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

### YAML parsing error
- Use 2-space indentation
- Quote strings with special chars
- Check for missing colons

---

## Validation Issues

### \"Schema not found\" in validate schema
```bash
kothaset schema list  # View available schemas
```

### \"Cannot access file\" in validate dataset
Check file exists and you have read permissions.

### \"Provider test failed\"
```bash
# Test your provider
kothaset provider test openai

# Common fixes:
# 1. Check API key is set
# 2. Verify base_url for custom endpoints
# 3. Increase timeout in .secrets.yaml
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

**Report bugs:** Include error message, command, config (redact keys), OS/version.

[GitHub Issues](https://github.com/shantoislamdev/kothaset/issues)
