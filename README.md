# KothaSet

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)
[![npm version](https://img.shields.io/npm/v/kothaset.svg)](https://www.npmjs.com/package/kothaset)
[![PyPI version](https://img.shields.io/pypi/v/kothaset.svg)](https://pypi.org/project/kothaset/)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

**KothaSet** is a powerful CLI tool for generating high-quality datasets using Large Language Models (LLMs) as teacher models. Create diverse training data for fine-tuning smaller models.

## Features

- **Multi-Provider** — OpenAI, and OpenAI-compatible APIs (DeepSeek, vLLM, Ollama)
- **Flexible Schemas** — Instruction (Alpaca), Chat (ShareGPT), Preference (DPO), Classification
- **Streaming Output** — Real-time generation with progress tracking
- **Resumable** — Atomic checkpointing, never lose progress
- **Multiple Formats** — JSONL, Native Parquet, HuggingFace datasets
- **Reproducible** — Required seed for deterministic LLM generation
- **Diversity Control** — Input files for sequential topic coverage
- **Validation** — Validate configs, schemas, datasets, and provider connectivity

---

## Installation

### pip (Python)
```bash
pip install kothaset
```

### npm (Node.js)
```bash
npm install -g kothaset
```

### Homebrew (macOS/Linux)
```bash
brew install shantoislamdev/tap/kothaset
```

### Binary Download
Download from [GitHub Releases](https://github.com/shantoislamdev/kothaset/releases).

### From Source
```bash
go install github.com/shantoislamdev/kothaset/cmd/kothaset@latest
```

---

## Quick Start

1. **Initialize configuration:**
   ```bash
   kothaset init
   ```

2. **Set your API key:**
   ```bash
   # Windows PowerShell
   $env:OPENAI_API_KEY = "sk-..."
   
   # Linux/macOS
   export OPENAI_API_KEY="sk-..."
   ```

3. **Generate a dataset:**
   ```bash
   kothaset generate -n 100 -s instruction --seed 42 -i topics.txt -o dataset.jsonl
   ```

---

## Configuration

## Configuration

KothaSet uses a **two-file configuration system** for better security and organization:

### 1. `kothaset.yaml` (Public)
Contains shared settings, context, and instructions. Safe to commit to git.

```yaml
version: "1.0"
global:
  provider: openai
  schema: instruction
  model: gpt-5.2
  concurrency: 4
  output_dir: ./output

# Context: Background info or persona injected into every prompt
context: |
  Generate high-quality training data for an AI assistant.
  The data should be helpful, accurate, and well-formatted.

# Instructions: Specific rules and guidelines for generation
instructions:
  - Be creative and diverse in topics and approaches
  - Vary the style and complexity of responses
  - Use clear and concise language
```

### 2. `.secrets.yaml` (Private)
Contains sensitive provider credentials. **Add this to your `.gitignore`!**

```yaml
providers:
  - name: openai
    type: openai
    api_key: env.OPENAI_API_KEY  # Reads from environment variable
    # api_key: sk-...            # Or hardcode key directly
    timeout: 1m
    rate_limit:
      requests_per_minute: 60

  # Custom endpoint example (DeepSeek, vLLM)
  - name: local
    type: openai
    base_url: http://localhost:8000/v1
    api_key: not-needed
```

---

## Usage

### Selecting a Schema

| Schema | Description | Use Case |
|--------|-------------|----------|
| `instruction` | Alpaca-style {instruction, input, output} | SFT |
| `chat` | ShareGPT multi-turn conversations | Chat fine-tuning |
| `preference` | {prompt, chosen, rejected} pairs | DPO/RLHF |
| `classification` | {text, label} pairs | Classifiers |

```bash
# Instruction dataset
kothaset generate -n 1000 -s instruction --seed 42 -i topics.txt -o instructions.jsonl

# Chat conversations
kothaset generate -n 500 -s chat --seed 123 -i conversations.txt -o conversations.jsonl

# Preference pairs for DPO  
kothaset generate -n 500 -s preference --seed 456 -i pairs.txt -o dpo_data.jsonl
```

### Output Formats

```bash
# JSONL (default)
kothaset generate -n 100 --seed 42 -i topics.txt -f jsonl -o dataset.jsonl

# Parquet
kothaset generate -n 100 --seed 42 -i topics.txt -f parquet -o dataset.parquet

# HuggingFace datasets format
kothaset generate -n 100 --seed 42 -i topics.txt -f hf -o ./my_dataset
```

### Advanced Options

```bash
# Use custom provider
kothaset generate -n 100 --seed 42 -i topics.txt -p local -o dataset.jsonl

# Control diversity with input file
kothaset generate -n 1000 --seed 42 -i topics.txt -o diverse.jsonl

# Resume interrupted generation
kothaset generate --resume dataset.jsonl.checkpoint

# Dry run (validate config)
kothaset generate --dry-run -n 100 --seed 42 -i topics.txt
```

---

## Documentation

**Getting Started**
- [Quick Start Guide](docs/quickstart.md)
- [Examples](docs/examples.md)

**Reference**
- [Configuration Reference](docs/configuration.md)
- [Schema Guide](docs/schemas.md)
- [Provider Setup](docs/providers.md)
- [CLI Reference](docs/api.md)

**Help**
- [FAQ](docs/faq.md)
- [Troubleshooting](docs/troubleshooting.md)

---

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Apache 2.0 License. See [LICENSE](LICENSE).
