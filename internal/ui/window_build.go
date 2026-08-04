package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

func (viewer *Viewer) build() {
	viewer.window = viewer.app.NewWindow("YAML Viewer")
	viewer.window.Resize(fyne.NewSize(1100, 720))

	viewer.searchEntry = widget.NewEntry()
	viewer.searchEntry.SetPlaceHolder("Search fields, paths, values, tags, comments…")
	viewer.searchEntry.OnChanged = func(_ string) {
		viewer.scheduleSearch()
	}
	viewer.searchSettingsButton = widget.NewButton(viewer.searchModeLabel(), viewer.showSearchSettings)

	openButton := widget.NewButton("Open", viewer.openDialog)
	reloadButton := widget.NewButton("Reload", viewer.reload)
	viewer.expandCollapseButton = widget.NewButton("Expand", viewer.toggleSelectedBranch)
	viewer.recentSelect = widget.NewSelect(nil, func(path string) {
		if path != "" {
			viewer.requestOpenPath(path)
		}
	})
	viewer.recentSelect.PlaceHolder = "Recent files"
	viewer.recentSelect.SetOptions(viewer.recent.List())
	clearButton := widget.NewButton("Clear", func() { viewer.searchEntry.SetText("") })
	searchActions := container.NewHBox(viewer.searchSettingsButton, clearButton)
	toolbar := container.NewBorder(nil, nil, container.NewHBox(openButton, reloadButton, viewer.expandCollapseButton), searchActions,
		container.NewBorder(nil, nil, nil, viewer.recentSelect, viewer.searchEntry))

	viewer.tree = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID { return viewer.childrenOf(id) },
		func(id widget.TreeNodeID) bool { return len(viewer.childrenOf(id)) > 0 },
		func(_ bool) fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TreeNodeID, _ bool, object fyne.CanvasObject) {
			object.(*widget.Label).SetText(viewer.items[id].label)
		},
	)
	viewer.tree.HideSeparators = true
	viewer.tree.OnSelected = func(id widget.TreeNodeID) { viewer.selectTreeItem(id) }
	viewer.tree.OnBranchOpened = func(id widget.TreeNodeID) {
		if !viewer.programmatic {
			viewer.state.Expanded[id] = true
			viewer.updateExpandCollapseButton()
		}
	}
	viewer.tree.OnBranchClosed = func(id widget.TreeNodeID) {
		if !viewer.programmatic {
			viewer.state.Expanded[id] = false
			viewer.updateExpandCollapseButton()
		}
	}

	viewer.inspector = container.NewVBox()
	inspectorScroll := container.NewVScroll(viewer.inspector)
	left := container.NewBorder(widget.NewLabelWithStyle("YAML hierarchy", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), nil, nil, nil, container.NewVScroll(viewer.tree))
	right := container.NewBorder(widget.NewLabelWithStyle("Node inspector", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), nil, nil, nil, inspectorScroll)
	split := container.NewHSplit(left, right)
	split.SetOffset(0.42)

	viewer.status = widget.NewLabel("No file open")
	viewer.errorLabel = widget.NewLabel("")
	viewer.errorLabel.Wrapping = fyne.TextWrapWord
	footer := container.NewBorder(nil, nil, viewer.errorLabel, nil, viewer.status)
	viewer.window.SetContent(container.NewBorder(toolbar, footer, nil, nil, split))
	viewer.buildMenus()
	viewer.window.SetCloseIntercept(viewer.requestClose)
	viewer.window.SetOnClosed(viewer.saveConfig)
	viewer.window.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		for _, uri := range uris {
			if uri.Scheme() == "file" {
				viewer.requestOpenPath(uri.Path())
				break
			}
		}
	})
	viewer.registerShortcuts()
	viewer.refreshTree()
	viewer.updateInspector(nil)
	viewer.updateCommands()
}

func (viewer *Viewer) buildMenus() {
	modifier := fyne.KeyModifierShortcutDefault
	shortcut := func(key fyne.KeyName, extra fyne.KeyModifier) *desktop.CustomShortcut {
		return &desktop.CustomShortcut{KeyName: key, Modifier: modifier | extra}
	}

	openItem := fyne.NewMenuItem("Open", viewer.openDialog)
	openItem.Shortcut = shortcut(fyne.KeyO, 0)
	viewer.saveItem = fyne.NewMenuItem("Save", viewer.save)
	viewer.saveItem.Shortcut = shortcut(fyne.KeyS, 0)
	viewer.saveAsItem = fyne.NewMenuItem("Save As", viewer.saveAs)
	viewer.saveAsItem.Shortcut = shortcut(fyne.KeyS, fyne.KeyModifierShift)
	viewer.reloadItem = fyne.NewMenuItem("Reload", viewer.reload)
	viewer.reloadItem.Shortcut = shortcut(fyne.KeyR, 0)
	quitItem := fyne.NewMenuItem("Quit", viewer.requestClose)
	viewer.fileMenu = fyne.NewMenu("File", openItem, viewer.saveItem, viewer.saveAsItem,
		viewer.reloadItem, fyne.NewMenuItemSeparator(), quitItem)

	viewer.editValueItem = fyne.NewMenuItem("Edit Value", viewer.beginEditValue)
	viewer.editValueItem.Shortcut = shortcut(fyne.KeyE, 0)
	viewer.undoItem = fyne.NewMenuItem("Undo", viewer.undo)
	viewer.undoItem.Shortcut = shortcut(fyne.KeyZ, 0)
	viewer.redoItem = fyne.NewMenuItem("Redo", viewer.redo)
	viewer.redoItem.Shortcut = shortcut(fyne.KeyY, 0)
	viewer.editMenu = fyne.NewMenu("Edit", viewer.editValueItem,
		fyne.NewMenuItemSeparator(), viewer.undoItem, viewer.redoItem)

	viewer.spaciousItem = fyne.NewMenuItem("Spacious View", func() { viewer.setViewMode(ViewModeSpacious) })
	viewer.compactItem = fyne.NewMenuItem("Compact View", func() { viewer.setViewMode(ViewModeCompact) })
	viewer.themeItem = fyne.NewMenuItem("Dark Mode", viewer.toggleThemeMode)
	viewer.searchSettingsItem = fyne.NewMenuItem("Search Settings…", viewer.showSearchSettings)
	viewer.viewMenu = fyne.NewMenu("View", viewer.spaciousItem, viewer.compactItem,
		fyne.NewMenuItemSeparator(), viewer.themeItem, viewer.searchSettingsItem)

	aboutItem := fyne.NewMenuItem("About YAML Viewer", viewer.showAbout)
	viewer.aboutMenu = fyne.NewMenu("About", aboutItem)
	viewer.mainMenu = fyne.NewMainMenu(viewer.fileMenu, viewer.editMenu, viewer.viewMenu, viewer.aboutMenu)
	viewer.window.SetMainMenu(viewer.mainMenu)

	for _, item := range []*fyne.MenuItem{openItem, viewer.saveItem, viewer.saveAsItem, viewer.reloadItem,
		viewer.editValueItem, viewer.undoItem, viewer.redoItem} {
		if item.Shortcut != nil {
			menuItem := item
			viewer.window.Canvas().AddShortcut(item.Shortcut, func(shortcut fyne.Shortcut) {
				if menuItem.Action != nil {
					menuItem.Action()
				}
			})
		}
	}
}

func (viewer *Viewer) registerShortcuts() {
	canvas := viewer.window.Canvas()
	modifier := fyne.KeyModifierShortcutDefault
	canvas.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyF, Modifier: modifier}, func(fyne.Shortcut) {
		canvas.Focus(viewer.searchEntry)
	})
	canvas.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyEscape}, func(fyne.Shortcut) {
		viewer.searchEntry.SetText("")
	})
}

// Show displays the window and starts the Fyne event loop when used directly.
// Most callers should use ShowAndRun from main instead.
func (viewer *Viewer) Show() {
	viewer.window.Show()
}

// ShowAndRun displays the window and blocks until it exits.
func (viewer *Viewer) ShowAndRun() {
	viewer.window.ShowAndRun()
}
