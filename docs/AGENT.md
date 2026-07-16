# Agent SOP (Standard Operating Procedure) — bananascaler

This file is the entry point for any AI agent (Gemini, Claude, GPT, etc.) working on this repository. Read it before touching anything.

## General Instructions

1. **Familiarization**: Read `docs/wiki/index.md` and `docs/wiki/development.md` before making any changes.
2. **Compliance**: Follow the laws defined in `docs/wiki/agent-sop.md`. No exceptions.
3. **Identity**: Read `docs/SOUL.md` to understand tone and principles.

## Agent Initialization

This is a Go project with a Bubbletea TUI. Build with `make build` (includes `go vet`).

```bash
# Build the binary
make build

# Verify it runs
./bin/bananascaler --help

# Run with a real input file
./bin/bananascaler <test_input.mp4>

# Run without TUI (for scripting/CI)
./bin/bananascaler <test_input.mp4> --no-tui
```

## Key Paths

- `src/cmd/root.go` — CLI entry point and TTY detection.
- `src/internal/pipeline/pipeline.go` — Core 3-stage engine. Accepts a `Logger` interface.
- `src/internal/tui/` — Bubbletea TUI layer (model, styles, messages, adapter).
- `src/internal/config/config.go` — Configuration struct with validation.
- `src/internal/hardware/detect.go` — GPU and media probing.
- `docs/wiki/` — Architecture decisions and development notes.
- `CHANGELOG.md` — Update this on every `feat` or `fix`.
