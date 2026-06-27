# Benchmarks

Separate Go module — competitor dependencies never leak into the main library.

## Run

```bash
# from repo root
make bench

# published module (after release)
go test -bench=. -benchmem -count=1 github.com/gopherust-io/env/bench@latest
```

## Regenerate fixtures

```bash
cd bench
go generate ./...
```

## Compare with benchstat

```bash
./bench/run.sh -count=10 > baseline.txt
# change something
./bench/run.sh -count=10 > candidate.txt
benchstat baseline.txt candidate.txt
```
