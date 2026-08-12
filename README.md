# YAML Viewer

YAML Viewer is a cross-platform desktop YAML browser and scalar editor built
with Go and Fyne.

It supports multi-document files, fuzzy search, scalar editing, undo/redo,
save, reload, drag-and-drop, recent files, and native file pickers.

## Commands

Run the application:

```bash
make run
```

Run with diagnostic logs written to `debug.log`:

```bash
make debug
```

Build a production binary for the current platform:

```bash
make build
```

Build the macOS application bundle and DMG:

```bash
make build-macos
```

Build the Windows installer:

```bash
make build-windows
```

Run all tests:

```bash
make test
```

Format the Go source:

```bash
make format
```

Remove generated build artifacts:

```bash
make clean
```

## Build requirements

- All platforms require Go, Make, and `uv`.
- macOS packaging uses the standard macOS command-line tools.
- Windows packaging requires Inno Setup 6 and `rsrc` on `PATH`.
- Linux requires the GTK 3 development libraries used by Fyne.

Production builds omit YAML Viewer's diagnostic logs. Platform packages are
written under `dist/`, and their version information comes from `FyneApp.toml`.

User settings are stored in `~/.config/yaml-viewer/config.yaml`.
