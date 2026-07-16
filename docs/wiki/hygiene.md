# Hygiene and Git Workflow

This project follows the **FMG Development Bible**. Deviations are not acceptable.

## Atomic Commits

**Conventional Commits** format is mandatory:

```
<type>(<scope>): <subject>

<body: technical justification>

<footer: Fixes #123>
```

### Allowed Types

- `feat`: New functionality.
- `fix`: Bug correction.
- `docs`: Documentation-only changes.
- `style`: Formatting, no logic change.
- `refactor`: Code restructuring that adds nothing and fixes nothing.
- `chore`: Maintenance tasks, dependency updates.

## Branch Workflow

- `main`: Production branch. Linear history only.
- `feat/*`: New features.
- `fix/*`: Bug corrections.

**Prohibited**: `git push --force` to `main`. Always.

## Release Process

1. Update `VERSION` file.
2. Update `CHANGELOG.md` — move `[Unreleased]` entries under new version header.
3. Commit: `chore(release): bump version to X.Y.Z`.
4. Tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`.
5. Push tag: `git push origin vX.Y.Z`.
