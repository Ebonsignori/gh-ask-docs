# Ask (GitHub) Docs CLI

[![Test and Lint](https://github.com/Ebonsignori/gh-ask-docs/actions/workflows/test-and-lint.yml/badge.svg)](https://github.com/Ebonsignori/gh-ask-docs/actions/workflows/test-and-lint.yml)

A [CLI Extension](https://docs.github.com/en/github-cli/github-cli/using-github-cli-extensions) for the [GitHub CLI](https://cli.github.com/) that lets you ask an LLM questions about GitHub using the official GitHub documentation.

Questions are answered by the AI search API provided by [docs.github.com](https://docs.github.com/en).

![Demonstration of asking `gh ask-docs` a question and getting a streamed response.](./docs/demo.gif)

## Installation

Install this extension using the GitHub CLI:

```bash
gh extension install ebonsignori/gh-ask-docs
```

### Prerequisites

You'll need the [GitHub CLI](https://cli.github.com/) installed first:
- macOS: `brew install gh`
- Windows: `winget install GitHub.cli`
- Linux: See [installation instructions](https://github.com/cli/cli#installation)


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
| `--theme` | Color theme: `auto` (default), `light`, `dark` |
| `--debug` | Show raw NDJSON from the API for troubleshooting |

## Development

Please see [development docs](./DEVELOPMENT.md).

## License

MIT
