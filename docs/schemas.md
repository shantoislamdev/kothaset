# Schema Guide

KothaSet provides built-in schemas for common dataset formats used in LLM fine-tuning. Each schema defines the structure, validation rules, and prompt templates for generating samples.

---

## Built-in Schemas

| Schema | Style | Use Case |
|--------|-------|----------|
| `instruction` | Alpaca-style | Supervised Fine-Tuning (SFT) |
| `chat` | ShareGPT | Multi-turn conversation training |
| `preference` | DPO/RLHF | Preference learning |
| `classification` | Text + Label | Classifier training |

---

## Instruction Schema

The `instruction` schema generates Alpaca-style instruction-response pairs for supervised fine-tuning.

### Structure

```json
{
  "instruction": "Explain the concept of recursion",
  "input": "",
  "output": "Recursion is a programming technique where..."
}
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `instruction` | string | ✓ | The task or question |
| `input` | string | | Optional context or input data |
| `output` | string | ✓ | Expected response |

### Usage

```bash
kothaset generate -n 1000 -s instruction --seed 42 -o dataset.jsonl
```

### Example Output

```jsonl
{"instruction":"Write a haiku about programming","input":"","output":"Lines of code unfold\nLogic dances through the night\nBugs hide in the dawn"}
{"instruction":"Explain what an API is","input":"","output":"An API (Application Programming Interface) is a set of protocols..."}
```

---

## Chat Schema

The `chat` schema generates multi-turn conversations in ShareGPT format, ideal for training conversational models.

### Structure

```json
{
  "conversations": [
    {"from": "human", "value": "Hello, can you help me?"},
    {"from": "gpt", "value": "Of course! What do you need help with?"},
    {"from": "human", "value": "I need to understand Python decorators"},
    {"from": "gpt", "value": "Python decorators are..."}
  ]
}
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `conversations` | array | ✓ | Array of message objects |
| `conversations[].from` | string | ✓ | Speaker (`human` or `gpt`) |
| `conversations[].value` | string | ✓ | Message content |

### Usage

```bash
kothaset generate -n 500 -s chat --seed 42 -o conversations.jsonl
```

---

## Preference Schema

The `preference` schema generates prompt/chosen/rejected triplets for Direct Preference Optimization (DPO) or RLHF training.

### Structure

```json
{
  "prompt": "Explain quantum computing",
  "chosen": "Quantum computing leverages quantum mechanical phenomena...",
  "rejected": "Quantum computing is like regular computing but faster..."
}
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `prompt` | string | ✓ | The input prompt |
| `chosen` | string | ✓ | Preferred/better response |
| `rejected` | string | ✓ | Less preferred response |

### Usage

```bash
kothaset generate -n 500 -s preference --seed 42 -o dpo_data.jsonl
```

---

## Classification Schema

The `classification` schema generates labeled text samples for training classifiers.

### Structure

```json
{
  "text": "I absolutely loved this product! Best purchase ever.",
  "label": "positive"
}
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `text` | string | ✓ | Input text to classify |
| `label` | string | ✓ | Classification label |
| `labels` | array | | Multiple labels (if applicable) |

### Usage

```bash
kothaset generate -n 1000 -s classification --seed 42 -o labels.jsonl
```

---

## Sample Metadata

Each generated sample includes metadata:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "fields": {
    "instruction": "...",
    "output": "..."
  },
  "metadata": {
    "generated_at": "2026-02-08T12:00:00Z",
    "provider": "openai",
    "model": "gpt-4o",
    "temperature": 0.7,
    "seed": 42,
    "tokens_used": 256,
    "latency": "1.2s",
    "topic": "programming"
  }
}
```

### Metadata Fields

| Field | Type | Description |
|-------|------|-------------|
| `generated_at` | timestamp | When the sample was created |
| `provider` | string | LLM provider used |
| `model` | string | Model used |
| `temperature` | float | Sampling temperature |
| `seed` | int64 | Random seed for reproducibility |
| `tokens_used` | int | Total tokens consumed |
| `latency` | duration | Generation time |
| `topic` | string | Topic/seed used (if any) |

---

## Controlling Diversity

Use seed files to guide topic diversity:

```bash
# topics.txt
machine learning
web development
data structures
algorithms
cloud computing
```

```bash
kothaset generate -n 1000 --seed 42 --seeds topics.txt -o diverse.jsonl
```

Each sample will be generated with a topic from the seed file, ensuring coverage across all specified areas.

---

## Output Formats

### JSONL (Default)

One JSON object per line:

```bash
kothaset generate -n 100 --seed 42 -f jsonl -o dataset.jsonl
```

### Parquet

Columnar format for efficient storage:

```bash
kothaset generate -n 100 --seed 42 -f parquet -o dataset.parquet
```

### HuggingFace Datasets

Directory structure compatible with `datasets` library:

```bash
kothaset generate -n 100 --seed 42 -f hf -o ./my_dataset
```

Load in Python:

```python
from datasets import load_from_disk
dataset = load_from_disk("./my_dataset")
```
