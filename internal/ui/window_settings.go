package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"yamlviewer/internal/assets"
	appconfig "yamlviewer/internal/config"
	"yamlviewer/internal/search"
)

func (viewer *Viewer) setViewMode(mode ViewMode) {
	if mode == ViewModeCompact {
		return
	}
	viewer.viewMode = mode
	viewer.updateInspector(viewer.selectedNode())
	viewer.updateCommands()
}

func (viewer *Viewer) toggleThemeMode() {
	if viewer.themeMode == appconfig.ThemeModeDark {
		viewer.setThemeMode(appconfig.ThemeModeLight)
		return
	}
	viewer.setThemeMode(appconfig.ThemeModeDark)
}

func (viewer *Viewer) setThemeMode(mode appconfig.ThemeMode) {
	mode = appconfig.NormalizeThemeMode(mode)
	viewer.themeMode = mode
	viewer.config.ThemeMode = mode
	variant := theme.VariantLight
	if mode == appconfig.ThemeModeDark {
		variant = theme.VariantDark
	}
	viewer.app.Settings().SetTheme(fixedVariantTheme{
		base:    theme.DefaultTheme(),
		variant: variant,
	})
	if viewer.inspector != nil {
		viewer.updateInspector(viewer.selectedNode())
	}
	if viewer.mainMenu != nil {
		viewer.updateCommands()
	}
}

func (viewer *Viewer) searchModeLabel() string {
	if viewer.searchMode == search.ModeKeyword {
		return "Search: Keyword Match"
	}
	return "Search: Smart Fuzzy"
}

func (viewer *Viewer) setSearchMode(mode search.Mode) {
	mode = search.NormalizeMode(mode)
	if viewer.searchMode == mode {
		viewer.updateSearchControls()
		return
	}
	viewer.searchMode = mode
	viewer.config.SearchMode = appconfig.SearchMode(mode)
	viewer.updateCommands()
	viewer.scheduleSearchNow()
}

func (viewer *Viewer) showSearchSettings() {
	options := []string{"Smart Fuzzy", "Keyword Match"}
	radio := widget.NewRadioGroup(options, nil)
	if viewer.searchMode == search.ModeKeyword {
		radio.SetSelected("Keyword Match")
	} else {
		radio.SetSelected("Smart Fuzzy")
	}
	radio.OnChanged = func(option string) {
		if option == "Keyword Match" {
			viewer.setSearchMode(search.ModeKeyword)
			return
		}
		viewer.setSearchMode(search.ModeSmartFuzzy)
	}

	description := widget.NewLabel("Choose how search terms are matched against YAML keys, paths, values, and metadata.")
	description.Wrapping = fyne.TextWrapWord
	smartDescription := widget.NewLabel("Smart Fuzzy matches case-insensitively and supports exact, prefix, substring, and character-order matches. It is more forgiving when you only know part of a field name.")
	smartDescription.Wrapping = fyne.TextWrapWord
	keywordDescription := widget.NewLabel("Keyword Match requires every keyword, ignores their order, and matches complete normalized keywords.")
	keywordDescription.Wrapping = fyne.TextWrapWord
	content := container.NewVBox(
		description,
		radio,
		widget.NewSeparator(),
		smartDescription,
		keywordDescription,
	)
	sizeHint := canvas.NewRectangle(color.Transparent)
	sizeHint.SetMinSize(fyne.NewSize(520, 220))
	content = container.NewStack(sizeHint, content)
	dialog.NewCustom("Search Settings", "Close", content, viewer.window).Show()
}

func (viewer *Viewer) showAbout() {
	icon := canvas.NewImageFromResource(assets.AppIcon())
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSquareSize(112))

	title := widget.NewLabelWithStyle("YAML Viewer", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	description := widget.NewLabel("A generic YAML browser and scalar editor built with Go and Fyne.")
	description.Alignment = fyne.TextAlignCenter
	description.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(
		container.NewCenter(icon),
		title,
		description,
	)
	dialog.NewCustom("About YAML Viewer", "Close", content, viewer.window).Show()
}
