package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestBuildMenusAndDefaultViewMode(t *testing.T) {
	application := test.NewApp()
	defer application.Quit()
	viewer := New(application)

	if viewer.viewMode != ViewModeSpacious {
		t.Fatalf("default view mode = %q, want spacious", viewer.viewMode)
	}
	if len(viewer.mainMenu.Items) != 4 {
		t.Fatalf("main menu contains %d menus, want 4", len(viewer.mainMenu.Items))
	}
	for index, want := range []string{"File", "Edit", "View", "About"} {
		if viewer.mainMenu.Items[index].Label != want {
			t.Errorf("menu %d = %q, want %q", index, viewer.mainMenu.Items[index].Label, want)
		}
	}
	if !viewer.saveItem.Disabled || !viewer.saveAsItem.Disabled || !viewer.editValueItem.Disabled {
		t.Fatal("empty viewer should disable save and edit commands")
	}
	if !viewer.spaciousItem.Checked {
		t.Fatal("spacious view should be checked by default")
	}
	if !viewer.compactItem.Disabled {
		t.Fatal("compact view should be reserved and disabled")
	}
}
