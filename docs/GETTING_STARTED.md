# Getting started

A copy-paste guide for new users. For the full API see [README](../README.md).

## Requirements

- Go **1.26.4+**

## 1. Install

```bash
go get github.com/gopherust-io/env@latest
go install github.com/gopherust-io/env/cmd/envgen@latest
```

Verify `envgen` is on your `PATH` (`$(go env GOPATH)/bin`).

## 2. Define a config struct

Create `internal/config/config.go`:

```go
package config

//go:generate envgen -type Config -output config_env_gen.go

type Config struct {
    Port  int    `env:"PORT" default:"8080"`
    Debug bool   `env:"DEBUG"`
    Host  string `env:"HOST" default:"localhost"`
}
```

**Tag rules:** every modifier needs `key:"value"` form — e.g. `required:"true"`, not bare `required` (fails `go vet`).

Not sure which struct to generate?

```bash
envgen -dir ./internal/config -list
```

## 3. Generate the loader

From your module root:

```bash
go generate ./...
```

Or once:

```bash
envgen -dir ./internal/config -type Config
```

Commit `config_env_gen.go` to git.

## 4. Load config in main

```go
package main

import (
    "log"

    "github.com/gopherust-io/env"
    "myapp/internal/config"
)

func main() {
    _ = env.LoadDotEnv(".env") // optional, local dev only

    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatal(err)
    }
    _ = cfg
}
```

`MustLoadConfig()` panics on error — fine for `main`, use `LoadConfig()` in libraries.

## 5. Local `.env` (optional)

```bash
# .env
PORT=3000
DEBUG=true
```

```go
_ = env.LoadDotEnv(".env")
cfg, err := config.LoadConfig()
```

Unset keys in the file are filled; existing process env vars win.

## CI

Install `envgen` before `go generate`:

```yaml
- uses: actions/setup-go@v5
  with:
    go-version: "1.26.4"
- run: go install github.com/gopherust-io/env/cmd/envgen@latest
- run: go generate ./...
- run: go test ./...
```

## Cheatsheet

| Task | Command / API |
|------|----------------|
| List structs | `envgen -dir . -list` |
| Generate | `go generate ./...` |
| Load | `config.LoadConfig()` |
| Reload in-place | `config.ReloadConfig(&cfg)` |
| Safe logging | `cfg.Masked()` |
| Without codegen | `reflectenv.Parse(&cfg)` |
| Nested struct | `` DB Database `prefix:"DB_"` `` |
| Cross-package nested | `` DB db.Database `prefix:"DB_"` `` |

## Troubleshooting

| Problem | Fix |
|---------|-----|
| `envgen: command not found` | `go install .../cmd/envgen@latest` and add `GOPATH/bin` to `PATH` |
| `go:generate` does nothing | Add `//go:generate` line above struct; run from module root |
| `go vet` / CI fails on tags | Use `required:"true"`, not `required` |
| `go test ...@version` for bench | Use `make bench-remote VERSION=v0.4.0` — `go test` has no `@version` |
| Changes to env not picked up | Call `ReloadConfig(&cfg)` or `env.ResetSnapshot()` in tests |
| Imported nested type fails | Field must be exported; imported package must be a struct with exported fields |

## Examples in this repo

| Path | What it shows |
|------|----------------|
| [examples/minimal](../examples/minimal/) | Smallest config (3 fields) |
| [examples/basic](../examples/basic/) | Nested, slices, maps, sensitive, expand, time |
| [examples/crosspkg](../examples/crosspkg/) | Nested struct from another package |

## Migrating from caarlos0/env

| caarlos0/env | env |
|--------------|-----|
| `env.Parse(&cfg)` | `cfg, err := LoadConfig()` |
| `envDefault:"8080"` | `default:"8080"` |
| `envPrefix:"DB_"` | `prefix:"DB_"` on nested field |
| `env:"HOST,required"` | `env:"HOST" required:"true"` |

Remove `github.com/caarlos0/env` from `go.mod` after migration.
