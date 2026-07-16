# Memory: bananascaler

Persistent notes for agents working on **bananascaler**. Update as decisions are made.

## Known Constraints

- Real-ESRGAN model defaults to `realesr-animevideov3-x2`. Model selection is exposed via `--model` flag but not yet validated against available models.
- Temp dirs created in `/tmp/`. Systems with small `/tmp` partitions may fail on very long videos.
- Audio remux assumes single audio stream (`-map 1:a`). Multi-audio files may need explicit stream selection.
- Bubbletea TUI only renders in terminals. When piped or with `--no-tui`, falls back to plain text via `StdoutLogger`.
- The `--gpu` flag is passed directly to realesrgan-ncnn-vulkan. No validation that the GPU index is actually available on the system.

## Past Decisions

- Chose JPEG for intermediate frames over PNG: ~60-70% disk reduction at negligible quality cost for super-resolution input.
- Used `ncnn-vulkan` backend instead of CUDA-only: broader GPU vendor support via Vulkan.
- Atomic rename pattern adopted from day 1: non-negotiable.
- **v0.2.0**: Rewrote pipeline in Go with `Logger` interface to decouple output from pipeline logic. This enables Bubbletea TUI, plain text, and future programmatic consumers.
- **v0.2.0**: Added Bubbletea TUI with Lipgloss styling. TTY auto-detection: terminal → TUI, piped → plain text.
- **v0.2.0**: `nvidia-smi` probe has 5s timeout to prevent hangs on broken driver states.

## Architecture Notes

- `pipeline.Logger` interface is the integration point for all output. Implement it to consume pipeline events.
- The TUI adapter (`tui.tuiLogger`) converts `Logger` calls into `PipelineEvent` structs sent through a channel to the Bubbletea model.
- Pipeline runs in a goroutine when using TUI; the model's `waitForEvent()` cmd polls the event channel.
- `cmd/root.go` is the decision point: checks `term.IsTerminal()` and `--no-tui` flag to choose between TUI and plain mode.
