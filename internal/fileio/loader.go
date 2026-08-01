// Package fileio provides filesystem operations used by the application.
package fileio

import (
	"fmt"
	"os"
	"path/filepath"

	"yamlviewer/internal/search"
	"yamlviewer/internal/yamlmodel"
)

// LoadedFile contains the source path and all derived read-only data needed by
// the UI. Index construction is intentionally part of loading so it can run
// off the UI thread.
type LoadedFile struct {
	Path  string
	Name  string
	Model *yamlmodel.File
	Index *search.Index
}

// Load reads, decodes, and indexes a YAML file.
func Load(path string) (*LoadedFile, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	model, err := yamlmodel.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &LoadedFile{Path: path, Name: filepath.Base(path), Model: model, Index: search.NewIndex(model)}, nil
}
