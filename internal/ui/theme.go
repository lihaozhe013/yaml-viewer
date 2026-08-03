package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
)

// fixedVariantTheme keeps the selected color scheme independent from the
// operating system's current theme preference. Fyne's public Settings API
// replaces themes but does not expose a setter for ThemeVariant.
type fixedVariantTheme struct {
	base    fyne.Theme
	variant fyne.ThemeVariant
}

func (theme fixedVariantTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return theme.base.Color(name, theme.variant)
}

func (theme fixedVariantTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.base.Font(style)
}

func (theme fixedVariantTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.base.Icon(name)
}

func (theme fixedVariantTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.base.Size(name)
}
