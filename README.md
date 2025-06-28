# Ask (GitHub) Docs CLI

[![CI](https://github.com/Ebonsignori/gh-ask-docs/workflows/CI/badge.svg)](https://github.com/Ebonsignori/gh-ask-docs/actions)

A [CLI Extension](https://docs.github.com/en/github-cli/github-cli/using-github-cli-extensions) for the [GitHub CLI](https://cli.github.com/) that lets you ask an LLM questions about GitHub using the official GitHub documentation.

Questions are answered by the AI search API provided by [docs.github.com](https://docs.github.com/en).

## Installation

Install this extension using the GitHub CLI:

```bash
gh extension install ebonsignori/gh-ask-docs
```

## Usage

```bash
gh ask-docs [flags] <query>
```

### Examples

Ask a basic question:
```bash
gh ask-docs "How do I create a pull request?"
```

Get sources with your answer:
```bash
gh ask-docs --sources "What are GitHub Actions?"
```

Query for Enterprise Server documentation:
```bash
gh ask-docs --version enterprise-server@3.17 "How to configure SAML?"
```

Stream raw markdown without rendering:
```bash
gh ask-docs --no-render "Git workflow best practices"
```

Query without streaming the response:
```bash
gh ask-docs --no-stream "How do I add GitHub Copilot to my IDE?"
```

## Flags

| Flag | Description |
|------|-------------|
| `--version` | Docs version (`free-pro-team`, `enterprise-cloud`, or `enterprise-server@<3.13-3.17>`) |
| `--sources` | Display reference links after the answer |
| `--no-render` | Stream raw Markdown without Glamour rendering |
| `--no-stream` | Don't stream answer, print only when complete (stdout-friendly) |
| `--wrap` | Word-wrap width when rendering (0 = no wrap) |
| `--debug` | Show raw NDJSON from the API for troubleshooting |


## Development

You can run this as a regular Go CLI executable without the `gh` CLI:

```bash
git clone https://github.com/ebonsignori/gh-ask-docs
cd gh-ask-docs
go build -o gh-ask-docs
./gh-ask-docs --sources "How do I create a branch?"
```

### Development Setup

Install development dependencies:

```bash
make dev-setup
```

### Common Development Tasks

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Run linter
make lint

# Format code
make fmt

# Run all checks (format, vet, lint, test)
make check

# Build binary
make build

# Clean build artifacts
make clean
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with race detection and coverage
go test -race -coverprofile=coverage.out -covermode=atomic ./...
```

## License

MIT