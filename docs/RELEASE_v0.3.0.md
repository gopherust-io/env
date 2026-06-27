# v0.3.0

## Install

```bash
go get github.com/gopherust-io/env@v0.3.0
go install github.com/gopherust-io/env/cmd/envgen@v0.3.0
```

```go
//go:generate envgen -type Config -output config_env_gen.go
```

## What's new

### Linting & CI

- **golangci-lint** in GitHub Actions — `govet` (incl. `fieldalignment`), `staticcheck`, `errcheck`, `modernize`, `prealloc`, and more
- Local: `make lint`

### Remote benchmarks

`go test` does not support `@version`. Use:

```bash
make bench-remote VERSION=v0.3.0
```

Bench submodule is tagged as `bench/v0.3.0`.

### Struct tags (breaking in docs only)

Boolean tags must use valid Go struct tag syntax:

```go
Host     string `env:"HOST" required:"true"`
Password string `env:"PASSWORD" sensitive:"true"`
BaseURL  string `env:"BASE_URL" default:"${NATS_URL}/api" expand:"true"`
```

Bare `required`, `sensitive`, `expand` fail `go vet` / CI.

### envgen

- Replaced deprecated `parser.ParseDir` / `ast.Package` with per-file `parser.ParseFile`
- No new dependencies; still stdlib-only root module

### Struct alignment

Runtime and example structs reordered for smaller memory footprint (`fieldalignment`). Does not change `LoadConfig()` performance characteristics.

## Full changelog

[CHANGELOG.md](../CHANGELOG.md#030---2026-06-27)
