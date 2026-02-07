# KothaSet

[![CI](https://github.com/shantoislamdev/kothaset/actions/workflows/ci.yml/badge.svg)](https://github.com/shantoislamdev/kothaset/actions/workflows/ci.yml)
[![npm version](https://badge.fury.io/js/kothaset.svg)](https://www.npmjs.com/package/kothaset)
[![Go Report Card](https://goreportcard.com/badge/github.com/shantoislamdev/kothaset)](https://goreportcard.com/report/github.com/shantoislamdev/kothaset)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

**KothaSet** is a powerful CLI tool for generating high-quality datasets using Large Language Models (LLMs) as teacher models. Create diverse training data for fine-tuning smaller models (0.6B-32B parameters).

## Features

- 🔌 **Multi-Provider** — OpenAI, and OpenAI-compatible APIs (DeepSeek, vLLM, Ollama)
- 📋 **Flexible Schemas** — Instruction (Alpaca), Chat (ShareGPT), Preference (DPO), Classification
- 🌊 **Streaming Output** — Real-time generation with progress tracking
- 💾 **Resumable** — Atomic checkpointing, never lose progress
- 📦 **Multiple Formats** — JSONL, Parquet, HuggingFace datasets
- 🎲 **Diversity Control** — Seed files for topic coverage

---

## Installation

### npm (Recommended)
```bash
npm install -g kothaset
```

### Homebrew (macOS/Linux)
```bash
brew install shantoislamdev/tap/kothaset
```

### Scoop (Windows)
```bash
scoop bucket add kothaset https://github.com/shantoislamdev/scoop-bucket
scoop install kothaset
```

### Docker
```bash
docker pull ghcr.io/shantoislamdev/kothaset:latest
docker run ghcr.io/shantoislamdev/kothaset version
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
   kothaset generate -n 100 -s instruction -o dataset.jsonl
   ```

---

## Configuration

Edit `.kothaset.yaml` in your project directory:

```yaml
version: "1.0"
global:
  default_provider: openai
  default_schema: instruction

providers:
  - name: openai
    type: openai
    base_url: https://api.openai.com/v1
    api_key: env.OPENAI_API_KEY  # or raw key: sk-...
    model: gpt-4
    
  # Custom endpoint (DeepSeek, vLLM, etc.)
  - name: local
    type: openai
    base_url: http://localhost:8000/v1
    api_key: not-needed
    model: meta-llama/Llama-2-7b-chat-hf
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
kothaset generate -n 1000 -s instruction -o instructions.jsonl

# Chat conversations
kothaset generate -n 500 -s chat -o conversations.jsonl

# Preference pairs for DPO  
kothaset generate -n 500 -s preference -o dpo_data.jsonl
```

### Output Formats

```bash
# JSONL (default)
kothaset generate -n 100 -f jsonl -o dataset.jsonl

# Parquet
kothaset generate -n 100 -f parquet -o dataset.parquet

# HuggingFace datasets format
kothaset generate -n 100 -f hf -o ./my_dataset
```

### Advanced Options

```bash
# Use custom provider
kothaset generate -n 100 -p local -o dataset.jsonl

# Control diversity with seed file
kothaset generate -n 1000 --seeds topics.txt -o diverse.jsonl

# Resume interrupted generation
kothaset generate --resume dataset.jsonl.checkpoint

# Dry run (validate config)
kothaset generate --dry-run -n 100
```

---

## Documentation

- [Configuration Reference](docs/configuration.md)
- [Schema Guide](docs/schemas.md)
- [Provider Setup](docs/providers.md)
- [API Reference](docs/api.md)

---

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Apache 2.0 License. See [LICENSE](LICENSE).
