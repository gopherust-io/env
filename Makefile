.PHONY: bench bench-remote lint test

bench:
	@chmod +x bench/run.sh
	@./bench/run.sh

bench-remote:
	@chmod +x bench/remote.sh
	@VERSION=$(VERSION) COUNT=$(COUNT) ./bench/remote.sh

VERSION ?= latest
COUNT ?= 1

lint:
	@cp -n go.work.example go.work 2>/dev/null || true
	@go work sync 2>/dev/null || true
	golangci-lint run ./...

test:
	go test ./...
	@chmod +x bench/run.sh
	@COUNT=1 ./bench/run.sh -run=^$$
