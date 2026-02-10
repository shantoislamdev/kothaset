# KothaSet

[![PyPI version](https://img.shields.io/pypi/v/kothaset.svg)](https://pypi.org/project/kothaset/)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

**KothaSet** is a powerful CLI tool for generating high-quality datasets using Large Language Models (LLMs) as teacher models. Create diverse training data for fine-tuning smaller models.

## Installation

```bash
pip install kothaset
```

## Quick Start

```bash
# Initialize configuration
kothaset init

# Set your API key
export OPENAI_API_KEY="sk-..."

# Generate a dataset
kothaset generate -n 100 -s instruction -i topics.txt -o dataset.jsonl
```

## Features

- **Multi-Provider** — OpenAI, and OpenAI-compatible APIs (DeepSeek, vLLM, Ollama)
- **Flexible Schemas** — Instruction (Alpaca), Chat (ShareGPT), Preference (DPO), Classification
- **Streaming Output** — Real-time generation with progress tracking
- **Resumable** — Atomic checkpointing, never lose progress
- **Multiple Formats** — JSONL, Native Parquet, HuggingFace datasets
- **Reproducible** — Required seed for deterministic LLM generation

## Documentation

Full documentation: [github.com/shantoislamdev/kothaset](https://github.com/shantoislamdev/kothaset)

## Also Available Via

- **npm**: `npm install -g kothaset`
- **Homebrew**: `brew install shantoislamdev/tap/kothaset`
- **Binary**: [GitHub Releases](https://github.com/shantoislamdev/kothaset/releases)

## License

Apache 2.0 License. See [LICENSE](https://github.com/shantoislamdev/kothaset/blob/main/LICENSE).
