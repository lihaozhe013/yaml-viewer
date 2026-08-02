// Package ui contains the Fyne presentation for the YAML viewer.
package ui

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"yamlviewer/internal/appstate"
	appconfig "yamlviewer/internal/config"
	"yamlviewer/internal/display"
	"yamlviewer/internal/fileio"
	"yamlviewer/internal/filepicker"
	"yamlviewer/internal/search"
	"yamlviewer/internal/yamlmodel"
)

const searchDebounce = 125 * time.Millisecond

type treeItem struct {
	node     *yamlmodel.Node
	label    string
	children []string
}

// Viewer is the desktop application controller. Its fields that reference
// widgets are accessed on the Fyne UI thread; load and search work happens in
// goroutines and returns through fyne.Do.
type Viewer struct {
	app    fyne.App
	window fyne.Window
	state  *appstate.State
	recent *fileio.RecentFiles
	picker filepicker.Picker
	config appconfig.Config

	current *fileio.LoadedFile
	items   map[string]treeItem
	visible map[string]bool
	results []search.Result

	tree         *widget.Tree
	inspector    *fyne.Container
	searchEntry  *widget.Entry
	recentSelect *widget.Select
	status       *widget.Label
	errorLabel   *widget.Label
	lastError    error
	programmatic bool
	searchMu     sync.Mutex
	searchTimer  *time.Timer
}

// New creates the application window and its widgets.
func New(application fyne.App) *Viewer {
	storedConfig, err := appconfig.Load()
	if err != nil {
		log.Printf("[config] load failed: %v", err)
	}
	viewer := &Viewer{
		app:     application,
		state:   appstate.New(),
		recent:  fileio.NewRecentFiles(10),
		picker:  filepicker.NewNative(),
		config:  storedConfig,
		items:   make(map[string]treeItem),
		visible: make(map[string]bool),
	}
	if storedConfig.LastOpenedFile != "" {
		viewer.recent.Add(storedConfig.LastOpenedFile)
	}
	viewer.build()
	return viewer
}

func (viewer *Viewer) build() {
	viewer.window = viewer.app.NewWindow("YAML Viewer")
	viewer.window.Resize(fyne.NewSize(1100, 720))

	viewer.searchEntry = widget.NewEntry()
	viewer.searchEntry.SetPlaceHolder("Search fields, paths, values, tags, comments…")
	viewer.searchEntry.OnChanged = func(_ string) {
		viewer.scheduleSearch()
	}

	openButton := widget.NewButton("Open", viewer.openDialog)
	reloadButton := widget.NewButton("Reload", viewer.reload)
	viewer.recentSelect = widget.NewSelect(nil, func(path string) {
		if path != "" {
			viewer.openPath(path)
		}
	})
	viewer.recentSelect.PlaceHolder = "Recent files"
	viewer.recentSelect.SetOptions(viewer.recent.List())
	clearButton := widget.NewButton("Clear", func() { viewer.searchEntry.SetText("") })
	toolbar := container.NewBorder(nil, nil, container.NewHBox(openButton, reloadButton), clearButton,
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
		}
	}
	viewer.tree.OnBranchClosed = func(id widget.TreeNodeID) {
		if !viewer.programmatic {
			viewer.state.Expanded[id] = false
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
	viewer.window.SetOnClosed(viewer.saveConfig)
	viewer.window.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		for _, uri := range uris {
			if uri.Scheme() == "file" {
				viewer.openPath(uri.Path())
				break
			}
		}
	})
	viewer.registerShortcuts()
	viewer.refreshTree()
	viewer.updateInspector(nil)
}

func (viewer *Viewer) registerShortcuts() {
	canvas := viewer.window.Canvas()
	modifier := fyne.KeyModifierShortcutDefault
	canvas.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: modifier}, func(fyne.Shortcut) { viewer.openDialog() })
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

// OpenPath opens a path asynchronously. It is exported for command-line
// startup and tests that want to exercise the same loading flow as the UI.
func (viewer *Viewer) OpenPath(path string) {
	viewer.openPath(path)
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
				viewer.openPath(path)
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
	viewer.openPath(viewer.current.Path)
}

func (viewer *Viewer) openPath(path string) {
	if strings.TrimSpace(path) == "" {
		return
	}
	generation := viewer.state.Begin()
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
				return
			}
			viewer.current = loaded
			viewer.lastError = nil
			viewer.config.LastOpenedFile = loaded.Path
			viewer.recent.Add(loaded.Path)
			viewer.recentSelect.SetOptions(viewer.recent.List())
			viewer.updateInspector(nil)
			viewer.refreshTree()
			viewer.scheduleSearchNow()
		})
	}()
}

func (viewer *Viewer) saveConfig() {
	if viewer.current != nil {
		viewer.config.LastOpenedFile = viewer.current.Path
	}
	if err := appconfig.Save(viewer.config); err != nil {
		log.Printf("[config] save failed: %v", err)
	}
}

func (viewer *Viewer) showError(err error) {
	viewer.lastError = err
	viewer.errorLabel.SetText("Error: " + err.Error())
	viewer.updateInspector(viewer.selectedNode())
	viewer.status.SetText("The previous document is still open")
}

func (viewer *Viewer) scheduleSearch() {
	viewer.searchMu.Lock()
	if viewer.searchTimer != nil {
		viewer.searchTimer.Stop()
	}
	viewer.searchTimer = time.AfterFunc(searchDebounce, viewer.runSearch)
	viewer.searchMu.Unlock()
}

func (viewer *Viewer) scheduleSearchNow() {
	viewer.searchMu.Lock()
	if viewer.searchTimer != nil {
		viewer.searchTimer.Stop()
	}
	viewer.searchMu.Unlock()
	viewer.runSearch()
}

func (viewer *Viewer) runSearch() {
	// Timer callbacks are not Fyne UI callbacks. Marshal the widget reads and
	// generation update back to the UI thread before launching the search work.
	fyne.Do(viewer.beginSearch)
}

func (viewer *Viewer) beginSearch() {
	query := viewer.searchEntry.Text
	loaded := viewer.current
	if loaded == nil {
		return
	}
	generation := viewer.state.Begin()
	go func() {
		results := loaded.Index.Search(query)
		fyne.Do(func() {
			if !viewer.state.ApplySearch(generation, query, results) {
				return
			}
			viewer.results = results
			viewer.visible = search.VisibleIDs(results)
			viewer.refreshTree()
			if len(results) > 0 && strings.TrimSpace(query) != "" {
				viewer.tree.Select(results[0].Node.ID)
			}
		})
	}()
}

func (viewer *Viewer) childrenOf(id string) []string {
	item, ok := viewer.items[id]
	if !ok {
		return nil
	}
	return item.children
}

func (viewer *Viewer) refreshTree() {
	viewer.items = make(map[string]treeItem)
	root := treeItem{label: "YAML hierarchy"}
	viewer.items["tree-root"] = root

	if viewer.current != nil && viewer.current.Model != nil {
		for _, document := range viewer.current.Model.Documents {
			if document.Root == nil {
				id := fmt.Sprintf("document-%d", document.Number)
				viewer.items[id] = treeItem{label: fmt.Sprintf("Document %d (empty)", document.Number)}
				if viewer.includeDocument(document.Number, nil) {
					root.children = append(root.children, id)
				}
				continue
			}
			if !viewer.includeDocument(document.Number, document.Root) {
				continue
			}
			if len(viewer.current.Model.Documents) > 1 {
				id := fmt.Sprintf("document-%d", document.Number)
				viewer.items[id] = treeItem{label: fmt.Sprintf("Document %d", document.Number), children: []string{document.Root.ID}}
				root.children = append(root.children, id)
			} else {
				root.children = append(root.children, document.Root.ID)
			}
			viewer.addNode(document.Root)
		}
	}
	if len(root.children) == 0 {
		if viewer.current == nil || viewer.current.Model == nil {
			root.label = "YAML hierarchy — open a file"
		} else if viewer.current.Model.Empty || allDocumentsEmpty(viewer.current.Model) {
			root.label = "YAML hierarchy — Empty Document"
		} else {
			root.label = "YAML hierarchy — no matches"
		}
	}
	viewer.items["tree-root"] = root
	viewer.tree.Root = "tree-root"
	viewer.tree.Refresh()
	viewer.restoreBranches()
	viewer.updateStatus()
}

func (viewer *Viewer) addNode(node *yamlmodel.Node) {
	if node == nil || !viewer.visibleNode(node) {
		return
	}
	children := make([]string, 0, len(node.Children))
	for _, child := range node.Children {
		if viewer.visibleNode(child) {
			viewer.addNode(child)
			children = append(children, child.ID)
		}
	}
	viewer.items[node.ID] = treeItem{node: node, label: nodeLabel(node), children: children}
}

func (viewer *Viewer) visibleNode(node *yamlmodel.Node) bool {
	return strings.TrimSpace(viewer.searchEntry.Text) == "" || viewer.visible[node.ID]
}

func (viewer *Viewer) includeDocument(number int, root *yamlmodel.Node) bool {
	if strings.TrimSpace(viewer.searchEntry.Text) == "" {
		return true
	}
	if root != nil && viewer.visible[root.ID] {
		return true
	}
	for _, result := range viewer.results {
		if result.Node != nil && result.Node.Path != "" && documentNumber(result.Node.ID) == number {
			return true
		}
	}
	return false
}

func documentNumber(id string) int {
	var number int
	if _, err := fmt.Sscanf(id, "doc-%d-node-", &number); err == nil {
		return number
	}
	return 0
}

func nodeLabel(node *yamlmodel.Node) string {
	label := display.NodeLabel(node)
	if node.KeySet && display.HumanizeKey(node.Key) != node.Key && node.Key != "" {
		label += " [" + node.Key + "]"
	}
	if node.Duplicate {
		label += " (duplicate key)"
	}
	return label
}

func (viewer *Viewer) restoreBranches() {
	viewer.programmatic = true
	defer func() { viewer.programmatic = false }()
	viewer.tree.CloseAllBranches()
	for id, expanded := range viewer.state.Expanded {
		if expanded {
			viewer.tree.OpenBranch(id)
		}
	}
	if strings.TrimSpace(viewer.searchEntry.Text) != "" {
		for id, item := range viewer.items {
			if id != "tree-root" && len(item.children) > 0 && viewer.visible[id] {
				viewer.tree.OpenBranch(id)
			}
		}
		viewer.tree.OpenBranch("tree-root")
	}
}

func (viewer *Viewer) selectTreeItem(id string) {
	item, ok := viewer.items[id]
	if !ok || item.node == nil {
		return
	}
	viewer.state.Selected = item.node
	viewer.updateInspector(item.node)
}

func (viewer *Viewer) selectedNode() *yamlmodel.Node {
	return viewer.state.Selected
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
	viewer.status.SetText(status)
}

func allDocumentsEmpty(file *yamlmodel.File) bool {
	if file == nil || len(file.Documents) == 0 {
		return true
	}
	for _, document := range file.Documents {
		if document.Root != nil {
			return false
		}
	}
	return true
}
