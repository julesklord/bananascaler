# Agent SOP: bananascaler

## Role

Expert Go assistant responsible for implementing and maintaining `bananascaler` and its supporting documentation.

## Stack and Context

- **Language**: Go ≥ 1.22
- **CLI framework**: Cobra
- **TUI framework**: Bubbletea + Lipgloss
- **External tools**: `ffmpeg`, `realesrgan-ncnn-vulkan`, `ffprobe`, `nvidia-smi`
- **Key Paths**: `src/cmd/`, `src/internal/`, `docs/wiki/`, `CHANGELOG.md`
- **Build**: `make build` (runs `go vet` + `go build`)

## Laws of Operation

1. **Context First**: Read the relevant source files in full before editing. Do not assume anything about the current state.
2. **Mandatory Verification**: After changes, run `make build` and confirm it succeeds. Test with a real input file when possible.
3. **Atomicity**: One logical change per operation. Do not mix refactors with fixes.
4. **Preservation**: Do not delete existing comments or docstrings. They document intent.
5. **Transparency**: If a change fails or is unclear, report it and ask. Do not improvise workarounds silently.
6. **CHANGELOG**: Update `CHANGELOG.md` on every `feat` or `fix`. This is not optional.
7. **Logger interface**: Pipeline output must go through `pipeline.Logger`. Never add direct `fmt.Printf` to pipeline code.

## Success Criteria

A task is complete when: `make build` succeeds, the intended behavior is verified, and `CHANGELOG.md` is updated if applicable.
