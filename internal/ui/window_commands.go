package ui

import (
	"fyne.io/fyne/v2/widget"

	appconfig "yamlviewer/internal/config"
	"yamlviewer/internal/search"
	"yamlviewer/internal/yamlmodel"
)

func (viewer *Viewer) updateCommands() {
	if viewer.saveItem == nil {
		return
	}
	dirty := viewer.isDirty()
	viewer.saveItem.Disabled = viewer.current == nil || !dirty || viewer.saving
	viewer.saveAsItem.Disabled = viewer.current == nil || viewer.saving
	viewer.reloadItem.Disabled = viewer.current == nil || viewer.saving
	node := viewer.selectedNode()
	viewer.editValueItem.Disabled = viewer.current == nil || node == nil || node.Kind != yamlmodel.ScalarNode || viewer.saving
	viewer.undoItem.Disabled = viewer.current == nil || !viewer.history.CanUndo() || viewer.saving
	viewer.redoItem.Disabled = viewer.current == nil || !viewer.history.CanRedo() || viewer.saving
	viewer.spaciousItem.Checked = viewer.viewMode == ViewModeSpacious
	viewer.compactItem.Checked = viewer.viewMode == ViewModeCompact
	viewer.compactItem.Disabled = true
	viewer.themeItem.Checked = viewer.themeMode == appconfig.ThemeModeDark
	viewer.updateSearchControls()
	viewer.updateExpandCollapseButton()
	viewer.mainMenu.Refresh()
}

func (viewer *Viewer) updateSearchControls() {
	if viewer.searchSettingsButton != nil {
		viewer.searchSettingsButton.SetText(viewer.searchModeLabel())
		if viewer.searchMode == search.ModeKeyword {
			viewer.searchSettingsButton.Importance = widget.HighImportance
		} else {
			viewer.searchSettingsButton.Importance = widget.MediumImportance
		}
		viewer.searchSettingsButton.Refresh()
	}
}

func (viewer *Viewer) updateExpandCollapseButton() {
	if viewer.expandCollapseButton == nil {
		return
	}
	viewer.expandCollapseButton.SetText(viewer.expandCollapseLabel())
	if viewer.canToggleSelectedBranch() {
		viewer.expandCollapseButton.Enable()
	} else {
		viewer.expandCollapseButton.Disable()
	}
}

func (viewer *Viewer) expandCollapseLabel() string {
	if viewer.selectedBranchFullyExpanded() {
		return "Collapse"
	}
	return "Expand"
}

func (viewer *Viewer) canToggleSelectedBranch() bool {
	node := viewer.selectedNode()
	return viewer.current != nil && node != nil && len(node.Children) > 0
}

func (viewer *Viewer) selectedBranchFullyExpanded() bool {
	if !viewer.canToggleSelectedBranch() {
		return false
	}
	return viewer.subtreeExpanded(viewer.selectedNode())
}

func (viewer *Viewer) subtreeExpanded(node *yamlmodel.Node) bool {
	if node == nil || len(node.Children) == 0 {
		return true
	}
	if viewer.tree == nil || !viewer.tree.IsBranchOpen(node.ID) {
		return false
	}
	for _, child := range node.Children {
		if !viewer.subtreeExpanded(child) {
			return false
		}
	}
	return true
}

func (viewer *Viewer) toggleSelectedBranch() {
	if !viewer.canToggleSelectedBranch() {
		return
	}
	expand := !viewer.selectedBranchFullyExpanded()
	viewer.programmatic = true
	defer func() { viewer.programmatic = false }()
	viewer.setSubtreeExpanded(viewer.selectedNode(), expand)
	viewer.updateExpandCollapseButton()
}

func (viewer *Viewer) setSubtreeExpanded(node *yamlmodel.Node, expanded bool) {
	if node == nil || len(node.Children) == 0 {
		return
	}
	if expanded {
		viewer.state.Expanded[node.ID] = true
		viewer.tree.OpenBranch(node.ID)
		for _, child := range node.Children {
			viewer.setSubtreeExpanded(child, true)
		}
		return
	}
	for _, child := range node.Children {
		viewer.setSubtreeExpanded(child, false)
	}
	viewer.state.Expanded[node.ID] = false
	viewer.tree.CloseBranch(node.ID)
}
