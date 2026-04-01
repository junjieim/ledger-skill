GO ?= go
GOFLAGS ?=
LEDGER_DIR := ledger-cli
BINARY := ledger

.PHONY: help fmt test vet build tidy check clean

help:
	@echo "Available targets:"
	@echo "  fmt    - format Go code"
	@echo "  test   - run unit tests"
	@echo "  vet    - run go vet"
	@echo "  build  - build ledger CLI"
	@echo "  tidy   - tidy go modules"
	@echo "  check  - run fmt, test, and vet"
	@echo "  clean  - remove local build artifacts"

fmt:
	$(GO) -C $(LEDGER_DIR) fmt ./...

test:
	$(GO) -C $(LEDGER_DIR) test ./...

vet:
	$(GO) -C $(LEDGER_DIR) vet ./...

build:
	$(GO) -C $(LEDGER_DIR) build $(GOFLAGS) -o $(BINARY) .

tidy:
	$(GO) -C $(LEDGER_DIR) mod tidy

check: fmt test vet

clean:
	rm -f $(LEDGER_DIR)/$(BINARY)
