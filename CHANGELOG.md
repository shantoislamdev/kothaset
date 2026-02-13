# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Released]

## [1.1.1] - 2026-02-12

### Changed
- **Checkpoint Interval**: Default checkpoint save interval changed from 50 to 10 samples and is now configurable via `global.checkpoint_every` in `kothaset.yaml`.
- **Parquet Output**: Removed placeholder fallback behavior; parquet output is now native parquet-only.
- **Configuration Cleanup**: Removed legacy profile-system structures from config internals.

### Fixed
- **Docs Consistency**: Updated checkpoint resume path examples to use `.kothaset/<output>.checkpoint` and clarified current parquet validation behavior in CLI docs/FAQ/troubleshooting.

## [1.1.0] - 2026-02-10

### Added
- **Progress Bar**: Added visual progress bar during dataset generation for better user feedback.
- **Streaming Output**: Samples are now written immediately to output files, preventing memory leaks on large datasets.

### Changed
- **Checkpoint Location**: Moved checkpoint files from output directory to `.kothaset` cache directory for cleaner workspace.
- **Init Behavior**: Removed automatic output directory creation during init; directories are created on-demand during generation.
- **Memory Management**: Removed in-memory samples storage, significantly reducing memory footprint for large datasets.

### Fixed
- **Checkpoint Resume**: Fixed issue where existing data was not properly preserved when resuming from a checkpoint.
- **HuggingFace Output**: Fixed writer interface compatibility for HuggingFace dataset format.
- **Parquet Output**: Fixed writer interface compatibility for Parquet format.

## [1.0.3] - 2026-02-10

### Added
- **Init Gitignore**: `kothaset init` now intelligently handles `.gitignore` - creates new file if missing, appends missing entries if exists, or does nothing if already configured.

### Changed
- **Init Output**: Removed `--seed 42` from the example command in init next steps.

## [1.0.2] - 2026-02-10

### Added
- **Seed Options**: Added `--seed random` option to generate per-request random seeds for more diverse dataset generation.
- **Responsive Website**: Website is now fully responsive for small devices and mobile screens.
- **Dynamic Version**: Website now dynamically fetches and displays the latest npm version.

### Fixed
- **npm CLI**: Fixed CLI comma issue that was causing failures after npm install.
- **Terminal Display**: Fixed line wrapping issues in the terminal component on the website.
- **Documentation**: Made `--seed` parameter optional in documentation examples for clarity.

## [1.0.1] - 2026-02-10

### Fixed
- PyPI package now includes README content in project description.

## [1.0.0] - 2026-02-10

### Added
- **Context Configuration**: Added `context` system for better context management.
- **Input Handling**: Added support for inline string input via `-i` flag.

### Breaking Changes
- **Removed Output Formats**: Removed Parquet and HuggingFace output writers to simplify the core and improve robustness. JSONL is now the only supported output format.

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
