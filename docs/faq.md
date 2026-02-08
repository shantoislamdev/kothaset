# Frequently Asked Questions

Common questions and answers about KothaSet.

---

## General

### What is KothaSet?

KothaSet is a CLI tool for generating high-quality datasets using LLMs as "teacher" models. These datasets are used to fine-tune smaller, more efficient models.

### What LLM providers are supported?

Currently, KothaSet supports:
- **OpenAI** (GPT-4, GPT-4o, GPT-3.5)
- **DeepSeek** (deepseek-chat, deepseek-reasoner)
- **Any OpenAI-compatible API** (vLLM, Ollama, LocalAI, etc.)

### Is an API key required?

Yes, you need an API key from your chosen provider. For OpenAI, get one at [platform.openai.com](https://platform.openai.com/api-keys).

---

## Installation

### How do I install KothaSet?

```bash
# npm (recommended)
npm install -g kothaset

# Homebrew (macOS/Linux)
brew install shantoislamdev/tap/kothaset

# From source
go install github.com/shantoislamdev/kothaset/cmd/kothaset@latest
```

### What are the system requirements?

- Windows, macOS, or Linux
- Internet connection for API calls
- For local models: sufficient RAM/GPU

---

## Usage

### Why is `--seed` required?

The `--seed` flag ensures reproducibility. With the same seed, you get the same dataset every time (assuming the same provider/model). This is essential for:
- Debugging and iteration
- Sharing reproducible experiments
- Consistent results across runs

### What schemas are available?

| Schema | Format | Use Case |
|--------|--------|----------|
| `instruction` | Alpaca-style | Supervised fine-tuning |
| `chat` | ShareGPT | Conversational AI |
| `preference` | DPO triplets | Preference learning |
| `classification` | Text + Label | Classifiers |

### How do I use a different model?

```bash
kothaset generate -m gpt-4o-mini --seed 42 -n 100 -o dataset.jsonl
```

### Can I resume interrupted generation?

Yes! KothaSet auto-saves checkpoints:

```bash
# Resume from checkpoint
kothaset generate --resume dataset.jsonl.checkpoint
```

---

## Output & Formats

### What output formats are supported?

- **JSONL** (default) — One JSON per line, streaming
- **Parquet** — Columnar, efficient storage
- **HuggingFace** — Directory format for `datasets` library

### How do I load the dataset in Python?

```python
# JSONL
from datasets import load_dataset
dataset = load_dataset("json", data_files="dataset.jsonl")

# Parquet
dataset = load_dataset("parquet", data_files="dataset.parquet")

# HuggingFace format
from datasets import load_from_disk
dataset = load_from_disk("./my_dataset")
```

---

## Cost & Performance

### How much does it cost?

Costs depend on your provider and model. Approximate costs per 1,000 samples:

| Model | ~Cost/1K samples |
|-------|------------------|
| gpt-4o | $2-5 |
| gpt-4o-mini | $0.10-0.30 |
| deepseek-chat | $0.05-0.15 |

KothaSet shows real-time cost estimates during generation.

### How can I reduce costs?

1. Use cheaper models (`gpt-4o-mini`, `deepseek-chat`)
2. Lower `--max-tokens` if samples can be shorter
3. Use `--dry-run` to validate without generating
4. Start with small batches to test quality

### How do I speed up generation?

Increase worker count:

```bash
kothaset generate -w 8 --seed 42 -n 1000 -o dataset.jsonl
```

Note: More workers = faster but may hit rate limits.

---

## Troubleshooting

### "API key is required"

Set your API key:
```bash
# Windows PowerShell
$env:OPENAI_API_KEY = "sk-..."

# Linux/macOS
export OPENAI_API_KEY="sk-..."
```

### "Rate limit exceeded"

Reduce workers or add rate limiting in config:
```yaml
providers:
  - name: openai
    rate_limit:
      requests_per_minute: 30
```

### "Connection refused"

For local providers, ensure the server is running:
```bash
# vLLM
python -m vllm.entrypoints.openai.api_server --model llama2 --port 8000

# Ollama
ollama serve
```

### Generation is slow

1. Check your internet connection
2. Reduce `--max-tokens`
3. Use a faster model
4. Increase workers (if not rate-limited)

---

## Data Quality

### How do I ensure diverse outputs?

Use a seed file with topics:

```bash
# topics.txt
machine learning
web development
databases

kothaset generate --seeds topics.txt --seed 42 -n 1000 -o diverse.jsonl
```

### Can I customize the prompts?

Yes, use `--system-prompt`:

```bash
kothaset generate --system-prompt "You are an expert Python tutor" --seed 42 -n 100 -o python.jsonl
```

### How do I validate output quality?

1. Generate a small batch first
2. Manually review samples
3. Adjust temperature (`--temperature 0.5` for focused, `0.9` for creative)
4. Iterate on system prompts

---

## Contributing

### How can I contribute?

See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines on:
- Setting up development environment
- Submitting pull requests
- Code style requirements

### Where do I report bugs?

Open an issue at [GitHub Issues](https://github.com/shantoislamdev/kothaset/issues).
