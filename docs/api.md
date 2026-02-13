# CLI Reference

Complete reference for all KothaSet commands and options.

---

## Commands

| Command | Description |
|---------|-------------|
| `kothaset init` | Initialize configuration file |
| `kothaset generate` | Generate dataset samples |
| `kothaset validate` | Validate configuration or setup |
| `kothaset schema` | Manage and view schemas |
| `kothaset provider` | Manage and test providers |
| `kothaset version` | Show version information |

---

## generate

Generate a dataset using an LLM as the teacher model.

### Synopsis

```bash
kothaset generate [flags]
```

### Required Flags

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--output` | `-o` | string | Output file path |
| `--input` | `-i` | string | Path to input file or inline topic |

Parent directories for `--output` are created automatically if they do not exist.

### Optional Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--count` | `-n` | int | `100` | Number of samples to generate |
| `--schema` | `-s` | string | `instruction` | Dataset schema |
| `--provider` | `-p` | string | from config | LLM provider |
| `--model` | `-m` | string | from config | Model to use |
| `--format` | `-f` | string | `jsonl` | Output format |
| `--temperature` | | float | `0.7` | Sampling temperature |
| `--max-tokens` | | int | `0` | Max tokens (0 = default/config) |
| `--system-prompt` | | string | | Custom system prompt |
| `--timeout` | | string | | Maximum total generation time (for example, `30m`, `2h`) |
| `--workers` | `-w` | int | `4` | Concurrent workers |
| `--seed` | | string | | Random seed for reproducibility (number or "random") |
| `--resume` | | string | | Resume from checkpoint |
| `--dry-run` | | bool | `false` | Validate without generating |

### Validation Rules

`kothaset generate` validates key numeric inputs before running:

- `--count` must be `>= 1`
- `--temperature` must be between `0` and `2.0`
- `--max-tokens` must be `>= 0`
- `--workers` must be `>= 1`

### Examples

```bash
# Basic generation with input file
kothaset generate -n 100 -s instruction -i topics.txt -o dataset.jsonl

# Single topic inline
kothaset generate -n 10 -s instruction -i "machine learning" -o ml_data.jsonl

# With custom provider and model
kothaset generate -n 500 -p openai -m gpt-4o -i prompts.txt -o output.jsonl

# Chat format with more workers
kothaset generate -n 1000 -s chat -w 8 -i conversations.txt -o chats.jsonl

# Preference data for DPO
kothaset generate -n 500 -s preference -i pairs.txt -o dpo.jsonl

# With seed for reproducibility
kothaset generate -n 100 --seed 42 -i topics.txt -o reproducible.jsonl

# With random seed per request (maximizes diversity)
kothaset generate -n 100 --seed random -i topics.txt -o diverse.jsonl

# Stop generation after an overall timeout
kothaset generate -n 1000 -i topics.txt -o dataset.jsonl --timeout 30m

# Resume interrupted generation (checkpoint stored in .kothaset/)
# Use the exact checkpoint filename present in `.kothaset/`
kothaset generate --resume .kothaset/<checkpoint-file>.checkpoint -i topics.txt

# Dry run to validate config
kothaset generate --dry-run -n 100 -i topics.txt
```

### Output

During generation, progress is displayed:

```
Generating 1000 samples using openai (gpt-4o)
Schema: instruction | Output: dataset.jsonl

[45%] 450/1000 samples | 125000 tokens | 15.2/min | ETA: 2m30s
```

On completion:

```
✓ Generation complete!
  Samples:      1000 successful, 0 failed
  Tokens:       278000
  Duration:     6m32s
  Output:       dataset.jsonl
```

---

## init

Initialize a new configuration file.

### Synopsis

```bash
kothaset init [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | bool | `false` | Overwrite existing config |

### Example

```bash
kothaset init
```

Creates `kothaset.yaml` (public) and `.secrets.yaml` (private) with default configuration.
`.secrets.yaml` is created with owner-only permissions (`0600` on Unix-like systems).

---

## Global Flags

These flags work with all commands:

| Flag | Type | Description |
|------|------|-------------|
| `--config` | string | Config file path (default: `kothaset.yaml`) |
| `--verbose`, `-v` | bool | Enable verbose output |
| `--quiet`, `-q` | bool | Suppress non-essential output |
| `--help` | bool | Show help |
| `--version` | bool | Show version |

---

## Exit Codes

| Code | Description |
|------|-------------|
| `0` | Success |
| `1` | General error |
| `2` | Configuration error |
| `3` | Provider error |
| `130` | Interrupted (Ctrl+C) |

---

## Checkpoints

KothaSet automatically saves checkpoints during generation:

- Checkpoint location: `.kothaset/<absolute-output-path-transformed>.checkpoint`
- Saved every 10 samples by default (configurable via `checkpoint_every` in global config)
- Resume with `--resume <checkpoint>`
- Tip: if unsure, list files in `.kothaset/` and pass the checkpoint path exactly.

### Checkpoint Contents

```json
{
  "timestamp": "2026-02-08T12:00:00Z",
  "config": { ... },
  "completed": 450,
  "failed": 2,
  "tokens_used": 125000
}
```

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `OPENAI_API_KEY` | OpenAI API key |
| `DEEPSEEK_API_KEY` | DeepSeek API key |
| `KOTHASET_CONFIG` | Custom config file path |

---

## Signals

| Signal | Behavior |
|--------|----------|
| `SIGINT` (Ctrl+C) | Graceful shutdown, save checkpoint |
| `SIGTERM` | Graceful shutdown, save checkpoint |

When interrupted:

```
Resume with: kothaset generate --resume .kothaset/<checkpoint-file>.checkpoint
```

---

## Retry and Write Durability Notes

- Retry timing uses exponential backoff with jitter and respects provider retry-after hints when available.
- Provider `rate_limit.requests_per_minute` is enforced during generation (`0` disables throttling).
- JSONL writes are buffered for performance; durability is enforced at checkpoint sync boundaries and on normal close.

---

## validate

Validate configuration and setup.

### Synopsis

```bash
kothaset validate [command]
```

### Subcommands

| Command | Description |
|---------|-------------|
| `config` | Validate configuration file structure |
| `schema` | Validate a schema definition |
| `dataset` | Validate an existing `.jsonl` dataset file |

### Examples

```bash
# Validate config file
kothaset validate config

# Validate a schema
kothaset validate schema instruction
# ✓ Schema 'instruction' is valid
#   Style:  instruction
#   Fields: 3 (2 required)

# Validate a dataset file
kothaset validate dataset output.jsonl
# Validating dataset: output.jsonl
#   Format: jsonl
#   Size:   12840 bytes
# ✓ Valid dataset
#   Rows: 50
```

Extension matching for dataset validation is case-insensitive (for example, `DATASET.JSONL` is recognized as `jsonl`).
Only `.jsonl` is currently supported for `validate dataset`.



---

## schema

Manage dataset schemas.

### Synopsis

```bash
kothaset schema [command]
```

### Subcommands

| Command | Description |
|---------|-------------|
| `list` | List available built-in schemas |
| `show` | Show detailed schema information |

### Examples

```bash
# List all schemas
kothaset schema list

# Show schema details
kothaset schema show instruction
# Name:        instruction
# Style:       instruction
# Description: Alpaca-style instruction-response pairs
# Fields:
#   NAME         TYPE    REQUIRED
#   instruction  string  yes
#   input        string  no
#   output       string  yes
```

---

## provider

Manage LLM providers.

### Synopsis

```bash
kothaset provider [command]
```

### Subcommands

| Command | Description |
|---------|-------------|
| `list` | List configured providers and their status |
| `test` | Test provider connectivity |

### Examples

```bash
# List configured providers
kothaset provider list

# Test provider connectivity
kothaset provider test openai
# Testing provider openai (openai)...
# ✓ Provider openai: connected
#   Type:     openai
#   Model:    gpt-5.2
#   Latency:  312ms
```
