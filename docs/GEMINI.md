# Gemini CLI Rules

Specific instructions for the **Gemini CLI / Antigravity** agent. Pay attention.

## Working Context

- This project follows the FMG Development Standard. Do not deviate.
- The project is written in Go with a Bubbletea TUI. Build with `make build`.
- Source files are in `src/cmd/`, `src/internal/`, and `src/main.go`.
- Use `grep_search` to navigate before editing. Read before writing.
- Check `git status` before every commit.

## Workflow

Research → Strategy → Execution (Plan-Act-Validate). Do not skip steps.

## Restrictions

- Do not modify files in `docs/` unless documenting a real feature or decision.
- Keep `CHANGELOG.md` updated with every `feat` or `fix`. No exceptions.
- Pipeline output must go through `pipeline.Logger`. Never add direct `fmt.Printf` to pipeline code.
- Never force-push to `main`.
