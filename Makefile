# bananascaler — Makefile
# Run from the repository root.

BIN_DIR  := bin
BINARY   := $(BIN_DIR)/bananascaler
GOFLAGS  := -ldflags="-s -w"
PREFIX   ?= /usr/local

.PHONY: build install clean test tidy help

## build: Compile bananascaler to ./bin/bananascaler (with vet)
build:
	@mkdir -p $(BIN_DIR)
	go vet ./...
	go build $(GOFLAGS) -o $(BINARY) .
	@echo "✅ Binary ready: $(BINARY)"

## install: Install bananascaler system-wide (e.g. /usr/local/bin)
install: build
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m 0755 $(BINARY) $(DESTDIR)$(PREFIX)/bin/bananascaler
	@echo "✅ Installed system-wide to $(DESTDIR)$(PREFIX)/bin/bananascaler"

## tidy: Sync go.mod and download dependencies (requires internet)
tidy:
	go mod tidy

## test: Run the test suite
test:
	go test ./...

## vet: Run go vet for static analysis
vet:
	go vet ./...

## clean: Remove build artifacts
clean:
	rm -rf $(BIN_DIR)
	go clean

## help: List available targets
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
