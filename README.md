# YAML Viewer

Read-only desktop YAML browser built with Go, Fyne, and `go.yaml.in/yaml/v3`.
It preserves YAML node order and metadata while supporting multi-document files,
humanized field labels, fuzzy search, node inspection, reload, drag-and-drop,
and recent files.

Run it with:

```bash
go run ./cmd/yamlviewer path/to/file.yaml
```

Verify the project with:

```bash
go test ./...
go vet ./...
go build ./cmd/yamlviewer
```
