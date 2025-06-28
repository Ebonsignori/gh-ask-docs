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

### Releasing
