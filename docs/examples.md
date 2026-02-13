# Examples

## Instruction Dataset (SFT)

```bash
kothaset generate -n 5000 -s instruction -i topics.txt -o sft_dataset.jsonl
```

Output:
```json
{"instruction": "Explain TCP vs UDP", "input": "", "output": "TCP is..."}
```

---

## Quick Generation (Inline Topic)

Generate a small dataset for a specific topic without creating an input file:

```bash
kothaset generate -n 5 -s instruction -i "python programming" -o python.jsonl
```

## Chat Conversations

```bash
kothaset generate -n 2000 -s chat -i conversations.txt -o conversations.jsonl
```

Output:
```json
{"conversations": [{"from": "human", "value": "..."}, {"from": "gpt", "value": "..."}]}
```

---

## Preference Data (DPO)

```bash
kothaset generate -n 3000 -s preference -i pairs.txt -o dpo_data.jsonl
```

Output:
```json
{"prompt": "...", "chosen": "Good response", "rejected": "Bad response"}
```

---

## Classification

```bash
kothaset generate -n 10000 -s classification -i text_samples.txt -o labels.jsonl
```

Output:
```json
{"text": "I loved this product!", "label": "positive"}
```

---

## Topic Diversity

**topics.txt:**
```
machine learning
web development
databases
```

```bash
kothaset generate -n 1000 -i topics.txt -o diverse.jsonl
```

---

## Reproducibility with Seed

For reproducible results, use `--seed`:

```bash
kothaset generate -n 1000 --seed 42 -i topics.txt -o reproducible.jsonl
```

> **Note:** `--seed` is optional. Use it when you need deterministic, reproducible output.

## Maximum Diversity with Random Seeds

To maximize diversity, use `--seed random` to generate a unique random seed for each AI request:

```bash
kothaset generate -n 1000 --seed random -i topics.txt -o diverse.jsonl
```

> **Tip:** Use `--seed random` when you want maximum variety in generated samples. Use `--seed 42` (or any fixed number) when you need reproducible results.

---

## Multi-Provider Setup

Define providers in `.secrets.yaml`:

```yaml
providers:
  - name: quality
    type: openai
    api_key: env.OPENAI_API_KEY
    timeout: 3m
  - name: fast
    type: openai
    api_key: env.OPENAI_API_KEY
    timeout: 1m
```

Generate using specific providers and models:

```bash
# Use quality provider with GPT-5.2 (default model)
kothaset generate -p quality -n 100 -i topics.txt -o premium.jsonl

# Use fast provider with cheaper model
kothaset generate -p fast -m gpt-4o-mini -n 5000 -i topics.txt -o bulk.jsonl
```

---

## Large-Scale Generation

```bash
kothaset generate -n 50000 -w 8 -i topics.txt -o large.jsonl

# Resume if interrupted (checkpoint stored in .kothaset/)
kothaset generate --resume .kothaset/large.jsonl.checkpoint
```

---

## Output Formats

```bash
kothaset generate -n 100 -i topics.txt -f jsonl -o data.jsonl    # Default
```

---

## Cost Tips

1. Use `--dry-run` first to validate
2. Start small: `-n 10`
3. Use cheaper models: `-m gpt-4o-mini`
