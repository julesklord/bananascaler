# Agent SOP (Standard Operating Procedure) — bananascaler

This file is the entry point for any AI agent (Gemini, Claude, GPT, etc.) working on this repository. Read it before touching anything.

## General Instructions

1. **Familiarization**: Read `docs/wiki/index.md` and `FMG-REPO-BIBLE.md` before making any changes.
2. **Compliance**: Follow the laws defined in `docs/wiki/agent-sop.md`. No exceptions.
3. **Identity**: Read `docs/SOUL.md` to understand tone and principles.

## Agent Initialization

This project is a Bash tool with no build step. Validation means running the script with a test video and verifying exit code 0.

```bash
# Verify dependencies are present
command -v ffmpeg && command -v realesrgan-ncnn-vulkan
# Run the script (requires a real input file)
bash src/bananascaler.sh <test_input.mp4>
```

## Key Paths

- `src/bananascaler.sh` — The only source file. Read it before editing.
- `docs/wiki/` — Architecture decisions and development notes.
- `CHANGELOG.md` — Update this on every `feat` or `fix`.
