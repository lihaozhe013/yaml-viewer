package ui

import (
	"fmt"
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

func (viewer *Viewer) showSettings() {
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
	themeOptions := []string{"Light", "Dark"}
	themeRadio := widget.NewRadioGroup(themeOptions, nil)
	if viewer.themeMode == appconfig.ThemeModeDark {
		themeRadio.SetSelected("Dark")
	} else {
		themeRadio.SetSelected("Light")
	}
	themeRadio.OnChanged = func(option string) {
		if option == "Dark" {
			viewer.setThemeMode(appconfig.ThemeModeDark)
			return
		}
		viewer.setThemeMode(appconfig.ThemeModeLight)
	}

	// YAML format settings
	formatTitle := widget.NewLabelWithStyle("YAML Format", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	indentLabel := widget.NewLabel("Indentation:")
	indentOptions := []string{"2 spaces", "4 spaces"}
	indentSelect := widget.NewSelect(indentOptions, nil)
	if viewer.formatIndent == 4 {
		indentSelect.SetSelected("4 spaces")
	} else {
		indentSelect.SetSelected("2 spaces")
	}
	indentSelect.OnChanged = func(option string) {
		if option == "4 spaces" {
			viewer.formatIndent = 4
		} else {
			viewer.formatIndent = 2
		}
	}
	sortCheck := widget.NewCheck("Sort mapping keys alphabetically when saving", func(checked bool) {
		viewer.formatSortKeys = checked
	})
	sortCheck.SetChecked(viewer.formatSortKeys)

	appearance := widget.NewLabelWithStyle("Appearance", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	content := container.NewVBox(
		description,
		radio,
		widget.NewSeparator(),
		smartDescription,
		keywordDescription,
		widget.NewSeparator(),
		formatTitle,
		container.NewHBox(indentLabel, indentSelect),
		sortCheck,
		widget.NewSeparator(),
		appearance,
		themeRadio,
	)
	sizeHint := canvas.NewRectangle(color.Transparent)
	sizeHint.SetMinSize(fyne.NewSize(520, 360))
	content = container.NewStack(sizeHint, content)
	dialog.NewCustom("Settings", "Close", content, viewer.window).Show()
}

func (viewer *Viewer) showAbout() {
	icon := canvas.NewImageFromResource(assets.AppIcon())
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSquareSize(112))

	title := widget.NewLabelWithStyle("YAML Viewer", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	description := widget.NewLabel("A generic YAML browser and scalar editor built with Go and Fyne.")
	description.Alignment = fyne.TextAlignCenter
	description.Wrapping = fyne.TextWrapWord
	metadata := viewer.app.Metadata()
	metadataLabels := make([]fyne.CanvasObject, 0, 2)
	if metadata.Version != "" {
		metadataLabels = append(metadataLabels, widget.NewLabel("Version "+metadata.Version))
	}
	if metadata.Build > 0 {
		metadataLabels = append(metadataLabels, widget.NewLabel(fmt.Sprintf("Build %d", metadata.Build)))
	}

	content := container.NewVBox(
		container.NewCenter(icon),
		title,
		description,
		container.NewCenter(container.NewHBox(metadataLabels...)),
	)
	sizeHint := canvas.NewRectangle(color.Transparent)
	sizeHint.SetMinSize(fyne.NewSize(420, 230))
	content = container.NewStack(sizeHint, content)
	dialog.NewCustom("About YAML Viewer", "Close", content, viewer.window).Show()
}
