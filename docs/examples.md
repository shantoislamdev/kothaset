# Examples

## Instruction Dataset (SFT)

```bash
kothaset generate -n 5000 -s instruction --seed 42 -i topics.txt -o sft_dataset.jsonl
```

Output:
```json
{"instruction": "Explain TCP vs UDP", "input": "", "output": "TCP is..."}
```

---

## Chat Conversations

```bash
kothaset generate -n 2000 -s chat --seed 123 -i conversations.txt -o conversations.jsonl
```

Output:
```json
{"conversations": [{"from": "human", "value": "..."}, {"from": "gpt", "value": "..."}]}
```

---

## Preference Data (DPO)

```bash
kothaset generate -n 3000 -s preference --seed 456 -i pairs.txt -o dpo_data.jsonl
```

Output:
```json
{"prompt": "...", "chosen": "Good response", "rejected": "Bad response"}
```

---

## Classification

```bash
kothaset generate -n 10000 -s classification --seed 789 -i text_samples.txt -o labels.jsonl
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
kothaset generate -n 1000 --seed 42 -i topics.txt -o diverse.jsonl
```

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
kothaset generate -p quality -n 100 --seed 42 -i topics.txt -o premium.jsonl

# Use fast provider with cheaper model
kothaset generate -p fast -m gpt-4o-mini -n 5000 --seed 42 -i topics.txt -o bulk.jsonl
```

---

## Large-Scale Generation

```bash
kothaset generate -n 50000 --seed 42 -w 8 -i topics.txt -o large.jsonl

# Resume if interrupted (input file needed for validation)
kothaset generate --resume large.jsonl.checkpoint
```

---

## Output Formats

```bash
kothaset generate -n 100 --seed 42 -i topics.txt -f jsonl -o data.jsonl    # Default
kothaset generate -n 100 --seed 42 -i topics.txt -f parquet -o data.parquet
kothaset generate -n 100 --seed 42 -i topics.txt -f hf -o ./my_dataset
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
