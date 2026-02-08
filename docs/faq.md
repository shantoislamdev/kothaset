# Frequently Asked Questions

---

## General

### What is KothaSet?
A CLI tool for generating datasets using LLMs as teacher models, for fine-tuning smaller models.

### What providers are supported?
OpenAI, DeepSeek, and any OpenAI-compatible API (vLLM, Ollama, etc.).

### Is an API key required?
Yes. Get one at [platform.openai.com](https://platform.openai.com/api-keys).

### Why is `--seed` required?
Ensures reproducibility—same seed = same dataset.

---

## Usage

### What schemas are available?

| Schema | Use Case |
|--------|----------|
| `instruction` | Supervised fine-tuning (Alpaca) |
| `chat` | Conversational AI (ShareGPT) |
| `preference` | DPO/RLHF training |
| `classification` | Text classifiers |

### How do I use a different model?
```bash
kothaset generate -m gpt-5.2 --seed 42 -i topics.txt -o dataset.jsonl
```

### Can I resume interrupted generation?
```bash
kothaset generate --resume dataset.jsonl.checkpoint
```

---

## Output

### What formats are supported?
- **JSONL** (default) — streaming
- **Parquet** — columnar storage  
- **HuggingFace** — `datasets` compatible

### How do I load in Python?
```python
from datasets import load_dataset
dataset = load_dataset("json", data_files="dataset.jsonl")

# Or HuggingFace format
from datasets import load_from_disk
dataset = load_from_disk("./my_dataset")
```

---



## Data Quality

### How to ensure diversity?
Use a diverse input file (seeds/topics):
```bash
kothaset generate -i topics.txt --seed 42 -n 1000 -o diverse.jsonl
```

### How to customize prompts?
```bash
kothaset generate --system-prompt "You are an expert Python tutor" --seed 42 -i python_topics.txt -o python.jsonl
```

### How to validate quality?
1. Generate small batch first (`-n 10`)
2. Manually review samples
3. Adjust `--temperature` (0.5=focused, 0.9=creative)

---

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for development setup and PR guidelines.

---

## More Help

- [Troubleshooting Guide](troubleshooting.md)
- [GitHub Issues](https://github.com/shantoislamdev/kothaset/issues)
