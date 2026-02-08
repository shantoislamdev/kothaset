# Examples

Real-world examples for common dataset generation tasks.

---

## Table of Contents

- [Instruction Dataset for SFT](#instruction-dataset-for-sft)
- [Chat Conversations](#chat-conversations)
- [Preference Data for DPO](#preference-data-for-dpo)
- [Classification Dataset](#classification-dataset)
- [Topic-Diverse Generation](#topic-diverse-generation)
- [Multi-Provider Workflow](#multi-provider-workflow)
- [Large-Scale Generation](#large-scale-generation)

---

## Instruction Dataset for SFT

Generate Alpaca-style instruction-response pairs for supervised fine-tuning.

```bash
kothaset generate \
  -n 5000 \
  -s instruction \
  --seed 42 \
  --temperature 0.8 \
  --max-tokens 1024 \
  -o sft_dataset.jsonl
```

**Output format:**
```json
{"instruction": "Explain the difference between TCP and UDP", "input": "", "output": "TCP (Transmission Control Protocol) and UDP (User Datagram Protocol) are..."}
```

**Use with HuggingFace:**
```python
from datasets import load_dataset
dataset = load_dataset("json", data_files="sft_dataset.jsonl")
```

---

## Chat Conversations

Generate multi-turn conversations in ShareGPT format.

```bash
kothaset generate \
  -n 2000 \
  -s chat \
  --seed 123 \
  -o conversations.jsonl
```

**Output format:**
```json
{
  "conversations": [
    {"from": "human", "value": "Can you explain machine learning?"},
    {"from": "gpt", "value": "Machine learning is a subset of AI..."},
    {"from": "human", "value": "What are some common algorithms?"},
    {"from": "gpt", "value": "Some common ML algorithms include..."}
  ]
}
```

---

## Preference Data for DPO

Generate prompt/chosen/rejected triplets for Direct Preference Optimization.

```bash
kothaset generate \
  -n 3000 \
  -s preference \
  --seed 456 \
  --temperature 0.9 \
  -o dpo_data.jsonl
```

**Output format:**
```json
{
  "prompt": "Explain quantum entanglement",
  "chosen": "Quantum entanglement is a phenomenon where two particles become correlated...",
  "rejected": "Quantum entanglement is when things are connected somehow..."
}
```

---

## Classification Dataset

Generate labeled text samples for training classifiers.

```bash
kothaset generate \
  -n 10000 \
  -s classification \
  --seed 789 \
  -o labels.jsonl
```

**Output format:**
```json
{"text": "I absolutely loved this product!", "label": "positive"}
{"text": "Terrible experience, never again", "label": "negative"}
```

---

## Topic-Diverse Generation

Use a seed file to ensure topic coverage.

**1. Create `topics.txt`:**
```
machine learning
web development
databases
cloud computing
cybersecurity
mobile development
data science
DevOps
```

**2. Generate with topics:**
```bash
kothaset generate \
  -n 1000 \
  -s instruction \
  --seed 42 \
  --seeds topics.txt \
  -o diverse_dataset.jsonl
```

Each sample will cover one of the specified topics.

---

## Multi-Provider Workflow

Configure multiple providers for different use cases.

**.kothaset.yaml:**
```yaml
providers:
  - name: quality
    type: openai
    api_key_env: OPENAI_API_KEY
    model: gpt-4o

  - name: fast
    type: openai
    api_key_env: OPENAI_API_KEY
    model: gpt-4o-mini

  - name: deepseek
    type: openai
    base_url: https://api.deepseek.com/v1
    api_key_env: DEEPSEEK_API_KEY
    model: deepseek-chat
```

**Generate with different providers:**
```bash
# High quality (expensive)
kothaset generate -p quality -n 100 --seed 42 -o premium.jsonl

# Fast & cheap
kothaset generate -p fast -n 5000 --seed 42 -o bulk.jsonl

# Cost-effective alternative
kothaset generate -p deepseek -n 2000 --seed 42 -o deepseek.jsonl
```

---

## Large-Scale Generation

Generate large datasets efficiently with checkpointing.

```bash
# Start generation (auto-checkpoints every 50 samples)
kothaset generate \
  -n 50000 \
  -s instruction \
  --seed 42 \
  -w 8 \
  -o large_dataset.jsonl
```

**If interrupted, resume:**
```bash
kothaset generate --resume large_dataset.jsonl.checkpoint
```

**Progress output:**
```
Generating 50000 samples using openai (gpt-4o)
Schema: instruction | Output: large_dataset.jsonl

[24%] 12000/50000 samples | 3.2M tokens | $48.00 | 45.2/min | ETA: 14m20s
```

---

## Output Format Comparison

### JSONL (Streaming, Default)
```bash
kothaset generate -n 100 --seed 42 -f jsonl -o data.jsonl
```
- ✅ Streams during generation
- ✅ Easy to append/modify
- ✅ Resumable

### Parquet (Columnar)
```bash
kothaset generate -n 100 --seed 42 -f parquet -o data.parquet
```
- ✅ Efficient storage
- ✅ Fast column reads
- ⚠️ Written at end

### HuggingFace Format
```bash
kothaset generate -n 100 --seed 42 -f hf -o ./my_dataset
```
- ✅ Direct `datasets` compatibility
- ✅ Split into train/test ready

```python
from datasets import load_from_disk
dataset = load_from_disk("./my_dataset")
```

---

## Cost Optimization Tips

1. **Use `--dry-run` first** to validate config
   ```bash
   kothaset generate --dry-run -n 1000 --seed 42
   ```

2. **Start small** for testing
   ```bash
   kothaset generate -n 10 --seed 42 -o test.jsonl
   ```

3. **Use cheaper models** for bulk generation
   ```bash
   kothaset generate -m gpt-4o-mini -n 10000 --seed 42 -o bulk.jsonl
   ```

4. **Use DeepSeek** for cost-effective generation
   ```bash
   kothaset generate -p deepseek -n 5000 --seed 42 -o affordable.jsonl
   ```
