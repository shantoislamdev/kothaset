# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.1] - 2026-02-07

### Added

- Initial release of KothaSet CLI
- **Core Features:**
  - Multi-provider support (OpenAI, OpenAI-compatible APIs)
  - 4 built-in schemas: instruction, chat, preference, classification
  - Concurrent generation with worker pool
  - Checkpointing for resumable generation
  - Real-time progress tracking
- **Output Formats:**
  - JSONL (streaming)
  - Parquet (columnar)
  - HuggingFace datasets format
- **Configuration:**
  - YAML configuration files
  - Environment variable support
  - Secret management (env vars, file references)
- **Distribution:**
  - npm package with postinstall binary download
  - Homebrew tap (macOS/Linux)
  - Scoop bucket (Windows)
  - Docker image

### Dependencies

- Go 1.22+
- spf13/cobra for CLI
- sashabaranov/go-openai for OpenAI API
