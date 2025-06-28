# Ask (GitHub) Docs CLI

[![CI](https://github.com/Ebonsignori/gh-ask-docs/workflows/CI/badge.svg)](https://github.com/Ebonsignori/gh-ask-docs/actions)

A [CLI Extension](https://docs.github.com/en/github-cli/github-cli/using-github-cli-extensions) for the [GitHub CLI](https://cli.github.com/) that lets you ask an LLM questions about GitHub using the official GitHub documentation.

Questions are answered by the AI search API provided by [docs.github.com](https://docs.github.com/en).

![Demonstration of asking `gh ask-docs` a question and getting a streamed response.](./docs/demo.gif)

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
| `--theme` | Color theme: `auto` (default), `light`, `dark` |
| `--debug` | Show raw NDJSON from the API for troubleshooting |

## Theme Detection

By default, the tool automatically detects your terminal's color theme using multiple methods:

1. **GitHub CLI theme** - Checks `GH_THEME` environment variable
2. **Terminal background** - Analyzes `COLORFGBG` environment variable 
3. **Terminal program** - Detects known terminals (Apple Terminal, iTerm2, VS Code, etc.)
4. **Platform defaults** - Falls back to sensible OS-specific defaults

You can override automatic detection with the `--theme` flag:

```bash
# Force light theme
gh ask-docs --theme light "How do I fork a repository?"

# Force dark theme  
gh ask-docs --theme dark "What are GitHub Pages?"

# Use automatic detection (default)
gh ask-docs --theme auto "How do I create an issue?"
```

## Development

Please see [development docs](./DEVELOPMENT.md).

## License

MIT