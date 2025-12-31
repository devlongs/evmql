# Contributing to EVMQL

Thank you for contributing to EVMQL!

## Quick Start

**Prerequisites:** Go 1.20+, golangci-lint

```bash
# Fork and clone the repo
git clone https://github.com/devlongs/evmql.git
cd evmql

# Run tests
go test ./...

# Build
go build -o evmql cmd/evmql/main.go
```

## Development

```bash
# Run tests with coverage
go test -cover ./...

# Format code
go fmt ./...

# Run linter
golangci-lint run

# Run the application
go run cmd/evmql/main.go
```

## Code Guidelines

- Follow standard Go conventions
- Write tests for new functionality (aim for 70%+ coverage)
- Use table-driven tests
- Validate all user input
- Add godoc comments for exported functions

## Commit Format

```
<type>: <subject>

<optional body>
```

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`

Example:
```
feat: add WHERE clause support

Implements basic filtering for query results.
```

## Pull Requests

1. Create a feature branch: `git checkout -b feature/your-feature`
2. Make changes and write tests
3. Run `go test ./...` and `golangci-lint run`
4. Push and create a PR

## Questions?

Open an issue for bugs or feature requests.

## License

MIT License
