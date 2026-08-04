package ui

import (
	"fmt"

	"yamlviewer/internal/search"
	"yamlviewer/internal/yamlmodel"
)

func (viewer *Viewer) undo() {
	if viewer.current == nil || viewer.saving {
		return
	}
	change, ok := viewer.history.Undo()
	if !ok {
		return
	}
	if err := viewer.current.Model.ApplyScalarChange(change, false); err != nil {
		viewer.showError(err)
		return
	}
	viewer.editingNode = ""
	viewer.reindexCurrent()
	viewer.refreshTree()
	viewer.updateInspector(viewer.selectedNode())
	viewer.scheduleSearchNow()
	viewer.updateCommands()
}

func (viewer *Viewer) redo() {
	if viewer.current == nil || viewer.saving {
		return
	}
	change, ok := viewer.history.Redo()
	if !ok {
		return
	}
	if err := viewer.current.Model.ApplyScalarChange(change, true); err != nil {
		viewer.showError(err)
		return
	}
	viewer.editingNode = ""
	viewer.reindexCurrent()
	viewer.refreshTree()
	viewer.updateInspector(viewer.selectedNode())
	viewer.scheduleSearchNow()
	viewer.updateCommands()
}

func (viewer *Viewer) beginEditValue() {
	node := viewer.selectedNode()
	if node == nil || node.Kind != yamlmodel.ScalarNode || viewer.current == nil {
		return
	}
	viewer.editingNode = node.ID
	viewer.updateInspector(node)
	if viewer.valueEditor != nil {
		viewer.window.Canvas().Focus(viewer.valueEditor)
	}
	viewer.updateCommands()
}

func (viewer *Viewer) cancelEditValue(node *yamlmodel.Node) {
	if node != nil && viewer.editingNode == node.ID {
		viewer.editingNode = ""
		viewer.updateInspector(node)
		viewer.updateCommands()
	}
}

func (viewer *Viewer) commitScalarEdit(node *yamlmodel.Node, literal string) error {
	if viewer.current == nil || node == nil {
		return fmt.Errorf("no scalar node is selected")
	}
	change, err := viewer.current.Model.SetScalarLiteral(node.ID, literal)
	if err != nil {
		return err
	}
	if change.Before == change.After {
		viewer.editingNode = ""
		viewer.updateInspector(node)
		viewer.updateCommands()
		return nil
	}
	viewer.history.Commit(change)
	viewer.editingNode = ""
	viewer.lastError = nil
	viewer.errorLabel.SetText("")
	viewer.reindexCurrent()
	viewer.refreshTree()
	viewer.updateInspector(node)
	viewer.scheduleSearchNow()
	viewer.updateCommands()
	return nil
}

func (viewer *Viewer) reindexCurrent() {
	if viewer.current != nil && viewer.current.Model != nil {
		viewer.current.Index = search.NewIndex(viewer.current.Model)
	}
}

func (viewer *Viewer) isDirty() bool {
	return viewer.current != nil && viewer.history.Dirty()
}
