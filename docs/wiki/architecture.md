# Architecture and Decisions (ADRs) — bananascaler

Design decisions for **bananascaler**.

## System Overview

A 4-stage sequential pipeline coordinated by a single Bash script:
1. **Extract** — FFmpeg decodes source video to JPEG frames.
2. **Upscale** — Real-ESRGAN (ncnn-vulkan) applies neural super-resolution per frame.
3. **Re-encode** — FFmpeg re-assembles frames + muxes original audio.
4. **Finalize** — Atomic rename from `.tmp` to final destination.

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
