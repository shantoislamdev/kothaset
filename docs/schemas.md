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
kothaset generate -n 1000 -s instruction --seed 42 -i topics.txt -o dataset.jsonl
```

### Example Output

```jsonl
{"instruction":"Write a haiku about programming","input":"","output":"Lines of code unfold\nLogic dances through the night\nBugs hide in the dawn"}
{"instruction":"Explain what an API is","input":"","output":"An API (Application Programming Interface) is a set of protocols..."}
```

---

## Chat Schema

The `chat` schema generates multi-turn conversations for conversational model training.

### Structure

```json
{
  "system": "Optional system prompt",
  "conversations": [
    {"role": "user", "content": "Hello, can you help me?"},
    {"role": "assistant", "content": "Of course! What do you need help with?"},
    {"role": "user", "content": "I need to understand Python decorators"},
    {"role": "assistant", "content": "Python decorators are..."}
  ]
}
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `conversations` | array | ✓ | Array of message objects |
| `conversations[].role` | string | ✓ | Speaker (`user` or `assistant`) |
| `conversations[].content` | string | ✓ | Message content |
| `system` | string | | Optional system prompt |

### Usage

```bash
kothaset generate -n 500 -s chat --seed 42 -i conversations.txt -o conversations.jsonl
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
kothaset generate -n 500 -s preference --seed 42 -i pairs.txt -o dpo_data.jsonl
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
kothaset generate -n 1000 -s classification --seed 42 -i text_samples.txt -o labels.jsonl
```

---

## Metadata Note

JSONL output contains schema fields only (for example, `instruction/input/output` for the instruction schema).
Generation metadata (provider/model/tokens/seed/latency) is tracked internally for progress and checkpointing.

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
kothaset generate -n 1000 --seed 42 -i topics.txt -o diverse.jsonl
```

Each sample will be generated with a topic from the seed file, ensuring coverage across all specified areas.

---

## Output Formats

### JSONL (Default)

One JSON object per line:

```bash
kothaset generate -n 100 --seed 42 -i topics.txt -f jsonl -o dataset.jsonl
```
