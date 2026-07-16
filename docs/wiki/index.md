# Wiki Index — bananascaler

Technical documentation for **bananascaler**. Start here.

## Main Sections

- [Architecture](./architecture.md): Design decisions and ADRs. (The 'Why'.)
- [Development](./development.md): Setup and contribution guide. (The 'How'.)
- [Git Hygiene](./hygiene.md): Commit standards and branch workflow. (The 'Rules'.)
- [Agent Protocol](./agent-sop.md): Instructions for AI collaborators. (The 'Law'.)

## Project Summary

Go CLI tool with Bubbletea TUI for GPU-accelerated neural video upscaling. Core pipeline in `src/internal/pipeline/`, TUI in `src/internal/tui/`, CLI in `src/cmd/`. Build with `make build`.

Dependencies: `ffmpeg`, `realesrgan-ncnn-vulkan`, Go ≥ 1.22.
