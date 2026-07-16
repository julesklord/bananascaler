# bananascaler — Makefile
# Run from the repository root.

BIN_DIR  := bin
BINARY   := $(BIN_DIR)/bananascaler
SRC_DIR  := src
GOFLAGS  := -ldflags="-s -w"

.PHONY: build install clean test tidy help

## build: Compile bananascaler to ./bin/bananascaler (with vet)
build:
	@mkdir -p $(BIN_DIR)
	cd $(SRC_DIR) && go vet ./...
	cd $(SRC_DIR) && go build $(GOFLAGS) -o ../$(BINARY) .
	@echo "✅ Binary ready: $(BINARY)"

## install: Install bananascaler to GOPATH/bin (available system-wide)
install:
	cd $(SRC_DIR) && go install $(GOFLAGS) .
	@echo "✅ Installed to $$(go env GOPATH)/bin/bananascaler"

## tidy: Sync go.mod and download dependencies (requires internet)
tidy:
	cd $(SRC_DIR) && go mod tidy

## test: Run the test suite
test:
	cd $(SRC_DIR) && go test ./...

## vet: Run go vet for static analysis
vet:
	cd $(SRC_DIR) && go vet ./...

## clean: Remove build artifacts
clean:
	rm -rf $(BIN_DIR)
	cd $(SRC_DIR) && go clean

## help: List available targets
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
