// Package history provides bounded document-level undo and redo for committed
// YAML edits.
package history

import "yamlviewer/internal/yamlmodel"

const defaultLimit = 1000

// History stores changes in source order and tracks the last saved position.
// It is intended to be accessed from the UI thread.
type History struct {
	entries    []yamlmodel.ScalarChange
	position   int
	savedAt    int
	limit      int
	savedValid bool
	version    uint64
}

// New returns an empty history with a bounded number of committed changes.
func New(limit int) *History {
	if limit <= 0 {
		limit = defaultLimit
	}
	return &History{limit: limit, savedValid: true}
}

// Reset clears all edits and marks the current document as saved.
func (history *History) Reset() {
	history.entries = nil
	history.position = 0
	history.savedAt = 0
	history.savedValid = true
	history.version++
}

// Commit appends one edit and discards any redo branch.
func (history *History) Commit(change yamlmodel.ScalarChange) {
	if change.Before == change.After {
		return
	}
	if history.position < len(history.entries) {
		if history.savedValid && history.savedAt > history.position {
			history.savedValid = false
		}
		history.entries = history.entries[:history.position]
	}
	history.entries = append(history.entries, change)
	history.position++
	history.version++
	if len(history.entries) > history.limit {
		removed := len(history.entries) - history.limit
		history.entries = history.entries[removed:]
		history.position -= removed
		if history.savedValid {
			history.savedAt -= removed
			if history.savedAt < 0 {
				history.savedValid = false
			}
		}
	}
}

// Undo moves one step backward and returns the change to reverse.
func (history *History) Undo() (yamlmodel.ScalarChange, bool) {
	if history.position == 0 {
		return yamlmodel.ScalarChange{}, false
	}
	history.position--
	history.version++
	return history.entries[history.position], true
}

// Redo moves one step forward and returns the change to reapply.
func (history *History) Redo() (yamlmodel.ScalarChange, bool) {
	if history.position == len(history.entries) {
		return yamlmodel.ScalarChange{}, false
	}
	change := history.entries[history.position]
	history.position++
	history.version++
	return change, true
}

// MarkSaved associates the current history position with the on-disk state.
func (history *History) MarkSaved() {
	history.savedAt = history.position
	history.savedValid = true
}

// Dirty reports whether the current history position differs from the last
// successfully saved position.
func (history *History) Dirty() bool {
	return !history.savedValid || history.position != history.savedAt
}

// Version changes whenever the current in-memory document state changes. It
// lets asynchronous saves distinguish their snapshot from later edits.
func (history *History) Version() uint64 { return history.version }

// CanUndo reports whether a committed change can be reversed.
func (history *History) CanUndo() bool { return history.position > 0 }

// CanRedo reports whether a committed change can be reapplied.
func (history *History) CanRedo() bool { return history.position < len(history.entries) }
