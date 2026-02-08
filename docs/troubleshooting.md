# Troubleshooting Guide

Solutions for common issues when using KothaSet.

---

## Quick Fixes

| Error | Quick Fix |
|-------|-----------|
| API key required | `export OPENAI_API_KEY="sk-..."` |
| Rate limit exceeded | Add `-w 2` to reduce workers |
| Connection refused | Check if API server is running |
| Timeout | Add `timeout: 5m` in config |

---

## Authentication Errors

### "API key is required"

**Cause:** API key not set or not found.

**Solution:**

```bash
# Windows PowerShell
$env:OPENAI_API_KEY = "sk-..."

# Linux/macOS
export OPENAI_API_KEY="sk-..."

# Verify it's set
echo $env:OPENAI_API_KEY  # Windows
echo $OPENAI_API_KEY      # Linux/macOS
```

**Or set in config:**
```yaml
providers:
  - name: openai
    api_key: sk-...  # Direct (less secure)
```

### "Invalid API key"

**Cause:** API key is incorrect or expired.

**Solution:**
1. Check the key at [platform.openai.com/api-keys](https://platform.openai.com/api-keys)
2. Generate a new key if needed
3. Ensure no extra spaces or characters

---

## Rate Limiting

### "Rate limit exceeded" / Error 429

**Cause:** Too many requests to the API.

**Solutions:**

1. **Reduce workers:**
   ```bash
   kothaset generate -w 2 --seed 42 -n 100 -o dataset.jsonl
   ```

2. **Add rate limiting:**
   ```yaml
   providers:
     - name: openai
       rate_limit:
         requests_per_minute: 30
         tokens_per_minute: 60000
   ```

3. **Wait and retry:** The tool auto-retries with exponential backoff.

---

## Connection Issues

### "Connection refused"

**Cause:** Server not running or wrong URL.

**For local providers:**
```bash
# Check if server is running
curl http://localhost:8000/v1/models

# Start vLLM
python -m vllm.entrypoints.openai.api_server --model llama2 --port 8000

# Start Ollama
ollama serve
```

**Check config:**
```yaml
providers:
  - name: local
    base_url: http://localhost:8000/v1  # Verify port
```

### "Network error" / "DNS resolution failed"

**Cause:** Internet connectivity issues.

**Solutions:**
1. Check internet connection
2. Try `ping api.openai.com`
3. Check firewall/proxy settings
4. For corporate networks, configure proxy in config

### "Request timed out"

**Cause:** API taking too long to respond.

**Solutions:**

1. **Increase timeout:**
   ```yaml
   providers:
     - name: openai
       timeout: 5m
   ```

2. **Reduce max tokens:**
   ```bash
   kothaset generate --max-tokens 512 --seed 42 -n 100 -o dataset.jsonl
   ```

3. **Try a faster model:**
   ```bash
   kothaset generate -m gpt-4o-mini --seed 42 -n 100 -o dataset.jsonl
   ```

---

## Generation Issues

### "Schema not found"

**Cause:** Invalid schema name.

**Valid schemas:** `instruction`, `chat`, `preference`, `classification`

```bash
kothaset generate -s instruction --seed 42 -o dataset.jsonl
```

### "Provider not configured"

**Cause:** Provider name not in config.

**Solution:** Check `.kothaset.yaml`:
```yaml
providers:
  - name: openai  # Must match -p flag
    type: openai
    api_key_env: OPENAI_API_KEY
    model: gpt-4o
```

```bash
kothaset generate -p openai --seed 42 -o dataset.jsonl
```

### "Failed to parse response"

**Cause:** LLM returned unexpected format.

**Solutions:**
1. Lower temperature for more consistent output:
   ```bash
   kothaset generate --temperature 0.5 --seed 42 -o dataset.jsonl
   ```

2. The tool auto-retries failed parses.

### Slow generation

**Causes & solutions:**

| Cause | Solution |
|-------|----------|
| High latency | Use faster model |
| Low concurrency | Increase workers: `-w 8` |
| Large responses | Reduce `--max-tokens` |
| Rate limiting | Check for 429 errors |

---

## Checkpoint & Resume Issues

### "Failed to load checkpoint"

**Cause:** Checkpoint file corrupted or wrong format.

**Solutions:**
1. Check file exists: `ls dataset.jsonl.checkpoint`
2. Verify it's valid JSON: `cat dataset.jsonl.checkpoint | head`
3. If corrupted, start fresh (data already in output file is preserved)

### Can't find checkpoint

Checkpoints are saved as `<output>.checkpoint`:
```bash
# If output was dataset.jsonl
kothaset generate --resume dataset.jsonl.checkpoint
```

---

## Configuration Issues

### "Config file not found"

**Solution:** Initialize config:
```bash
kothaset init
```

Or specify path:
```bash
kothaset generate --config /path/to/.kothaset.yaml --seed 42 -o dataset.jsonl
```

### YAML parsing error

**Cause:** Invalid YAML syntax.

**Common fixes:**
- Use consistent indentation (2 spaces)
- Quote strings with special characters
- Check for missing colons

**Validate YAML:**
```bash
# Online: yamllint.com
# Or install yamllint
pip install yamllint
yamllint .kothaset.yaml
```

---

## Output Issues

### Empty output file

**Cause:** All samples failed.

**Check:**
1. Look for error messages during generation
2. Verify API key and connectivity
3. Try with `-n 1` to debug single sample

### Incomplete samples

**Cause:** `max_tokens` too low.

**Solution:**
```bash
kothaset generate --max-tokens 2048 --seed 42 -o dataset.jsonl
```

---

## Getting Help

### Debug mode

Enable verbose output:
```yaml
# .kothaset.yaml
logging:
  level: debug
```

### Dry run

Validate without generating:
```bash
kothaset generate --dry-run -n 100 --seed 42
```

### Report bugs

Include in your report:
1. Full error message
2. Command you ran
3. `.kothaset.yaml` (redact API keys)
4. OS and KothaSet version (`kothaset version`)

[Open an issue](https://github.com/shantoislamdev/kothaset/issues)
