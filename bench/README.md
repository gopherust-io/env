# Benchmarks

```bash
make bench
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
./bench/run.sh -count=10 > candidate.txt
benchstat baseline.txt candidate.txt
```
