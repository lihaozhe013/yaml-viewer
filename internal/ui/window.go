// Package ui contains the Fyne presentation for the YAML viewer.
package ui

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"yamlviewer/internal/appstate"
	appconfig "yamlviewer/internal/config"
	"yamlviewer/internal/fileio"
	"yamlviewer/internal/filepicker"
	"yamlviewer/internal/history"
	"yamlviewer/internal/logging"
	"yamlviewer/internal/search"
)

// Viewer is the desktop application controller. Its fields that reference
// widgets are accessed on the Fyne UI thread; load and search work happens in
// goroutines and returns through fyne.Do.
type Viewer struct {
	app        fyne.App
	window     fyne.Window
	state      *appstate.State
	recent     *fileio.RecentFiles
	picker     filepicker.Picker
	config     appconfig.Config
	history    *history.History
	viewMode   ViewMode
	themeMode  appconfig.ThemeMode
	searchMode search.Mode

	current *fileio.LoadedFile
	items   map[string]treeItem
	visible map[string]bool
	results []search.Result

	tree                 *widget.Tree
	inspector            *fyne.Container
	searchEntry          *widget.Entry
	searchSettingsButton *widget.Button
	expandCollapseButton *widget.Button
	recentSelect         *widget.Select
	status               *widget.Label
	errorLabel           *widget.Label
	valueEditor          *focusCancelEntry
	lastError            error
	editingNode          string
	documentID           uint64
	saving               bool
	closing              bool
	mainMenu             *fyne.MainMenu
	fileMenu             *fyne.Menu
	editMenu             *fyne.Menu
	viewMenu             *fyne.Menu
	aboutMenu            *fyne.Menu
	saveItem             *fyne.MenuItem
	saveAsItem           *fyne.MenuItem
	reloadItem           *fyne.MenuItem
	editValueItem        *fyne.MenuItem
	undoItem             *fyne.MenuItem
	redoItem             *fyne.MenuItem
	spaciousItem         *fyne.MenuItem
	compactItem          *fyne.MenuItem
	themeItem            *fyne.MenuItem
	searchSettingsItem   *fyne.MenuItem
	programmatic         bool
	searchMu             sync.Mutex
	searchTimer          *time.Timer
}

// New creates the application window and its widgets.
func New(application fyne.App) *Viewer {
	storedConfig, err := appconfig.Load()
	if err != nil {
		logging.Debugf("config", "load failed: %v", err)
	}
	viewer := &Viewer{
		app:        application,
		state:      appstate.New(),
		recent:     fileio.NewRecentFiles(10),
		picker:     filepicker.NewNative(),
		config:     storedConfig,
		history:    history.New(1000),
		viewMode:   ViewModeSpacious,
		themeMode:  appconfig.NormalizeThemeMode(storedConfig.ThemeMode),
		searchMode: search.NormalizeMode(search.Mode(storedConfig.SearchMode)),
		items:      make(map[string]treeItem),
		visible:    make(map[string]bool),
	}
	viewer.setThemeMode(viewer.themeMode)
	if storedConfig.LastOpenedFile != "" {
		viewer.recent.Add(storedConfig.LastOpenedFile)
	}
	viewer.build()
	return viewer
}
