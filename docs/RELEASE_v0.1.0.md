# v0.1.0 — First release

**env** is a codegen-first environment variable parser for Go: struct tags at compile time, zero reflection and zero allocations at runtime.

## Install

```bash
go get github.com/gopherust-io/env@v0.1.0
```

```go
//go:generate go run github.com/gopherust-io/env/cmd/envgen@v0.1.0 -type Config -output config_env_gen.go
```

## What's included

### Runtime (zero dependencies)

- `EnvSnapshot` — one `os.Environ()` pass, indexed lookup
- Fast primitive parsers: `bool`, integers, floats, `time.Duration`, slices, maps
- Aggregated errors — all field failures reported in one pass
- `Unmarshaler` for custom types

### Codegen (`cmd/envgen`)

Generates `LoadConfig`, `MustLoadConfig`, and `Masked()` from struct tags:

`env`, `default`, `required`, `prefix`, `sep`, `kvsep`, `sensitive`, `env:"-"`

### Performance

Benchmarks on Apple M4 Pro (darwin/arm64):

| Config | env | caarlos0/env |
|--------|-----|--------------|
| 10 fields | **74 ns**, 0 allocs | 11,619 ns, 220 allocs |
| 50 fields | **398 ns**, 0 allocs | 18,373 ns, 298 allocs |
| 100 fields | **946 ns**, 0 allocs | 26,236 ns, 410 allocs |

Run benchmarks:

```bash
go test -bench=. -benchmem -count=1 github.com/gopherust-io/env/bench@v0.1.0
```

## Quick start

```go
type Config struct {
    Port int  `env:"PORT" default:"8080"`
    Debug bool `env:"DEBUG"`
}

cfg, err := LoadConfig()
```

See [README](https://github.com/gopherust-io/env#quick-start) for nested structs, slices, maps, and sensitive fields.

## Migration from caarlos0/env

| Before | After |
|--------|-------|
| `env.Parse(&cfg)` | `LoadConfig()` |
| `envDefault:"8080"` | `default:"8080"` |
| `envPrefix:"DB_"` | `prefix:"DB_"` |

## Not in this release

- `${VAR}` expansion
- `time.Time` parsing
- `.env` file support
- Hot-reload

Tracked in [CHANGELOG.md](./CHANGELOG.md).

## Full changelog

[CHANGELOG.md](./CHANGELOG.md#010---2026-06-27)
