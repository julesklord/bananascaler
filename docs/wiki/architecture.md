# Architecture and Decisions (ADRs) — bananascaler

Design decisions for **bananascaler**.

## System Overview

A 3-stage sequential pipeline coordinated by a Go CLI with a Bubbletea TUI:
1. **Extract** — FFmpeg decodes source video to JPEG frames.
2. **Upscale** — Real-ESRGAN (ncnn-vulkan) applies neural super-resolution per frame.
3. **Re-encode** — FFmpeg re-assembles frames + muxes original audio → atomic rename.

---

## ADRs

### ADR 0001: JPEG for intermediate frames

**Status**: Accepted  
**Date**: 2026-07-16

#### Context

Frame extraction produces thousands of intermediate images. PNG (lossless) is the naive choice but generates 3–5× more disk I/O than JPEG.

#### Decision

Use JPEG at `-q:v 2` (near-lossless). Real-ESRGAN input quality at this level introduces no perceptible difference in the upscaled output.

#### Consequences

- **Positive**: ~60–70% reduction in `/tmp/` disk usage; lower I/O pressure on NVMe storage.
- **Negative**: Technically lossy intermediate. Unacceptable for archival workflows requiring pixel-perfect round-trips.

---

### ADR 0002: ncnn-Vulkan backend over CUDA-only

**Status**: Accepted  
**Date**: 2026-07-16

#### Context

Real-ESRGAN offers both a CUDA-only binary and an ncnn-Vulkan binary. CUDA requires NVIDIA; Vulkan runs on NVIDIA, AMD, and Intel Arc.

#### Decision

Use `realesrgan-ncnn-vulkan`. GPU vendor portability outweighs any CUDA-specific optimizations.

#### Consequences

- **Positive**: Works on any Vulkan-capable GPU, including iGPUs.
- **Negative**: ncnn may be slightly slower than native CUDA on high-end NVIDIA cards.

---

### ADR 0003: Atomic output via `.tmp` rename

**Status**: Accepted  
**Date**: 2026-07-16

#### Context

A multi-hour encode interrupted at 99% leaves a corrupt file that passes size checks and silently deceives the user.

#### Decision

Encode to `output.mp4.tmp`. Rename to `output.mp4` only on exit code 0.

#### Consequences

- **Positive**: Interrupted runs always leave either a valid output or a clearly-named `.tmp` artifact.
- **Negative**: Requires disk space for both `.tmp` and final file simultaneously during the rename moment (negligible: rename is instantaneous on same filesystem).

---

### ADR 0004: Logger interface for pipeline decoupling

**Status**: Accepted  
**Date**: 2026-07-16

#### Context

The pipeline needs to display progress and log messages, but the output method varies: interactive TUI in terminals, plain text when piped, and potentially programmatic consumers in the future.

#### Decision

Define a `pipeline.Logger` interface with methods `Info`, `OK`, `Warn`, `Step`, `Err`, and `Progress`. The pipeline accepts this interface in `Run()` and never writes to stdout/stderr directly.

#### Consequences

- **Positive**: Pipeline is fully decoupled from output. TUI, plain text, or test mocks can all consume pipeline events.
- **Negative**: Slight overhead from interface method calls (negligible for this use case).

---

### ADR 0005: Bubbletea TUI with auto-detection

**Status**: Accepted  
**Date**: 2026-07-16

#### Context

A 3-stage pipeline running for minutes to hours needs real-time feedback. A static progress bar is insufficient for showing multiple stages, logs, and system info simultaneously.

#### Decision

Use Charm's Bubbletea framework for an interactive TUI dashboard. Auto-detect TTY via `term.IsTerminal()`: if stdout is a terminal, launch TUI; otherwise fall back to plain text. Add `--no-tui` flag for explicit opt-out.

#### Consequences

- **Positive**: Rich, real-time dashboard with stage tracking, progress bars, and scrollable logs. Graceful degradation to plain text.
- **Negative**: Adds ~800KB to binary size (5.0MB total). Requires terminal for full experience.

---

### ADR 0006: Go rewrite from Bash

**Status**: Accepted  
**Date**: 2026-07-16

#### Context

The original Bash script (`bananascaler.sh`) works but lacks structured error handling, progress reporting, and a TUI. Bash limitations make it difficult to add features like parallel processing or programmatic APIs.

#### Decision

Rewrite in Go with idiomatic project layout (`cmd/`, `internal/`). Preserve the same pipeline logic and engineering patterns (atomic output, session isolation, hardware detection).

#### Consequences

- **Positive**: Structured error handling, interfaces for extensibility, Bubbletea TUI, proper signal handling, and a path to parallel processing.
- **Negative**: Requires Go compiler for building. Binary is larger than a script. Bash version retained for reference.
