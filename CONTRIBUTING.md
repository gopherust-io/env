# Contributing

## Prerequisites

- Go version from [`go.mod`](go.mod)
- `make` helpers (optional)

## Development

```bash
make test          # unit tests
make test-race     # race detector
make bench         # benchmark suite
make lint          # static analysis/lint
```

## Pull requests

1. Keep changes focused and include tests for behavior changes.
2. Run `make test` and `make lint` before opening a PR.
3. Update `README.md` and `CHANGELOG.md` when public APIs or tags change.
4. Do not commit secrets or private credentials.

Security reports: see [SECURITY.md](SECURITY.md). Do not open public issues for vulnerabilities.
