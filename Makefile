.PHONY: bench bench-remote test

bench:
	@chmod +x bench/run.sh
	@./bench/run.sh

bench-remote:
	go test -bench=. -benchmem -count=1 github.com/gopherust-io/env/bench@$(VERSION)

VERSION ?= latest

test:
	go test ./...
	@chmod +x bench/run.sh
	@COUNT=1 ./bench/run.sh -run=^$$
