# Development Guide

## Prerequisites

| Tool | Version | Purpose |
|---|---|---|
| `go` | ≥ 1.22 | Compiler |
| `ffmpeg` | Any recent | Frame I/O, encoding (runtime) |
| `realesrgan-ncnn-vulkan` | v0.2.5.0+ | Neural upscaling (runtime) |
| `ffprobe` | Bundled with ffmpeg | Framerate detection (runtime) |
| NVIDIA drivers + CUDA | Optional | Hardware acceleration (runtime) |

## Local Setup

```bash
# 1. Clone
git clone https://github.com/julesklord/bananascaler.git
cd bananascaler

# 2. Build
make build
# Binary ready at ./bin/bananascaler

# 3. Install runtime dependencies (Arch Linux / CachyOS)
sudo pacman -S ffmpeg

# Real-ESRGAN
mkdir -p ~/.local/share/realesrgan && cd ~/.local/share/realesrgan
curl -sL -O "https://github.com/xinntao/Real-ESRGAN/releases/download/v0.2.5.0/realesrgan-ncnn-vulkan-20220424-ubuntu.zip"
unzip realesrgan-ncnn-vulkan-20220424-ubuntu.zip && rm *.zip
chmod +x realesrgan-ncnn-vulkan
ln -sf ~/.local/share/realesrgan/realesrgan-ncnn-vulkan ~/.local/bin/realesrgan-ncnn-vulkan
```

## Useful Commands

```bash
# Build with vet
make build

# Install system-wide
make install

# Run tests
make test

# Static analysis
make vet

# Tidy dependencies
make tidy

# Clean build artifacts
make clean

# Show all targets
make help
```

## Project Layout

```
src/
├── main.go                          # Entrypoint
├── cmd/root.go                      # Cobra CLI + TTY detection
├── internal/
│   ├── config/config.go             # Config struct + validation
│   ├── hardware/detect.go           # GPU + media probing
│   ├── pipeline/pipeline.go         # Core 3-stage engine + Logger interface
│   └── tui/                         # Bubbletea TUI layer
│       ├── model.go                 # Bubbletea Model (Init/Update/View)
│       ├── styles.go                # Lipgloss styles
│       ├── messages.go              # Event types
│       └── pipeline_adapter.go      # Logger → tea.Msg bridge
├── go.mod
└── go.sum
```

## Architecture

The pipeline accepts a `Logger` interface. Two implementations exist:
- **`StdoutLogger`**: Plain text with ANSI colors (used when piped or `--no-tui`).
- **`tuiLogger`**: Sends events to the Bubbletea model via a channel (used in terminals).

To add a new consumer, implement `pipeline.Logger` and pass it to `pipeline.Run(cfg, log)`.

## Validation

After changes, verify:
1. `make build` succeeds (includes `go vet`)
2. `bananascaler --help` shows all flags
3. Running with a real video produces correct output
4. TUI renders in terminal, plain text when piped
