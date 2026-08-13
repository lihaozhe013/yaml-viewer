package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"

	"yamlviewer/internal/fileio"
	"yamlviewer/internal/search"
	"yamlviewer/internal/yamlmodel"
)

func TestFocusCancelEntryCancelsOnFocusLost(t *testing.T) {
	application := test.NewApp()
	defer application.Quit()

	cancelled := false
	entry := newFocusCancelEntry(func() { cancelled = true })
	window := test.NewWindow(entry)
	defer window.Close()
	window.Resize(fyne.NewSize(400, 200))
	window.Show()

	window.Canvas().Focus(entry)
	if window.Canvas().Focused() != entry {
		t.Fatal("editor did not receive focus")
	}
	window.Canvas().Unfocus()
	if !cancelled {
		t.Fatal("editor did not cancel after losing focus")
	}
}

func TestTappableValueInvokesEditAction(t *testing.T) {
	called := false
	value := newTappableValue(canvas.NewRectangle(nil), func() { called = true })

	value.Tapped(nil)
	if !called {
		t.Fatal("clicking the value did not invoke the edit action")
	}
}

func TestViewerFocusesAndCancelsValueEditor(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	application := test.NewApp()
	defer application.Quit()

	model, err := yamlmodel.Decode([]byte("answer: 42\n"))
	if err != nil {
		t.Fatal(err)
	}
	node := model.Documents[0].Root.Children[0]
	viewer := New(application)
	viewer.current = &fileio.LoadedFile{Model: model, Index: search.NewIndex(model)}
	viewer.state.Selected = node
	viewer.updateInspector(node)
	viewer.beginEditValue()

	if viewer.editingNode != node.ID {
		t.Fatal("value editor did not open")
	}
	if viewer.window.Canvas().Focused() != viewer.valueEditor {
		t.Fatal("value editor did not receive focus")
	}
	viewer.window.Canvas().Unfocus()
	if viewer.editingNode != "" {
		t.Fatal("value editor did not cancel after losing focus")
	}
}

func TestEditorActionButtonDoesNotCancelOnClick(t *testing.T) {
	application := test.NewApp()
	defer application.Quit()

	cancelled := false
	entry := newFocusCancelEntry(func() { cancelled = true })
	called := false
	button := newEditorActionButton("Apply", entry.preserveNextFocusLoss, func() { called = true })
	window := test.NewWindow(container.NewVBox(entry, button))
	defer window.Close()
	window.Show()
	window.Canvas().Focus(entry)

	button.Tapped(nil)
	if !called {
		t.Fatal("editor action button did not invoke its action")
	}
	if cancelled {
		t.Fatal("editor action button incorrectly cancelled editing")
	}
}

func TestEditMenuCommandsDispatchToFocusedEntry(t *testing.T) {
	application := test.NewApp()
	defer application.Quit()

	viewer := New(application)
	entry := viewer.searchEntry
	entry.SetText("")
	viewer.window.Canvas().Focus(entry)
	entry.TypedRune('a')
	entry.TypedRune('b')
	entry.TypedRune('c')
	entry.TypedRune('d')

	viewer.undoItem.Action()
	if entry.Text != "" {
		t.Fatalf("focused entry after undo = %q, want empty", entry.Text)
	}
	viewer.redoItem.Action()
	if entry.Text != "abcd" {
		t.Fatalf("focused entry after redo = %q, want %q", entry.Text, "abcd")
	}

	viewer.selectAllItem.Action()
	viewer.copyItem.Action()
	if got := application.Clipboard().Content(); got != "abcd" {
		t.Fatalf("clipboard content = %q, want %q", got, "abcd")
	}
	viewer.cutItem.Action()
	if entry.Text != "" {
		t.Fatalf("focused entry after cut = %q, want empty", entry.Text)
	}
	viewer.pasteItem.Action()
	if entry.Text != "abcd" {
		t.Fatalf("focused entry after paste = %q, want %q", entry.Text, "abcd")
	}
}
