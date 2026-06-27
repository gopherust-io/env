# env

[![CI](https://github.com/gopherust-io/env/actions/workflows/ci.yml/badge.svg)](https://github.com/gopherust-io/env/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/gopherust-io/env.svg)](https://pkg.go.dev/github.com/gopherust-io/env)
[![Go Report Card](https://goreportcard.com/badge/github.com/gopherust-io/env)](https://goreportcard.com/report/github.com/gopherust-io/env)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Blazing-fast, zero-allocation environment configuration for Go.**

`github.com/gopherust-io/env` parses environment variables into typed structs using compile-time code generation. No reflection at runtime. No external dependencies. One `os.Environ()` pass, then direct field assignment.

```text
  caarlos0/env   11,619 ns/op   220 allocs
  viper           3,146 ns/op    70 allocs
  stdlib            150 ns/op     0 allocs
  env                  74 ns/op     0 allocs
```

---

## Install

```bash
go get github.com/gopherust-io/env@latest
go install github.com/gopherust-io/env/cmd/envgen@latest
```

---

## Quick start

```go
package config

import "time"

//go:generate envgen -type Config -output config_env_gen.go

type Database struct {
    Host     string `env:"HOST" required:"true"`
    Password string `env:"PASSWORD" sensitive:"true"`
    Port     int    `env:"PORT" default:"5432"`
}

type Config struct {
    Started time.Time         `env:"STARTED" layout:"2006-01-02"`
    Labels  map[string]string `env:"LABELS" sep:"," kvsep:":"`
    DB      Database          `prefix:"DB_"`
    BaseURL string            `env:"BASE_URL" default:"${NATS_URL}/api" expand:"true"`
    Tags    []string          `env:"TAGS" sep:","`
    Port    int               `env:"PORT" default:"8080"`
    Timeout time.Duration     `env:"TIMEOUT" default:"10s"`
    Debug   bool              `env:"DEBUG"`
}
```

```bash
go generate ./...
```

```go
_ = env.LoadDotEnv(".env") // optional, local dev

cfg, err := config.LoadConfig()
if err != nil {
    log.Fatal(err)
}

log.Printf("config: %+v", cfg.Masked())
```

---

## How it works

```mermaid
flowchart LR
    subgraph compile [Compile time]
        Struct[Config struct]
        Envgen[envgen]
        Gen[config_env_gen.go]
        Struct --> Envgen --> Gen
    end
    subgraph runtime [Runtime]
        Snap[EnvSnapshot]
        Load[LoadConfig]
        Snap --> Load
    end
    Gen --> Load
```

1. Define a struct with `env` tags.
2. `go generate` runs `envgen` and emits `LoadConfig`, `MustLoadConfig`, and `Masked()`.
3. `LoadConfig()` indexes the environment once and assigns fields with zero reflection.

---

## Struct tags

| Tag | Description |
|-----|-------------|
| `env:"KEY"` | Environment variable name |
| `default:"..."` | Value when unset |
| `required:"true"` | Error if unset and no default |
| `prefix:"FOO_"` | Prefix for nested struct fields |
| `sep:","` | Slice separator (default `,`) |
| `kvsep:":"` | Map key/value separator (default `:`) |
| `layout:"..."` | `time.Time` parse layout (default RFC3339) |
| `expand:"true"` | Expand `${VAR}` and `$VAR` in values |
| `sensitive:"true"` | Redact in `Masked()` |
| `env:"-"` | Skip field |

Nested prefixes compose: `prefix:"DB_"` + `env:"HOST"` → `DB_HOST`.

---

## Generated API

| Function | Description |
|----------|-------------|
| `LoadConfig()` | Parse env into `Config` |
| `MustLoadConfig()` | Panics on error |
| `(Config) Masked()` | Copy with sensitive fields redacted |

Errors are collected in one pass:

```text
env: DB.Host (DB_HOST): required; Port (PORT): parse: strconv.Atoi: parsing "abc": invalid syntax
```

---

## Local development (.env)

```go
_ = env.LoadDotEnv(".env")
cfg, err := config.LoadConfig()
```

`LoadDotEnv` fills unset variables from a file and refreshes the snapshot. Existing process variables are preserved.

For read-only merging without touching `os.Environ`:

```go
snap, err := env.SnapshotWithDotEnv(".env")
```

---

## Variable expansion

With the `expand` tag, defaults and values can reference other variables:

```go
BaseURL string `env:"BASE_URL" default:"${NATS_URL}/api" expand:"true"`
```

Supports `${VAR}` and `$VAR` syntax.

---

## Performance

```bash
make bench
make bench-remote VERSION=v0.3.0
```

| Fixture | env | caarlos0/env | Speedup |
|---------|----:|-------------:|--------:|
| 10 fields | **74 ns**, 0 allocs | 11,619 ns, 220 allocs | **157×** |
| 50 fields | **398 ns**, 0 allocs | 18,373 ns, 298 allocs | **46×** |
| 100 fields | **946 ns**, 0 allocs | 26,236 ns, 410 allocs | **28×** |

Measured on darwin/arm64, Apple M4 Pro. Full tables in [bench/README.md](bench/README.md).

---

## Migration from caarlos0/env

| caarlos0/env | env |
|--------------|-----|
| `env.Parse(&cfg)` | `LoadConfig()` |
| `envDefault:"8080"` | `default:"8080"` |
| `envPrefix:"DB_"` | `prefix:"DB_"` |
| `env:"HOST,required"` | `env:"HOST" required:"true"` |

---

## Custom types

```go
type Mode string

func (m *Mode) UnmarshalEnv(key, value string) error {
    switch value {
    case "dev", "staging", "prod":
        *m = Mode(value)
        return nil
    default:
        return fmt.Errorf("unknown mode %q", value)
    }
}
```

---

## Runtime API

```go
snap := env.Snapshot()
snap.Lookup("PORT")

env.ParseInt("8080")
env.ParseTime("2026-06-27", "2006-01-02")
env.Expand("${HOST}:${PORT}", snap)
env.LoadDotEnv(".env")
```

---

## Changelog

See [CHANGELOG.md](CHANGELOG.md).

## License

MIT — see [LICENSE](LICENSE).
