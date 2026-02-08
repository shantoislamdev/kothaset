# Examples

## Instruction Dataset (SFT)

```bash
kothaset generate -n 5000 -s instruction --seed 42 -o sft_dataset.jsonl
```

Output:
```json
{"instruction": "Explain TCP vs UDP", "input": "", "output": "TCP is..."}
```

---

## Chat Conversations

```bash
kothaset generate -n 2000 -s chat --seed 123 -o conversations.jsonl
```

Output:
```json
{"conversations": [{"from": "human", "value": "..."}, {"from": "gpt", "value": "..."}]}
```

---

## Preference Data (DPO)

```bash
kothaset generate -n 3000 -s preference --seed 456 -o dpo_data.jsonl
```

Output:
```json
{"prompt": "...", "chosen": "Good response", "rejected": "Bad response"}
```

---

## Classification

```bash
kothaset generate -n 10000 -s classification --seed 789 -o labels.jsonl
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
kothaset generate -n 1000 --seed 42 --seeds topics.txt -o diverse.jsonl
```

---

## Multi-Provider Setup

```yaml
# .kothaset.yaml
providers:
  - name: quality
    type: openai
    model: gpt-4o
  - name: fast
    type: openai
    model: gpt-4o-mini
```

```bash
kothaset generate -p quality -n 100 --seed 42 -o premium.jsonl
kothaset generate -p fast -n 5000 --seed 42 -o bulk.jsonl
```

---

## Large-Scale Generation

```bash
kothaset generate -n 50000 --seed 42 -w 8 -o large.jsonl

# Resume if interrupted
kothaset generate --resume large.jsonl.checkpoint
```

---

## Output Formats

```bash
kothaset generate -n 100 --seed 42 -f jsonl -o data.jsonl    # Default
kothaset generate -n 100 --seed 42 -f parquet -o data.parquet
kothaset generate -n 100 --seed 42 -f hf -o ./my_dataset
```

Load HuggingFace format:
```python
from datasets import load_from_disk
dataset = load_from_disk("./my_dataset")
```

---

## Cost Tips

1. Use `--dry-run` first to validate
2. Start small: `-n 10`
3. Use cheaper models: `-m gpt-4o-mini`
