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
| `--seed` | | int64 | Random seed for reproducibility |

### Optional Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--count` | `-n` | int | `100` | Number of samples to generate |
| `--schema` | `-s` | string | `instruction` | Dataset schema |
| `--provider` | `-p` | string | from config | LLM provider |
| `--model` | `-m` | string | from config | Model to use |
| `--format` | `-f` | string | `jsonl` | Output format |
| `--temperature` | | float | `0.7` | Sampling temperature |
| `--max-tokens` | | int | `2048` | Maximum tokens per response |
| `--system-prompt` | | string | | Custom system prompt |
| `--workers` | `-w` | int | `4` | Concurrent workers |
| `--input` | `-i` | string | | Path to input file (required) |
| `--resume` | | string | | Resume from checkpoint |
| `--dry-run` | | bool | `false` | Validate without generating |

### Examples

```bash
# Basic generation
kothaset generate -n 100 -s instruction --seed 42 -i topics.txt -o dataset.jsonl

# With custom provider and model
kothaset generate -n 500 -p openai -m gpt-4o --seed 123 -i prompts.txt -o output.jsonl

# Chat format with more workers
kothaset generate -n 1000 -s chat -w 8 --seed 456 -i conversations.txt -o chats.jsonl

# Preference data for DPO
kothaset generate -n 500 -s preference --seed 789 -i pairs.txt -o dpo.jsonl



# Different output formats
kothaset generate -n 100 --seed 42 -i topics.txt -f parquet -o dataset.parquet
kothaset generate -n 100 --seed 42 -i topics.txt -f hf -o ./hf_dataset

# Resume interrupted generation
kothaset generate --resume dataset.jsonl.checkpoint

# Dry run to validate config
kothaset generate --dry-run -n 100 --seed 42 -i topics.txt
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

- Checkpoint file: `<output>.checkpoint`
- Saved every 50 samples by default
- Resume with `--resume <checkpoint>`

### Checkpoint Contents

```json
{
  "timestamp": "2026-02-08T12:00:00Z",
  "config": { ... },
  "samples": [ ... ],
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
Resume with: kothaset generate --resume dataset.jsonl.checkpoint
```

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
| `schema` | Validate a schema definition (Phase 3) |
| `dataset` | Validate an existing dataset (Phase 3) |

### Examples

```bash
# Validate config file
kothaset validate config

# Validate specific config
kothaset validate config --config my-config.yaml
```

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
| `show` | Show schema details (Phase 3) |

### Examples

```bash
# List all schemas
kothaset schema list
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
| `test` | Test provider connection (Phase 2) |

### Examples

```bash
# List configured providers
kothaset provider list
```
