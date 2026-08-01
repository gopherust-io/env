# env — Architecture

Zero-allocation, codegen-first environment → typed struct loader for Go.

## Overview

`github.com/gopherust-io/env` parses process environment (and optional `.env`) into typed structs. Compile-time `envgen` emits `LoadConfig` / `ReloadConfig` / `Masked()`; runtime does one `os.Environ()` index then direct field assignment—no reflection on the hot path. Used by **tel** and other gopherust-io services for config.

Ecosystem: [gopherust-io](https://github.com/gopherust-io/gopherust-io/blob/main/ARCHITECTURE.md)

## Layer / package overview

```
┌─────────────────────────────────────────────────────────────┐
│  Application code                                           │
│  calls generated LoadConfig / ReloadConfig / Masked         │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│  Generated loaders (*_env_gen.go)                           │
│  Parse* helpers + EnvSnapshot lookups                       │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│  Runtime library (package env)                              │
│  snapshot, parse, dotenv, expand, errors                    │
└─────────────────────────────────────────────────────────────┘

Compile time (separate process):
  cmd/envgen → internal/tag → internal/codegen → *_env_gen.go

Optional: reflectenv (reflection fallback for prototypes)
```

## Packages

| Path | Responsibility |
|------|----------------|
| `env` (root) | `EnvSnapshot`, `Parse*`, dotenv, expand, errors, unsafe bytesconv |
| `cmd/envgen` | CLI: load package AST, emit loaders (`-type`, `-list`, `-output`) |
| `internal/tag` | Struct-tag parsing, field resolution, nested prefixes |
| `internal/codegen` | Emits `*_env_gen.go` (`codegen.Generate`) |
| `reflectenv` | Opt-in reflection `Parse` / `ParseWithSnapshot` |
| `examples/` | End-to-end samples (`basic`, `minimal`, `crosspkg`) |
| `bench/` | Comparison benches vs caarlos0/env, viper, stdlib |

## Key design rules

- **Codegen vs runtime:** tags and AST live only in `envgen`; generated code and `package env` form the public runtime surface.
- **One snapshot:** prefer a single `EnvSnapshot` (or `FromEnviron` / `FromMap`) per load; avoid repeated `os.Getenv` in generated loops.
- **Zero alloc on happy path:** parsers and snapshot lookups stay allocation-sensitive; masking is an explicit API, not implicit logging.
- **Custom types:** implement `Unmarshaler` on the field type; `SensitiveMask` marks secrets for `Masked()`.
- **Dependency direction:** `internal/*` is for the generator only; apps import `github.com/gopherust-io/env` and generated package code—not `internal/`.

## Core APIs / interfaces

```go
// Runtime — package env
type EnvSnapshot struct { /* environ index */ }
func Snapshot() *EnvSnapshot
func (s *EnvSnapshot) Lookup(key string) (string, bool)

type Unmarshaler interface {
    UnmarshalEnv(value string) error
}

// Codegen — internal/codegen (used by envgen)
type Options struct { /* type, package, output, … */ }
func Generate(opts Options) error

// Generated pattern (per struct)
func LoadConfig() (Config, error)
func ReloadConfig() (Config, error)
func (c Config) Masked() Config
```

## Request / call flow

Example: generate and load `Config`

1. Author tags a struct (`env:"PORT" default:"8080"`) and `//go:generate envgen -type Config`.
2. `envgen` finds the type via `tag.FindStruct` / `CollectFields`.
3. `codegen.Generate` writes `LoadConfig` that calls `env.Snapshot()` and `Parse*` per field.
4. App calls `config.LoadConfig()` (or `LoadDotEnv` then load) and uses the typed value.

## Bootstrap / lifecycle

- **Library:** import-only; no process lifecycle. Call `LoadConfig` / `ReloadConfig` when the process needs config; `ResetSnapshot` only for tests.
- **Generator:** install `envgen`, run `go generate ./...` or `envgen -type Config` in CI/local before build.
- **Optional dotenv:** `env.LoadDotEnv(".env")` before load so the snapshot sees file-backed vars.

## Adding a feature

1. Prefer extending root `Parse*` / snapshot / tag semantics if many types need it.
2. Update `internal/tag` if a new tag key or nested-prefix rule is required.
3. Update `internal/codegen` so generated loaders emit the new behavior.
4. Add examples and benches when the public contract changes.
5. Keep `reflectenv` in sync only when the reflection path must match codegen semantics.

## Related docs

- [README](README.md)
- [Getting started](docs/GETTING_STARTED.md)
- [Examples](examples/basic/)
- [Changelog](CHANGELOG.md)
