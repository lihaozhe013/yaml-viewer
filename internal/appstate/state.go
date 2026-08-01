// Package appstate owns the mutable application state independently of the UI.
package appstate

import (
	"sync"

	"yamlviewer/internal/fileio"
	"yamlviewer/internal/search"
	"yamlviewer/internal/yamlmodel"
)

// State is safe for background load/search generations to update without
// allowing stale results to replace newer user actions.
type State struct {
	mu sync.RWMutex

	Current  *fileio.LoadedFile
	Selected *yamlmodel.Node
	Query    string
	Results  []search.Result
	Visible  map[string]bool
	Error    error

	generation uint64
	Expanded   map[string]bool
}

func New() *State {
	return &State{Visible: make(map[string]bool), Expanded: make(map[string]bool)}
}

// Begin starts a new async operation and returns its generation token.
func (state *State) Begin() uint64 {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.generation++
	return state.generation
}

// ApplyLoad commits a load only if it belongs to the current generation.
func (state *State) ApplyLoad(generation uint64, loaded *fileio.LoadedFile, err error) bool {
	state.mu.Lock()
	defer state.mu.Unlock()
	if generation != state.generation {
		return false
	}
	state.Error = err
	if err != nil {
		return true
	}
	state.Current = loaded
	state.Selected = nil
	state.Results = nil
	state.Visible = make(map[string]bool)
	return true
}

// ApplySearch commits query results only if they belong to the latest
// generation. It intentionally preserves Selected when it remains present.
func (state *State) ApplySearch(generation uint64, query string, results []search.Result) bool {
	state.mu.Lock()
	defer state.mu.Unlock()
	if generation != state.generation {
		return false
	}
	state.Query = query
	state.Results = append([]search.Result(nil), results...)
	state.Visible = search.VisibleIDs(results)
	return true
}

func (state *State) Generation() uint64 {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.generation
}
