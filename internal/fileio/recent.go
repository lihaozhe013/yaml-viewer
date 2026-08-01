package fileio

import (
	"path/filepath"
	"sync"
)

// RecentFiles is a small in-memory MRU list. Persistence can be added at the
// application boundary without coupling the domain model to Fyne preferences.
type RecentFiles struct {
	mu    sync.Mutex
	paths []string
	limit int
}

func NewRecentFiles(limit int) *RecentFiles {
	if limit <= 0 {
		limit = 10
	}
	return &RecentFiles{limit: limit}
}

func (recent *RecentFiles) Add(path string) {
	if recent == nil {
		return
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return
	}
	recent.mu.Lock()
	defer recent.mu.Unlock()
	filtered := recent.paths[:0]
	for _, existing := range recent.paths {
		if existing != path {
			filtered = append(filtered, existing)
		}
	}
	recent.paths = append([]string{path}, filtered...)
	if len(recent.paths) > recent.limit {
		recent.paths = recent.paths[:recent.limit]
	}
}

func (recent *RecentFiles) List() []string {
	if recent == nil {
		return nil
	}
	recent.mu.Lock()
	defer recent.mu.Unlock()
	return append([]string(nil), recent.paths...)
}
