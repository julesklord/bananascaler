# Changelog

All notable changes to this project will be documented in this file.
Format: [keepachangelog.com](https://keepachangelog.com) · Versioning: [semver.org](https://semver.org)

## [Unreleased]

## [0.4.0] - 2026-07-16

### Added

- **Hardware profile system** (`internal/hardware/profile.go`): Auto-detects GPU tier (low-end / mid-range / high-end) via VRAM query through `nvidia-smi` and applies optimized pipeline parameters (tile size, model, encoding preset, CRF).
- **3 performance presets**: `--profile fast|balanced|quality` adapts speed/quality tradeoff to detected hardware. Each preset is customized per tier with VRAM-safe tile/model pairings.
- **`--auto` flag**: Explicitly enable auto-detection and apply the balanced preset. Profiles are also auto-detected when using the TUI without any `--profile` flag.
- **`bananascaler detect` subcommand**: Scans hardware and displays all available profiles (fast/balanced/quality) adapted to the detected GPU, plus a full reference table of all tier×preset combinations.
- **Tile-model VRAM safety check** (`CheckTileSafety`): Warns at pipeline start if the tile size may exceed safe limits for the detected VRAM and model, preventing OOM/SEGV crashes.
- **TUI profile cycling** (`p` key): Cycle between fast → balanced → quality presets in the file explorer, with profile displayed in the settings bar and pipeline header.
- **Profile-aware NVENC encoding**: NVENC preset (`-preset p1`–`p7`) is now set from the profile instead of relying on driver defaults, giving predictable encode speed/quality across runs.

### Changed

- **Tier boundaries adjusted**: low-end ≤4GB, mid-range 4–8GB, high-end ≥8GB (previously binary NVIDIA/no-NVIDIA detection only).
- **Tile sizes made VRAM-conservative**: Mid-range balanced uses `tile=300` with lightweight model (matching v0.3.0 behavior); heavier models (`x4plus-anime`, `x4plus`) only used at larger VRAM budgets with reduced tile sizes to prevent crashes.
- **Config extended** (`config.go`): Added `Profile`, `AutoDetect`, `PresetStr` fields and `ResolveProfile()` method for profile resolution before pipeline execution.
- **Pipeline parameterized** (`pipeline.go`): Hardcoded tile size (`-t 400`), JPEG quality (`-q:v 2`), NVENC flags, and x265 preset/CRF are now driven by the active profile instead of constants.

### Fixed

- GPU crash (SEGV_MAPERR) when using `realesrgan-x4plus-anime` model with tile size 400 on 6GB GPUs — tile sizes are now scaled to VRAM budget per model weight class.

## [0.3.0] - 2026-07-16

### Added

- **`bananascaler tui` subcommand**: Launches an interactive file-selection TUI in the current working directory, allowing users to browse folders, pick a video, and start upscaling — all without providing CLI arguments.
- **File explorer view**: Full keyboard-navigable file browser (`↑/↓`, `j/k`, `Enter`, `Backspace/h`) with visual distinction between directories, video files, and other files.
- **In-TUI settings**: Cycle scale factor (`s`), GPU index (`g`), and model (`m`) interactively before launching the pipeline from the explorer.
- **GPU-accelerated frame extraction**: Added `-hwaccel cuda` to the FFmpeg extraction stage so NVDEC decodes input frames on the GPU rather than the CPU.
- **Tile-based VRAM safety** for Real-ESRGAN: Added `-t 400` flag to `realesrgan-ncnn-vulkan` to prevent out-of-memory crashes on high-resolution or high-scale-factor runs.
- **`src/Makefile`**: Secondary Makefile inside `src/` for running `make build/install/test/tidy/clean` directly from the Go module directory.
- **System-wide installation via `make install`**: Both Makefiles now use `install(1)` with `PREFIX ?= /usr/local`, placing the binary in `/usr/local/bin/bananascaler` when invoked with `sudo`.

### Changed

- **Redesigned TUI design system** (`internal/tui/styles.go`, `model.go`):
  - New 9-color curated palette: warm gold `#F5C542`, mint green `#3DD68C`, calm blue `#5B9CF6`, amber `#FBBF24`, lavender `#A78BFA`, and a full blue-gray scale for hierarchy.
  - Premium progress bar with leading-edge glow (filled body `█` + amber lead `▓` + empty `░`); completes in green.
  - Stage rows show `n/3 — Name`, status icon, progress bar, and `% n/total` in a compact single-width layout.
  - Log entries now use icon prefixes: `✔ ok`, `⚠ warn`, `✖ err`, `◆ step`, `· info`.
  - Completion and error banners rendered in rounded-border boxes (`╭─╮`) colored green or red.
  - Footer keybind row: key in gold + description in gray, separated by `·` dividers — consistent across both views.
  - File-list selection rendered as full-width highlighted block (dark background + gold text) instead of a simple `❯` prefix.
- **PersistentFlags** in `cmd/root.go`: All flags (`--output`, `--scale`, `--gpu`, `--model`, `--verbose`, `--no-tui`) promoted to `PersistentFlags()` so they are inherited by the `tui` subcommand.
- **Conditional pipeline start** in `RunTUI`: The background goroutine is only launched if `cfg.Input != ""`, deferring launch to file selection when in explorer mode.

### Fixed

- `make install` previously used `go install` (installed to `$GOPATH/bin`); now correctly installs system-wide via the standard `install(1)` utility.

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
