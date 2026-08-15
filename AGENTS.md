# Agent Guide

## Project

YAML Viewer is a cross-platform Go desktop application built with Fyne and
`go.yaml.in/yaml/v3`. The entrypoint is `main.go`; feature code lives under
`internal/`, grouped by responsibility (`ui`, `config`, `yamlmodel`, `fileio`,
`filepicker`, `search`, `appstate`, and supporting packages).

Read the relevant implementation and tests before changing behavior. Keep
filesystem and expensive parsing work off the Fyne UI thread, and marshal
background results back to the UI with `fyne.Do`. Preserve YAML node order,
source metadata, multi-document support, stable search result ordering, atomic
writes, and user changes outside the requested scope.

## Required conventions

- Write all new or modified comments, documentation, commit messages, and PR
  descriptions in English. Do not add Chinese text to project artifacts.
- Use a complete Conventional Commit message for every commit: include a
  valid type and a specific imperative subject, adding a body when context is
  useful. For example: `fix: write debug diagnostics to a file`.
- Run the project formatter before every commit. For this repository, run
  `make format`, then run the relevant tests. Do not commit generated binaries,
  packages, `debug.log`, or local user configuration files.
- If a file exceeds 1,000 lines, assess cohesion, coupling, and whether it
  should be split before adding more code. Record the decision in the change
  summary when the file remains large.
- Keep changes cross-platform. Do not depend on Windows-only tools or shell
  behavior. When platform-specific automation is needed, prefer a Python
  script that uses portable standard-library behavior. Minimize environment
  variables and avoid hardcoded platform executable names.
- Use `rg` for repository searches, `fd` when available for file discovery,
  `uv run` for Python commands, and `vim` as the default terminal editor.
- Do not run destructive Git or filesystem commands without explicit approval.
  Preserve unrelated worktree changes.

## Debug logging

Debug builds write application diagnostics to `debug.log` in the working
directory and must not use the terminal as the default destination. Prefix
diagnostics with `[feature_name]` so feature flows can be filtered. Keep normal
builds quiet. A useful debugging command is:

```bash
make debug
```

Then inspect the focused output with:

```bash
rg '\[feature_name\]' debug.log > feature-debug.log
```

## Documentation

Treat persistent project documentation primarily as concise agent reference.
Put new reference material under `docs/reference/`; keep it short and focused
on architecture, invariants, workflows, and verification commands. Avoid
duplicating source code or writing narrative details that agents can derive
from the implementation and tests.

## Verification

Before handing off a change, run the narrowest relevant tests and formatting.
For a full validation pass, use:

```bash
make format
make test
```