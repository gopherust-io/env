# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.0] - 2026-06-27

### Added

- **`ReloadConfig` / `MustReloadConfig`** — in-place config reload after env changes
- **`LoadConfigFrom(snap)`** — load from a custom `EnvSnapshot`
- **`env.Reload()`** — refresh cached snapshot (used by generated reloaders)
- **Cross-package nested structs** — `envgen` resolves imported types via `go/packages`
- **`reflectenv` package** — opt-in `reflectenv.Parse(&cfg)` without codegen
- **`examples/crosspkg`** — nested struct from another package

### Changed

- **`envgen`** uses `go/packages` for import-aware AST resolution
- Generated `LoadConfig` delegates to `ReloadConfig`

### Known limitations (updated)

- Production service migration (planned after feature complete)
- `reflectenv` allocates and uses reflection — not for hot paths

## [0.3.0] - 2026-06-27

### Added

- **golangci-lint** in CI — `govet` (incl. `fieldalignment`), `staticcheck`, `errcheck`, `modernize`, `prealloc`, `unconvert`, `makezero`, `nilerr`, `errorlint`, `durationcheck`, `misspell`
- **`make lint`** — local lint with workspace sync
- **`bench/remote.sh`** — run published benchmarks without cloning (`make bench-remote VERSION=v0.3.0`)
- **`bench/v0.4.0`** submodule tag for `go get github.com/gopherust-io/env/bench@v0.4.0`

### Changed

- **Struct tags** — `required:"true"`, `sensitive:"true"`, `expand:"true"` (valid Go syntax; bare flags fail `go vet`)
- **`envgen`** — `parser.ParseFile` per source file instead of deprecated `parser.ParseDir` / `ast.Package`
- **Struct field alignment** — `FieldError`, `Error`, `tag.Field`, examples (smaller structs; no hot-path regression)
- **CI** — `golangci-lint` replaces standalone `go vet` job

### Fixed

- Remote benchmark docs — `go test ...@version` is invalid; documented `bench/remote.sh` workflow
- `bench/v0.1.0` tag — lightweight tag for nested bench module on pkg.go.dev

## [0.2.0] - 2026-06-27

### Added

- **`expand` tag** — `${VAR}` and `$VAR` substitution via `env.Expand`
- **`time.Time` parsing** — `layout:"..."` struct tag (default RFC3339)
- **`.env` file support** — `LoadDotEnv`, `ParseDotEnv`, `SnapshotWithDotEnv`
- **`ParseTime`** runtime helper
- **GitHub Actions CI** — tests, race detector, vet
- **Installable `envgen`** — `go install github.com/gopherust-io/env/cmd/envgen@latest`

### Changed

- `go:generate` examples use `envgen` binary instead of `go run`
- README badges (CI, pkg.go.dev, Go Report Card)

## [0.1.0] - 2026-06-27

First public release of **env** — a codegen-first environment variable parser for Go.

### Highlights

- **Zero runtime dependencies** — the library uses only the Go standard library
- **Zero allocations** on config load in benchmarks (small / medium / large fixtures)
- **~157× faster** than `caarlos0/env` on a 10-field config (74 ns vs 11,619 ns, Apple M4 Pro)
- **No reflection at runtime** — `cmd/envgen` generates type-specific loaders at compile time

### Added

#### Runtime (`github.com/gopherust-io/env`)

- `EnvSnapshot` — single-pass index over `os.Environ()` with O(1) lookup
- `Snapshot()` — cached process-level snapshot; `ResetSnapshot()` for tests
- Primitive parsers without reflection: `string`, `bool`, all `int`/`uint` widths, `float32`/`float64`, `time.Duration`
- Collection parsers: `[]string`, `[]int`, `map[string]string` with configurable separators
- Aggregated errors: `FieldError`, `Error`, `NewError` — all field failures in one pass
- `Unmarshaler` interface for custom types
- `SensitiveMask` constant for generated `Masked()` output

#### Code generator (`cmd/envgen`)

- AST-based struct walker and Go source emitter
- Generated functions: `LoadConfig`, `MustLoadConfig`, `loadConfig` (internal)
- Generated `Masked()` when `sensitive` tags are present
- Struct tag support:
  - `env:"KEY"` — variable name
  - `default:"..."` — fallback value
  - `required` — error when unset
  - `prefix:"FOO_"` — nested struct prefix (composes recursively)
  - `sep:","` / `kvsep:":"` — slice and map parsing
  - `sensitive` — redaction in `Masked()`
  - `env:"-"` — skip field
- Automatic `UnmarshalEnv` detection for custom types in the same package

#### Tooling & examples

- `examples/basic` — nested config, slices, maps, sensitive fields, tests
- `Makefile` with `bench` and `bench-remote` targets
- `go.work.example` for local development across modules

#### Benchmarks (`github.com/gopherust-io/env/bench`)

| Competitor | Approach |
|------------|----------|
| stdlib | `os.LookupEnv` + `strconv` |
| caarlos0/env | Reflection |
| cleanenv | Reflection |
| goenv | Low-allocation reflection |
| envconfig | Reflection |
| viper | `AutomaticEnv` + mapstructure |

```bash
make bench-remote VERSION=v0.1.0
```

### Performance (darwin/arm64, Apple M4 Pro)

| Fixture | env | caarlos0/env | Speedup |
|---------|----:|-------------:|--------:|
| Small (10 fields) | 74 ns, 0 allocs | 11,619 ns, 220 allocs | **157×** |
| Medium (50 fields) | 398 ns, 0 allocs | 18,373 ns, 298 allocs | **46×** |
| Large (100 fields) | 946 ns, 0 allocs | 26,236 ns, 410 allocs | **28×** |

### Install

```bash
go get github.com/gopherust-io/env@v0.1.0
```

```go
//go:generate go run github.com/gopherust-io/env/cmd/envgen@v0.1.0 -type Config -output config_env_gen.go
```

### Migration from caarlos0/env

| caarlos0/env | env |
|--------------|-----|
| `env.Parse(&cfg)` | `cfg, err := LoadConfig()` |
| `envDefault:"8080"` | `default:"8080"` |
| `envPrefix:"DB_"` | `prefix:"DB_"` |
| `env:"HOST,required"` | `env:"HOST" required:"true"` |

### Known limitations

- Production service migration (planned when feature set is complete)

### Requirements

- Go 1.26.4+

[0.4.0]: https://github.com/gopherust-io/env/releases/tag/v0.4.0
[0.3.0]: https://github.com/gopherust-io/env/releases/tag/v0.3.0
[0.2.0]: https://github.com/gopherust-io/env/releases/tag/v0.2.0
[0.1.0]: https://github.com/gopherust-io/env/releases/tag/v0.1.0
