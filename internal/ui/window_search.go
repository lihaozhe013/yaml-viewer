package ui

import (
	"strings"
	"time"

	"fyne.io/fyne/v2"

	"yamlviewer/internal/search"
)

const searchDebounce = 125 * time.Millisecond

func (viewer *Viewer) scheduleSearch() {
	viewer.searchMu.Lock()
	if viewer.searchTimer != nil {
		viewer.searchTimer.Stop()
	}
	viewer.searchTimer = time.AfterFunc(searchDebounce, viewer.runSearch)
	viewer.searchMu.Unlock()
}

func (viewer *Viewer) scheduleSearchNow() {
	viewer.searchMu.Lock()
	if viewer.searchTimer != nil {
		viewer.searchTimer.Stop()
	}
	viewer.searchMu.Unlock()
	viewer.runSearch()
}

func (viewer *Viewer) runSearch() {
	// Timer callbacks are not Fyne UI callbacks. Marshal the widget reads and
	// generation update back to the UI thread before launching the search work.
	fyne.Do(viewer.beginSearch)
}

func (viewer *Viewer) beginSearch() {
	query := viewer.searchEntry.Text
	loaded := viewer.current
	if loaded == nil {
		return
	}
	generation := viewer.state.Begin()
	mode := viewer.searchMode
	go func() {
		results := loaded.Index.Search(query, mode)
		fyne.Do(func() {
			if !viewer.state.ApplySearch(generation, query, results) {
				return
			}
			viewer.results = results
			viewer.visible = search.VisibleIDs(results)
			viewer.refreshTree()
			if len(results) > 0 && strings.TrimSpace(query) != "" {
				viewer.tree.Select(results[0].Node.ID)
			}
		})
	}()
}
