# YAML Viewer Agent Guide

This document describes the project structure, runtime flow, and configuration
rules for future contributors and coding agents.

## Project Overview

YAML Viewer is a cross-platform desktop YAML browser and scalar editor built
with Go, Fyne, and `go.yaml.in/yaml/v3`.

The application supports:

- YAML hierarchy browsing with preserved node order and source metadata.
- Multi-document YAML files.
- Search across keys, display labels, paths, values, tags, anchors, aliases,
  comments, and document names.
- Smart fuzzy search and order-independent keyword search.
- Scalar editing with YAML literal parsing.
- Undo and redo history.
- Save, Save As, reload, drag-and-drop, recent files, and native file pickers.
- A bundled blue application icon for runtime use and desktop packaging.

The application entrypoint is the root `main.go`. 

## Architecture

The code is organized by responsibility under `internal/`:

- `main.go` creates the Fyne application, installs the icon, chooses command
  line versus remembered-file startup, and starts the window event loop.
- `internal/ui` owns the `Viewer` controller and all Fyne presentation. It
  builds the toolbar, menus, hierarchy tree, inspector, dialogs, editing
  actions, and asynchronous UI callbacks.
- `internal/config` owns persistent application configuration, the embedded
  configuration template, default merging, validation, and atomic writes.
- `internal/yamlmodel` decodes YAML into the application tree model. Nodes keep
  paths, source locations, comments, tags, anchors, aliases, styles, and the
  source representation needed for editing.
- `internal/fileio` loads YAML files, builds search indexes, tracks recent
  files, and writes YAML atomically.
- `internal/filepicker` abstracts native open and save dialogs for macOS,
  Windows, Linux, and other supported desktop environments.
- `internal/search` creates normalized searchable projections and ranks search
  results. `ModeSmartFuzzy` preserves forgiving matching; `ModeKeyword`
  requires all complete normalized keywords while ignoring their order.
- `internal/appstate` owns generation-aware state transitions so stale async
  loads or searches cannot replace newer user actions.
- `internal/display` contains presentation-independent formatting such as key
  humanization and word splitting.
- `internal/history` manages scalar edit history for undo and redo.
- `internal/assets` embeds application visual resources such as the SVG icon.

### Runtime data flow

1. `main.go` creates the Fyne app and calls `ui.New`.
2. `ui.New` calls `config.Load`, initializes the viewer state, and builds the
   window and menus.
3. A command line path takes precedence. Without one, `OpenLastPath` opens the
   path stored in config, if any. With no path available, the viewer starts
   empty.
4. File loading and search indexing happen off the UI thread. Results return
   through `fyne.Do` and are accepted only when their `appstate` generation is
   still current.
5. Search input is debounced. The current search mode is captured before a
   background search starts so mode changes cannot race with the search worker.
6. Scalar edits update the YAML model, commit to history, rebuild the search
   index, refresh the tree and inspector, and mark the document dirty.
7. On normal window close, `Viewer.saveConfig` writes the current file path and
   search mode to the user config.

Keep filesystem work and expensive parsing off the Fyne UI thread. Keep widget
reads and mutations on the UI thread and marshal background results with
`fyne.Do`.

## Configuration Rules

### Files and source of defaults

The user configuration is stored at:

```text
~/.config/yaml-viewer/config.yaml
```

The default template is:

```text
internal/config/default.yaml
```

The template is embedded into the binary with `go:embed`. It is not read from
the project directory at runtime. Changing the template therefore requires a
new build before users receive the new defaults.

The template is the source of truth for normal default values. Hardcoded
fallbacks in Go may exist only as safety fallbacks for a missing or malformed
template value; they must not replace a valid template value.

### Startup behavior

`config.Load` must be called during application startup before the viewer is
built. It performs these steps:

1. Decode the embedded template.
2. Resolve the user config path.
3. If the user config does not exist or is empty, create it immediately from
   the template using the same atomic writer as normal saves.
4. If the user config exists, merge it over the template.
5. Add template fields that are missing from the user config.
6. Validate known fields and replace invalid values with the corresponding
   template value.
7. Persist the merged result immediately when fields were added or repaired.
8. Return the typed `Config` value to the UI layer.

This means an upgraded application repairs and expands an old config during
startup, before the rest of the application consumes it.

### Precedence and merge semantics

The precedence order is:

```text
valid user value > template value > hardcoded safety fallback
```

The template does not overwrite an existing valid user value. It only supplies
missing values and repairs invalid known values. Map values are merged
recursively. Unknown user fields are retained so future or user-defined
settings are not silently deleted.

For example, if a future template adds:

```yaml
search_mode: smart_fuzzy
theme: light
```

and an existing user config contains:

```yaml
search_mode: keyword
custom_setting: true
```

the merged config must retain `search_mode: keyword` and
`custom_setting: true`, while adding `theme: light`.

Known fields with an invalid type or unsupported value are replaced by the
template value and written back. A syntactically unparseable user YAML file is
not overwritten automatically. The application uses template defaults in
memory, marks the source as invalid, logs the load failure, and refuses to
overwrite that source during shutdown.

### Shutdown behavior

Configuration changes are applied to the in-memory `Config` immediately, but
normal persistence happens when the application closes:

- `LastOpenedFile` is updated when a file opens or when Save As changes the
  current path.
- `SearchMode` is updated when the user changes the search settings.
- `Viewer.saveConfig` writes both values on normal window close.

Do not add ad-hoc config writes from individual widgets. If a new setting is
changed in the UI, update the in-memory config and let the centralized close
handler persist it, unless the feature explicitly requires crash-resistant
immediate persistence.

### Adding a new config field

When adding a persistent setting:

1. Add its default value to `internal/config/default.yaml`.
2. Add a typed field to `internal/config.Config` with a stable snake_case YAML
   tag.
3. Ensure the field is read into runtime state during `ui.New` or the relevant
   feature initialization.
4. Ensure the current runtime value is copied back to `viewer.config` before
   shutdown persistence.
5. If the field has an enum or other constraint, extend
   `normalizeKnownFields` so invalid values fall back to the template value.
6. Keep `MarshalYAML` synchronized with the typed fields so current values are
   written while unknown raw fields remain preserved.
7. Add tests for template defaults, old configs missing the field, user value
   precedence, invalid values, and preservation of unknown fields.

Never use `omitempty` for a setting that must be materialized into upgraded
configs. The saved file should contain the complete set of known template
fields.

The config writer creates the parent directory with mode `0700`, writes a
temporary file with mode `0600`, calls `Sync`, closes it, and atomically renames
it into place. Preserve this behavior for all config writes.

## UI and Feature Guidelines

- Keep user-facing UI text in the existing English style.
- Keep search mode state synchronized between the search toolbar button, the
  Search Settings dialog, the View menu, and `viewer.config`.
- A new search behavior belongs in `internal/search`, not directly in Fyne
  callbacks. Add a search mode or matcher there, then pass it from `Viewer`.
- Preserve stable result ordering and ancestor visibility when changing search
  ranking or matching rules.
- Use the existing native file picker abstraction instead of introducing a
  platform-specific dialog in the UI package.
- Keep the current Spacious View behavior intact; Compact View is reserved and
  currently disabled.
- Preserve user edits in unrelated files and avoid broad refactors when adding
  a feature.

## Change Conventions

- Comments, documentation, and commit messages must be written in English.
- Commit messages must follow Conventional Commits, such as `feat:`, `fix:`,
  `refactor:`, `test:`, `docs:`, or `chore:`.
- Use `rg` for repository searches.
- Do not commit generated binaries or local user config files.
- Preserve atomic writes and avoid destructive filesystem or Git commands.
