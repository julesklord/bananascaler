# Changelog

All notable changes to this project will be documented in this file.
Format: [keepachangelog.com](https://keepachangelog.com) · Versioning: [semver.org](https://semver.org)

## [Unreleased]

### Changed

- `src/bananascaler.sh`: major logic overhaul
  - Add `set -euo pipefail` for strict error propagation
  - Replace string flags with Bash arrays — eliminates word-splitting on `$DEC_FLAGS`/`$ENC_FLAGS`
  - Add `cleanup` trap on EXIT/ERR/INT — temp dirs and `.tmp` files always removed
  - Add `--help` and `--gpu N` flags via proper argument parser loop
  - Add upfront dependency check (ffmpeg, ffprobe, realesrgan-ncnn-vulkan) before any processing
  - Add scale factor validation (must be 2, 3, or 4)
  - Add audio stream detection via `ffprobe` — handles video-only files gracefully
  - Add framerate validation — fails with clear error instead of silently producing broken output
  - Add colored output (cyan/green/yellow/red) with TTY detection fallback
  - Report extracted frame count after Stage 1

## [0.1.0] - 2026-07-16

### Added

- Repository initialization following the FMG Development Standard.
- `src/bananascaler.sh`: Core Bash script for GPU-accelerated video upscaling via Real-ESRGAN and FFmpeg.
- Automatic NVIDIA GPU detection with hardware-accelerated fallback to CPU.
- Atomic write pattern: frames rendered to `.tmp` file, renamed only on success.
- Session-scoped temp directories prefixed `bananascaler_` to prevent conflicts on parallel runs.
- `docs/` structure: wiki, AGENT.md, GEMINI.md, SOUL.md, IDENTITY.md, MEMORY.md.
- Hardened `.gitignore`, `LICENSE` (MIT), and `VERSION` (0.1.0).
