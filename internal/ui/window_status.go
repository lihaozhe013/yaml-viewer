package ui

import (
	"fmt"
	"strings"

	appconfig "yamlviewer/internal/config"
	"yamlviewer/internal/logging"
)

func (viewer *Viewer) saveConfig() {
	if viewer.current != nil {
		viewer.config.LastOpenedFile = viewer.current.Path
	}
	viewer.config.SearchMode = appconfig.SearchMode(viewer.searchMode)
	viewer.config.ThemeMode = viewer.themeMode
	if err := appconfig.Save(viewer.config); err != nil {
		logging.Debugf("config", "save failed: %v", err)
	}
}

func (viewer *Viewer) showError(err error) {
	viewer.lastError = err
	viewer.errorLabel.SetText("Error: " + err.Error())
	viewer.updateInspector(viewer.selectedNode())
	viewer.status.SetText("The previous document is still open")
}

func valueOrDash(value string) string {
	if value == "" {
		return "—"
	}
	return value
}

func (viewer *Viewer) copy(value string) {
	viewer.app.Clipboard().SetContent(value)
	viewer.status.SetText("Copied to clipboard")
}

func (viewer *Viewer) updateStatus() {
	if viewer.current == nil || viewer.current.Model == nil {
		if viewer.lastError == nil {
			viewer.status.SetText("No file open")
		}
		viewer.window.SetTitle("YAML Viewer")
		return
	}
	documents := len(viewer.current.Model.Documents)
	status := fmt.Sprintf("%s | %d document(s)", viewer.current.Name, documents)
	if viewer.current.Model.Empty || allDocumentsEmpty(viewer.current.Model) {
		status = viewer.current.Name + " | Empty Document"
	}
	if strings.TrimSpace(viewer.searchEntry.Text) != "" {
		status += fmt.Sprintf(" | %d match(es)", len(viewer.results))
	}
	if viewer.isDirty() {
		status += " | Unsaved changes"
	}
	viewer.status.SetText(status)
	title := "YAML Viewer — " + viewer.current.Name
	if viewer.isDirty() {
		title += " *"
	}
	viewer.window.SetTitle(title)
}
