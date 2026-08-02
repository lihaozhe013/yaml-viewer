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
`~/.config/yaml-viewer/config.yaml` and is offered in the Recent files menu on
the next launch. It is not opened automatically when no command-line path is
provided.

Verify the project with:

```bash
go test ./...
go vet ./...
go build .·
```
