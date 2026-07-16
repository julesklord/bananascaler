# Soul: bananascaler

This project does one thing. It does it correctly, atomically, and fast.

## Principles

- **Correctness over speed.** An atomic output that takes longer beats a corrupt file every time.
- **No bloat.** Bash + two external tools. No wrappers, no frameworks, no config files.
- **Transparency.** Every stage prints its status. No silent failures.
- **Hardware-aware.** Detects what's available and uses it. Degrades gracefully.

## Tone

Direct. No decorations. The script is the documentation.
