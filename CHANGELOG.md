# Changelog

All notable changes to this project will be documented in this file.
Format: [keepachangelog.com](https://keepachangelog.com) · Versioning: [semver.org](https://semver.org)

## [Unreleased]

## [0.1.0] - 2026-07-16

### Added

- Repository initialization following the FMG Development Standard.
- `src/bananascaler.sh`: Core Bash script for GPU-accelerated video upscaling via Real-ESRGAN and FFmpeg.
- Automatic NVIDIA GPU detection with hardware-accelerated fallback to CPU.
- Atomic write pattern: frames rendered to `.tmp` file, renamed only on success.
- Session-scoped temp directories prefixed `bananascaler_` to prevent conflicts on parallel runs.
- `docs/` structure: wiki, AGENT.md, GEMINI.md, SOUL.md, IDENTITY.md, MEMORY.md.
- Hardened `.gitignore`, `LICENSE` (MIT), and `VERSION` (0.1.0).
