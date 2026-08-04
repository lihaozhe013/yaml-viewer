package ui

import (
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"yamlviewer/internal/fileio"
	"yamlviewer/internal/filepicker"
)

// OpenPath opens a path asynchronously. It is exported for command-line
// startup and tests that want to exercise the same loading flow as the UI.
func (viewer *Viewer) OpenPath(path string) {
	viewer.openPath(path)
}

// OpenLastPath restores the file saved in the application configuration. It
// is intentionally separate from New so an explicit command-line path can
// take precedence during startup.
func (viewer *Viewer) OpenLastPath() {
	if viewer.config.LastOpenedFile != "" {
		viewer.openPath(viewer.config.LastOpenedFile)
	}
}

func (viewer *Viewer) openDialog() {
	startDir := viewer.startDirectory()
	go func() {
		path, err := viewer.picker.Open(startDir)
		fyne.Do(func() {
			if err != nil {
				if !filepicker.IsCancelled(err) {
					viewer.showError(err)
				}
				return
			}
			if path != "" {
				viewer.requestOpenPath(path)
			}
		})
	}()
}

func (viewer *Viewer) startDirectory() string {
	if viewer.current != nil {
		return filepath.Dir(viewer.current.Path)
	}
	if viewer.config.LastOpenedFile != "" {
		return filepath.Dir(viewer.config.LastOpenedFile)
	}
	return ""
}

func (viewer *Viewer) reload() {
	if viewer.current == nil {
		return
	}
	path := viewer.current.Path
	viewer.requestOpenPath(path)
}

func (viewer *Viewer) requestOpenPath(path string) {
	if strings.TrimSpace(path) == "" {
		return
	}
	if viewer.isDirty() {
		viewer.confirmUnsaved(func() { viewer.openPath(path) })
		return
	}
	viewer.openPath(path)
}

func (viewer *Viewer) requestClose() {
	if viewer.closing {
		return
	}
	if viewer.isDirty() {
		viewer.confirmUnsaved(viewer.finishClose)
		return
	}
	viewer.finishClose()
}

func (viewer *Viewer) finishClose() {
	viewer.closing = true
	viewer.window.SetCloseIntercept(nil)
	viewer.window.Close()
}

func (viewer *Viewer) confirmUnsaved(continueAction func()) {
	message := widget.NewLabel("The current YAML file has unsaved changes.")
	message.Wrapping = fyne.TextWrapWord
	unsaved := dialog.NewCustomWithoutButtons("Unsaved Changes", container.NewVBox(message), viewer.window)
	buttons := []fyne.CanvasObject{
		widget.NewButton("Cancel", func() { unsaved.Dismiss() }),
		widget.NewButton("Discard", func() {
			unsaved.Dismiss()
			continueAction()
		}),
		widget.NewButton("Save", func() {
			unsaved.Dismiss()
			viewer.saveCurrent(func() { continueAction() })
		}),
	}
	buttons[2].(*widget.Button).Importance = widget.HighImportance
	unsaved.SetButtons(buttons)
	unsaved.Show()
}

func (viewer *Viewer) save() {
	if viewer.current == nil || !viewer.isDirty() || viewer.saving {
		return
	}
	viewer.saveCurrent(nil)
}

func (viewer *Viewer) saveAs() {
	if viewer.current == nil || viewer.saving {
		return
	}
	startDir := viewer.startDirectory()
	startFile := viewer.current.Name
	go func() {
		path, err := viewer.picker.Save(startDir, startFile)
		fyne.Do(func() {
			if err != nil {
				if !filepicker.IsCancelled(err) {
					viewer.showError(err)
				}
				return
			}
			if path != "" {
				viewer.saveToPath(path, true, nil)
			}
		})
	}()
}

func (viewer *Viewer) saveCurrent(onSuccess func()) {
	if viewer.current == nil || viewer.saving {
		return
	}
	viewer.saveToPath(viewer.current.Path, false, onSuccess)
}

func (viewer *Viewer) saveToPath(path string, updatePath bool, onSuccess func()) {
	if viewer.current == nil || viewer.saving {
		return
	}
	path, err := filepath.Abs(path)
	if err != nil {
		viewer.showError(err)
		return
	}
	data, err := viewer.current.Model.Marshal()
	if err != nil {
		viewer.showError(err)
		return
	}
	loaded := viewer.current
	documentID := viewer.documentID
	version := viewer.history.Version()
	viewer.saving = true
	viewer.status.SetText("Saving " + filepath.Base(path) + "…")
	viewer.updateCommands()
	go func() {
		err := fileio.WriteAtomic(path, data)
		fyne.Do(func() {
			viewer.saving = false
			if err != nil {
				viewer.showError(err)
				viewer.updateCommands()
				return
			}
			if viewer.current == loaded && viewer.documentID == documentID {
				if updatePath {
					loaded.Path = path
					loaded.Name = filepath.Base(path)
					viewer.config.LastOpenedFile = path
					viewer.recent.Add(path)
					viewer.recentSelect.SetOptions(viewer.recent.List())
				}
				if viewer.history.Version() == version {
					viewer.history.MarkSaved()
				}
				viewer.lastError = nil
				viewer.errorLabel.SetText("")
				viewer.updateStatus()
			}
			viewer.updateCommands()
			if onSuccess != nil {
				onSuccess()
			}
		})
	}()
}

func (viewer *Viewer) openPath(path string) {
	if strings.TrimSpace(path) == "" {
		return
	}
	generation := viewer.state.Begin()
	viewer.documentID++
	documentID := viewer.documentID
	viewer.errorLabel.SetText("")
	viewer.status.SetText("Loading " + filepath.Base(path) + "…")
	go func() {
		loaded, err := fileio.Load(path)
		fyne.Do(func() {
			if !viewer.state.ApplyLoad(generation, loaded, err) {
				return
			}
			if err != nil {
				viewer.showError(err)
				viewer.updateCommands()
				return
			}
			if viewer.documentID != documentID {
				return
			}
			viewer.current = loaded
			viewer.history.Reset()
			viewer.editingNode = ""
			viewer.lastError = nil
			viewer.config.LastOpenedFile = loaded.Path
			viewer.recent.Add(loaded.Path)
			viewer.recentSelect.SetOptions(viewer.recent.List())
			viewer.updateInspector(nil)
			viewer.refreshTree()
			viewer.scheduleSearchNow()
			viewer.updateCommands()
		})
	}()
}
