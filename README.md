# KothaSet

KothaSet is a powerful CLI tool for generating high-quality datasets using Large Language Models (LLMs) as teacher models. It allows you to create diverse training data for fine-tuning smaller models (0.6B-32B parameters).

## Features

- 🔌 **Multiple Providers**: Support for OpenAI, and OpenAI-compatible APIs (DeepSeek, vLLM, Ollama, etc.)
- 📋 **Flexible Schemas**: Built-in support for Instruction (Alpaca), Chat (ShareGPT), Preference (DPO), and Classification datasets.
- 🌊 **Stream Processing**: Real-time generation with progress tracking.
- 💾 **Resumable**: Atomic checkpointing ensures you never lose progress.
- 🎲 **Diversity**: Seed file support to ensure diverse topic coverage.

## Installation

### From Source

```bash
git clone https://github.com/shantoislamdev/kothaset.git
cd kothaset
go build -o kothaset.exe ./cmd/kothaset
```

## Quick Start

1. **Initialize Configuration**:
   ```bash
   ./kothaset init
   ```
   This creates a `.kothaset.yaml` file in the current directory.

2. **Set API Key**:
   Set your API key as an environment variable (recommended):
   ```bash
   # Windows PowerShell
   $env:OPENAI_API_KEY="sk-..."
   
   # Linux/Mac
   export OPENAI_API_KEY="sk-..."
   ```

3. **Generate Data**:
   ```bash
   ./kothaset generate -n 10 -o output.jsonl
   ```

## Configuration

KothaSet uses a layered configuration system. It looks for config files in:
1. `./.kothaset.yaml` (Project directory)
2. `~/.config/kothaset/config.yaml` (User directory)

### setting up Providers

You can configure providers in the `.kothaset.yaml` file.

#### OpenAI (Default)
```yaml
providers:
  - name: openai
    type: openai
    api_key_env: OPENAI_API_KEY  # Reads from env var
    model: gpt-4
```

#### Custom Endpoint (e.g., DeepSeek, LocalAI, vLLM)
To use a custom API URL (Base URL), configure a provider with `base_url`:

```yaml
providers:
  - name: local-model
    type: openai
    base_url: "http://localhost:8000/v1"  # Your API base URL
    api_key: "not-needed"                 # Or use api_key_env
    model: meta-llama/Llama-2-7b-chat-hf
```

To use this provider:
```bash
./kothaset generate --provider local-model ...
```

## Usage

### Selecting a Schema

Use the `-s` or `--schema` flag to select the dataset format.

**Available Schemas:**
- `instruction`: (Default) Alpaca-style `{instruction, input, output}`
- `chat`: ShareGPT-style `{conversations: [{role, content}, ...]}`
- `preference`: DPO style `{prompt, chosen, rejected}`
- `classification`: Text classification `{text, label}`

**Examples:**

```bash
# Generate Instruction Dataset (Default)
./kothaset generate -n 50 -s instruction -o instructions.jsonl

# Generate Chat/Conversation Dataset
./kothaset generate -n 50 -s chat -o conversation.jsonl

# Generate Preference/DPO Dataset
./kothaset generate -n 50 -s preference -o dpo_data.jsonl
```

### Controlling Diversity with Seeds

To ensure your dataset covers specific topics, verify you can use a seed file:

1. Create a text file `topics.txt` with one topic per line:
   ```text
   Python programming best practices
   History of the Roman Empire
   Quantum mechanics for beginners
   Healthy cooking recipes
   ```

2. Run generation with `--seeds`:
   ```bash
   ./kothaset generate -n 100 --seeds topics.txt -o diversity.jsonl
   ```

### Other Useful Flags

- `--dry-run`: Validate configuration without making API calls.
- `-w, --workers`: Number of concurrent generation workers (default: 4).
- `--resume`: Resume from a checkpoint file if generation was interrupted.

```bash
# Resume generation
./kothaset generate --resume output.jsonl.checkpoint
```
