# Wiki Index — bananascaler

Technical documentation for **bananascaler**. Start here.

## Guides & Documentation

- [Installation](./installation.md): Install system dependencies (FFmpeg, Real-ESRGAN) and build **bananascaler** on different operating systems.
- [Usage Tutorial](./usage.md): Walkthrough of CLI usage, hotkeys inside TUI mode, and headless automation scripts.
- [Advanced Architecture](./architecture.md): Concurrency model, event dispatch, stage breakdown, and safety guarantees.
- [Development](./development.md): Setup, compilation, testing, and contribution guide.
- [Git Hygiene](./hygiene.md): Commit standards, semantic naming conventions, and branch workflow.
- [Agent Protocol](./agent-sop.md): Collaboration instructions and rules for AI agents.

## Project Summary

Go CLI tool with a Bubbletea TUI for GPU-accelerated neural video upscaling. Core pipeline in `src/internal/pipeline/`, TUI in `src/internal/tui/`, CLI in `src/cmd/`. Build with `make build`.

Dependencies: `ffmpeg`, `realesrgan-ncnn-vulkan`, Go ≥ 1.22.
