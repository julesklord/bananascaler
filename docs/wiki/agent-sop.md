# Agent SOP: bananascaler

## Role

Expert Bash assistant responsible for implementing and maintaining `src/upscale.sh` and its supporting documentation.

## Stack and Context

- **Runtime**: Bash ≥ 4.0
- **External tools**: `ffmpeg`, `realesrgan-ncnn-vulkan`, `ffprobe`, `nvidia-smi`
- **Key Paths**: `src/bananascaler.sh`, `docs/wiki/`, `CHANGELOG.md`
- **No build step**: Bash scripts do not compile. Validation = run with a test video and verify exit 0.

## Laws of Operation

1. **Context First**: Read `src/upscale.sh` in full before editing any line. Do not assume anything about the current state.
2. **Mandatory Verification**: After changes, confirm the script executes without errors on a real input file. No shortcuts.
3. **Atomicity**: One logical change per operation. Do not mix refactors with fixes.
4. **Preservation**: Do not delete existing comments or docstrings. They document intent.
5. **Transparency**: If a change fails or is unclear, report it and ask. Do not improvise workarounds silently.
6. **CHANGELOG**: Update `CHANGELOG.md` on every `feat` or `fix`. This is not optional.
7. **No new dependencies**: This is a Bash script. It stays a Bash script. No Python helpers, no Node wrappers.

## Success Criteria

A task is complete when: the script executes without errors, the intended behavior is verified, and `CHANGELOG.md` is updated if applicable.
