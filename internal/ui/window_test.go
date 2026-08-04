package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"

	appconfig "yamlviewer/internal/config"
	"yamlviewer/internal/fileio"
	"yamlviewer/internal/search"
	"yamlviewer/internal/yamlmodel"
)

func TestBuildMenusAndDefaultViewMode(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	application := test.NewApp()
	defer application.Quit()
	viewer := New(application)

	if viewer.viewMode != ViewModeSpacious {
		t.Fatalf("default view mode = %q, want spacious", viewer.viewMode)
	}
	if viewer.searchMode != search.ModeSmartFuzzy {
		t.Fatalf("default search mode = %q, want smart fuzzy", viewer.searchMode)
	}
	if viewer.themeMode != appconfig.ThemeModeLight {
		t.Fatalf("default theme mode = %q, want light", viewer.themeMode)
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
	if viewer.searchSettingsButton.Text != "Search: Smart Fuzzy" {
		t.Fatalf("search settings button = %q", viewer.searchSettingsButton.Text)
	}
	if viewer.expandCollapseButton.Text != "Expand" {
		t.Fatalf("expand/collapse button = %q, want Expand", viewer.expandCollapseButton.Text)
	}
	if !viewer.expandCollapseButton.Disabled() {
		t.Fatal("expand/collapse button should be disabled without a selected branch")
	}
	if len(viewer.viewMenu.Items) != 5 || viewer.viewMenu.Items[3] != viewer.themeItem || viewer.viewMenu.Items[4] != viewer.searchSettingsItem {
		t.Fatal("view menu should expose the theme toggle before search settings")
	}
	if viewer.themeItem.Checked {
		t.Fatal("light mode should leave the dark mode toggle unchecked")
	}

	viewer.themeItem.Action()
	if viewer.themeMode != appconfig.ThemeModeDark || !viewer.themeItem.Checked {
		t.Fatal("theme toggle should switch to dark mode and check the menu item")
	}
	if viewer.config.ThemeMode != appconfig.ThemeModeDark {
		t.Fatalf("saved theme mode = %q, want %q", viewer.config.ThemeMode, appconfig.ThemeModeDark)
	}
	selectedTheme, ok := viewer.app.Settings().Theme().(fixedVariantTheme)
	if !ok || selectedTheme.variant != theme.VariantDark {
		t.Fatalf("selected theme = %#v, want a fixed dark variant", viewer.app.Settings().Theme())
	}
	viewer.themeItem.Action()
	if viewer.themeMode != appconfig.ThemeModeLight || viewer.themeItem.Checked {
		t.Fatal("theme toggle should switch back to light mode and uncheck the menu item")
	}
}

func TestSelectedBranchToggleExpandsAndCollapsesAllDescendants(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	application := test.NewApp()
	defer application.Quit()

	model, err := yamlmodel.Decode([]byte("root:\n  branch:\n    leaf: value\n    deep:\n      end: value\n  sibling: value\n"))
	if err != nil {
		t.Fatal(err)
	}

	viewer := New(application)
	viewer.current = &fileio.LoadedFile{Model: model, Index: search.NewIndex(model)}
	viewer.refreshTree()
	selected := model.Documents[0].Root.Children[0]
	viewer.selectTreeItem(selected.ID)

	if viewer.expandCollapseButton.Text != "Expand" {
		t.Fatalf("initial button = %q, want Expand", viewer.expandCollapseButton.Text)
	}
	if viewer.expandCollapseButton.Disabled() {
		t.Fatal("expand/collapse button should be enabled for a selected branch")
	}

	viewer.expandCollapseButton.Tapped(nil)
	if viewer.expandCollapseButton.Text != "Collapse" {
		t.Fatalf("expanded button = %q, want Collapse", viewer.expandCollapseButton.Text)
	}
	assertBranchOpen(t, viewer, selected)

	viewer.expandCollapseButton.Tapped(nil)
	if viewer.expandCollapseButton.Text != "Expand" {
		t.Fatalf("collapsed button = %q, want Expand", viewer.expandCollapseButton.Text)
	}
	assertBranchClosed(t, viewer, selected)
}

func assertBranchOpen(t *testing.T, viewer *Viewer, node *yamlmodel.Node) {
	t.Helper()
	if len(node.Children) > 0 && !viewer.tree.IsBranchOpen(node.ID) {
		t.Errorf("branch %s is closed", node.ID)
	}
	for _, child := range node.Children {
		assertBranchOpen(t, viewer, child)
	}
}

func assertBranchClosed(t *testing.T, viewer *Viewer, node *yamlmodel.Node) {
	t.Helper()
	if len(node.Children) > 0 && viewer.tree.IsBranchOpen(node.ID) {
		t.Errorf("branch %s is open", node.ID)
	}
	for _, child := range node.Children {
		assertBranchClosed(t, viewer, child)
	}
}
