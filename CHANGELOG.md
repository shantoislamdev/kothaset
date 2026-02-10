# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-02-10

### Added
- **Context Configuration**: Added `context` system for better context management.
- **Input Handling**: Added support for inline string input via `-i` flag.

### Changed
- **CLI Interface**: Renamed `--seeds` flag to `--input` (`-i`) for clarity.
- **Configuration**: Removed `api_key_env` in favor of consistent `env.VAR_NAME` syntax for secrets.
- **Input Requirement**: Input file or string is now **mandatory**; removed hardcoded default topics.
- **Concurrency**: Improved worker pool implementation to prevent goroutine leaks.

### Fixed
- Fixed unbounded goroutine spawning in generator loop.
- Fixed various documentation inconsistencies and SEO issues on the website.

### Removed
- Removed `api_key_env` configuration field.
- Removed built-in default topics.

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

### Dependencies

- Go 1.22+
- spf13/cobra for CLI
- sashabaranov/go-openai for OpenAI API
