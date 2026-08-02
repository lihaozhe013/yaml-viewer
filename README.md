# YAML Viewer

Read-only desktop YAML browser built with Go, Fyne, and `go.yaml.in/yaml/v3`.
It preserves YAML node order and metadata while supporting multi-document files,
humanized field labels, fuzzy search, node inspection, reload, drag-and-drop,
and recent files.

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

The Open action uses the host operating system's native file picker: Finder on
macOS, the Windows file picker on Windows, and GTK on Linux and other Unix-like
systems. Linux systems need the GTK 3 runtime libraries installed.

Verify the project with:

```bash
go test ./...
go vet ./...
go build .
```
