# Changelog

All notable changes to this project will be documented in this file.
Format: [keepachangelog.com](https://keepachangelog.com) · Versioning: [semver.org](https://semver.org)

## [Unreleased]

## [0.2.0] - 2026-07-16

### Added

- **Interactive TUI**: Full Bubbletea dashboard with live progress bars, stage tracking, and scrollable logs.
- **Logger interface**: Pipeline output decoupled from `fmt.Printf` via `pipeline.Logger` interface.
- **`--no-tui` flag**: Explicit opt-out of the TUI for scripting, CI, and `nohup` usage.
- **`config.Validate()` method**: Centralized input validation in the config package.
- **`config.DefaultModel` constant**: Single source of truth for the default Real-ESRGAN model name.
- **TTY auto-detection**: Automatically uses TUI in terminals, plain text when piped.
- Dependencies: `charmbracelet/bubbletea`, `charmbracelet/lipgloss`, `charmbracelet/bubbles`.

### Changed

- **Go rewrite**: Core pipeline ported from Bash to Go with idiomatic project layout (`cmd/`, `internal/`).
- Pipeline `Run()` signature now accepts a `Logger` interface instead of writing to stdout directly.
- `cmd/root.go`: TTY detection via `golang.org/x/term`, launches Bubbletea or plain fallback.
- `Makefile`: `go vet` integrated into `build` target.

### Fixed

- `nvidia-smi` probe now has a 5-second timeout to prevent hanging on stuck drivers.
- `signal.Stop(sigCh)` added to prevent goroutine leak on signal handling.
- `cmd.Wait()` called after `cmd.Process.Kill()` to avoid race condition with done channel.
- Error messages in `hardware/detect.go` no longer embed `\n` (idiomatic Go error wrapping).

## [0.1.0] - 2026-07-16

### Added

- Repository initialization following the FMG Development Standard.
- `src/bananascaler.sh`: Core Bash script for GPU-accelerated video upscaling via Real-ESRGAN and FFmpeg.
- Automatic NVIDIA GPU detection with hardware-accelerated fallback to CPU.
- Atomic write pattern: frames rendered to `.tmp` file, renamed only on success.
- Session-scoped temp directories prefixed `bananascaler_` to prevent conflicts on parallel runs.
- `docs/` structure: wiki, AGENT.md, GEMINI.md, SOUL.md, IDENTITY.md, MEMORY.md.
- Hardened `.gitignore`, `LICENSE` (MIT), and `VERSION` (0.1.0).
