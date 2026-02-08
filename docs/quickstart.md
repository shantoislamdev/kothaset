# Quick Start Guide

Get started with KothaSet in under 5 minutes.

---

## 1. Install

```bash
npm install -g kothaset
```

Other options: [Homebrew](https://brew.sh), [Binary downloads](https://github.com/shantoislamdev/kothaset/releases)

---

## 2. Set API Key

```bash
# Windows PowerShell
$env:OPENAI_API_KEY = "sk-..."

# Linux/macOS
export OPENAI_API_KEY="sk-..."
```

---

## 3. Initialize Config

```bash
kothaset init
```

This creates `.kothaset.yaml` with default settings.

---

## 4. Generate Your First Dataset

```bash
kothaset generate -n 10 -s instruction --seed 42 -o my_dataset.jsonl
```

**What this does:**
- `-n 10` → Generate 10 samples
- `-s instruction` → Alpaca-style instruction/response pairs
- `--seed 42` → Reproducible random seed (required)
- `-o my_dataset.jsonl` → Output file

---

## 5. View Results

```bash
# View first few lines
head -3 my_dataset.jsonl

# Or on Windows
Get-Content my_dataset.jsonl -Head 3
```

**Example output:**
```json
{"instruction":"Explain recursion","input":"","output":"Recursion is..."}
{"instruction":"Write a haiku","input":"about coding","output":"Lines of code..."}
```

---

## What's Next?

| Task | Command |
|------|---------|
| Generate more samples | `kothaset generate -n 1000 --seed 42 -o dataset.jsonl` |
| Use chat format | `kothaset generate -s chat --seed 42 -o chats.jsonl` |
| Use different model | `kothaset generate -m gpt-4o-mini --seed 42 -o dataset.jsonl` |
| Add topic diversity | `kothaset generate --seeds topics.txt --seed 42 -o diverse.jsonl` |

---

## Learn More

- [Configuration Reference](configuration.md)
- [Schema Guide](schemas.md)
- [Provider Setup](providers.md)
- [CLI Reference](api.md)
- [Examples](examples.md)
- [FAQ](faq.md)
