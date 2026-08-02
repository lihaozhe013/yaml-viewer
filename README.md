# YAML Viewer

Desktop YAML browser and scalar editor built with Go, Fyne, and
`go.yaml.in/yaml/v3`.
It preserves YAML node order and metadata while supporting multi-document files,
humanized field labels, fuzzy search, node inspection, scalar editing,
undo/redo, save, reload, drag-and-drop, and recent files.

Run it with:

```bash
go run . path/to/file.yaml
```

To start with an empty viewer and choose a file from the UI:

```bash
go run .
```

The last successfully opened file is remembered in
`~/.config/yaml-viewer/config.yaml` and automatically reopened on the next
launch when no command-line path is provided. A command-line path takes
precedence; on the first launch, the viewer starts empty.

Search defaults to **Smart Fuzzy** matching. Use the search button or **View →
Search Settings** to enable **Keyword Match**, which requires every keyword but
ignores keyword order. For example, `player speed attack` and
`player attack speed` can both match `player.attack_speed`. Search mode is
stored in the config and missing fields are filled from the embedded template
at `internal/config/default.yaml` when the application starts.

Select a scalar node and choose **Edit Value** to edit one YAML scalar literal.
The value is parsed as YAML, so quotes, numbers, booleans, and `null` retain
their YAML meaning. Use **Save** to write the current file or **Save As** to
choose another path. Unsaved changes are confirmed before opening another file,
reloading, or closing the window.

The current inspector layout is **Spacious View**. **Compact View** is reserved
in the View menu for a future layout.

The application uses a bundled blue YAML document icon at runtime. The SVG
source is in `internal/assets/yaml-viewer.svg`; `Icon.png` and `FyneApp.toml`
are included for Fyne desktop packaging on macOS, Windows, and Linux.

The Open action uses the host operating system's native file picker: Finder on
macOS, the Windows file picker on Windows, and GTK on Linux and other Unix-like
systems. Linux systems need the GTK 3 runtime libraries installed.

Verify the project with:

```bash
go test ./...
go vet ./...
go build .
```
