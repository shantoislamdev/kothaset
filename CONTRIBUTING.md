# Contributing to KothaSet

Thank you for your interest in contributing to KothaSet! We welcome contributions from the community to help make this tool better for everyone.

## Development Setup

### Prerequisites

- **Go**: Version 1.25 or later.
- **Node.js**: (Optional) For testing npm package builds.

### Building from Source

1.  **Clone the repository**:
    ```bash
    git clone https://github.com/shantoislamdev/kothaset.git
    cd kothaset
    ```

2.  **Install dependencies**:
    ```bash
    go mod download
    ```

3.  **Build the binary**:
    ```bash
    go build -o bin/kothaset ./cmd/kothaset
    ```

4.  **Run tests**:
    ```bash
    go test ./...
    ```

## Project Structure

- `cmd/`: Application entry points.
- `internal/`: Private application code.
    - `cli/`: Command-line interface definitions (Cobra).
    - `config/`: Configuration loading and parsing.
    - `generator/`: Core generation logic and worker pool.
    - `provider/`: LLM provider implementations (OpenAI, etc.).
    - `schema/`: Dataset schema definitions and validation.
    - `output/`: Output format writers (JSONL, Parquet).
- `npm/`: npm package wrapper and scripts.

## Pull Request Process

1.  Fork the repository and create your branch from `main`.
2.  If you've added code that should be tested, add tests.
3.  Ensure the test suite passes.
4.  Make sure your code follows the existing style (Go standard formatting).
5.  Open a Pull Request with a clear title and description of your changes.

## Style Guide

- Run `go fmt ./...` before committing.
- Exported functions and types should have comments.
- Handle all errors explicitly; avoid `panic` in library code.

## Code of Conduct

Please note that this project is released with a [Code of Conduct](CODE_OF_CONDUCT.md). By participating in this project you agree to abide by its terms.
