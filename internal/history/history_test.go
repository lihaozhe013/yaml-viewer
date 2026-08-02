package history

import (
	"testing"

	"yamlviewer/internal/yamlmodel"
)

func change(before, after string) yamlmodel.ScalarChange {
	return yamlmodel.ScalarChange{
		NodeID: "node",
		Before: yamlmodel.ScalarState{Value: before},
		After:  yamlmodel.ScalarState{Value: after},
	}
}

func TestHistoryUndoRedoAndSavedPosition(t *testing.T) {
	history := New(10)
	if history.Dirty() {
		t.Fatal("new history should be clean")
	}
	history.Commit(change("one", "two"))
	history.Commit(change("two", "three"))
	if !history.Dirty() || !history.CanUndo() {
		t.Fatal("committed history should be dirty and undoable")
	}
	if history.CanRedo() {
		t.Fatal("new commits should not create a redo entry")
	}
	history.MarkSaved()
	if history.Dirty() {
		t.Fatal("MarkSaved() should clear dirty state")
	}
	if _, ok := history.Undo(); !ok || !history.Dirty() {
		t.Fatal("Undo() should make a saved document dirty")
	}
	if _, ok := history.Redo(); !ok || history.Dirty() {
		t.Fatal("Redo() should return to the saved state")
	}
}

func TestHistoryNewCommitClearsRedoBranch(t *testing.T) {
	history := New(10)
	history.Commit(change("one", "two"))
	history.Commit(change("two", "three"))
	if _, ok := history.Undo(); !ok {
		t.Fatal("Undo() failed")
	}
	history.Commit(change("two", "four"))
	if history.CanRedo() {
		t.Fatal("new commit should clear redo branch")
	}
}

func TestHistoryLimitInvalidatesDroppedSavedPosition(t *testing.T) {
	history := New(1)
	history.Commit(change("one", "two"))
	history.Commit(change("two", "three"))
	if !history.Dirty() {
		t.Fatal("history should be dirty after a new commit")
	}
	if _, ok := history.Undo(); !ok {
		t.Fatal("Undo() failed")
	}
	if !history.Dirty() {
		t.Fatal("dropped saved position should remain dirty")
	}
}
