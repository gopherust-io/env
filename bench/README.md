# Benchmarks

Local (from a clone):

```bash
make bench
```

Published module (no clone). `go test` does not accept `@version`; use:

```bash
make bench-remote VERSION=v0.3.0
# or
VERSION=v0.3.0 ./bench/remote.sh
```

Requires **Go 1.26.4+**. The bench submodule is tagged as `bench/v0.3.0` (same commit as root `v0.3.0`).

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
