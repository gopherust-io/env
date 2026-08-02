# Benchmarks

Local (from a clone):

```bash
make bench
# or quick Small-only matrix:
cd bench && go test -bench=Small -benchmem -count=10
```

Published module (no clone). `go test` does not accept `@version`; use:

```bash
make bench-remote VERSION=v0.6.0
# or
VERSION=v0.6.0 ./bench/remote.sh
```

Requires **Go 1.26.5+** (or a toolchain that auto-downloads it). The bench submodule is tagged as `bench/v0.6.0` (same commit as root `v0.6.0`).

## Competitors

| Library | Approach | Benchmark suffix |
|---------|----------|------------------|
| **env** (this) | Codegen, no reflection | `Envgen` |
| stdlib | Hand-written `LookupEnv` + `strconv` | `Stdlib` |
| caarlos0/env | Reflection | `Carl` |
| cleanenv | Reflection | `Cleanenv` |
| goenv | Low-allocation reflection | `Goenv` |
| envconfig | Reflection | `Envconfig` |
| viper | `AutomaticEnv` + mapstructure | `Viper` |

Fixtures: Small (10 fields), Medium (50), Large (100). Names look like `BenchmarkSmallEnvgen`.

## Sample results (darwin/arm64, Apple M4 Pro)

Medians of `-count=10`. Raw output: [results.sample.txt](results.sample.txt).

### Small (10 fields)

| Library | ns/op | allocs/op | vs env |
|---------|------:|----------:|-------:|
| **env** | **74** | **0** | 1× |
| stdlib | 140 | 0 | ~2× |
| cleanenv | 2,737 | 57 | ~37× |
| envconfig | 3,068 | 81 | ~42× |
| viper | 3,596 | 70 | ~49× |
| goenv | 6,023 | 8 | ~82× |
| caarlos0/env | 12,622 | 244 | ~172× |

### Medium (50 fields)

| Library | ns/op | allocs/op | vs env |
|---------|------:|----------:|-------:|
| **env** | **402** | **0** | 1× |
| stdlib | 764 | 0 | ~2× |
| goenv | 8,200 | 6 | ~20× |
| viper | 11,622 | 236 | ~29× |
| cleanenv | 14,171 | 308 | ~35× |
| envconfig | 16,002 | 502 | ~40× |
| caarlos0/env | 20,532 | 322 | ~51× |

### Large (100 fields)

| Library | ns/op | allocs/op | vs env |
|---------|------:|----------:|-------:|
| **env** | **977** | **0** | 1× |
| stdlib | 1,593 | 0 | ~2× |
| goenv | 11,574 | 6 | ~12× |
| viper | 21,457 | 439 | ~22× |
| cleanenv | 28,516 | 609 | ~29× |
| caarlos0/env | 29,634 | 434 | ~30× |
| envconfig | 31,938 | 1,002 | ~33× |

## Regenerate fixtures

```bash
cd bench
go generate ./...
```

## Compare with benchstat

```bash
./bench/run.sh -count=10 > baseline.txt
./bench/run.sh -count=10 > candidate.txt
benchstat baseline.txt candidate.txt
```
