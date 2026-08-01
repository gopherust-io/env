.PHONY: align bench bench-remote lint test

GOALIGN_VERSION := v1.3.0
GOALIGN_BIN := $(HOME)/go/bin/goalign
GOALIGN_FLAGS := analyze -r --arch=amd64 --fail-on-findings --min-waste=1 -e bench/,examples/,bin/ .

$(GOALIGN_BIN):
	go install github.com/gopherust-io/goalign@$(GOALIGN_VERSION)

align: $(GOALIGN_BIN)
	$(GOALIGN_BIN) $(GOALIGN_FLAGS)

bench:
	@chmod +x bench/run.sh
	@./bench/run.sh

bench-remote:
	@chmod +x bench/remote.sh
	@VERSION=$(VERSION) COUNT=$(COUNT) ./bench/remote.sh

VERSION ?= latest
COUNT ?= 1

lint: align
	@cp -n go.work.example go.work 2>/dev/null || true
	@go work sync 2>/dev/null || true
	golangci-lint run ./...

test:
	go test ./...
	@chmod +x bench/run.sh
	@COUNT=1 ./bench/run.sh -run=^$$
