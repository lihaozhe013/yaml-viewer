package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type detailSection string

const (
	sectionIdentity detailSection = "identity"
	sectionValue    detailSection = "value"
	sectionMetadata detailSection = "metadata"
	sectionComments detailSection = "comments"
	sectionWarning  detailSection = "warning"
)

type detailPalette struct {
	foreground color.Color
	muted      color.Color

	identitySurface color.Color
	identityAccent  color.Color
	valueSurface    color.Color
	valueAccent     color.Color
	metadataSurface color.Color
	metadataAccent  color.Color
	commentSurface  color.Color
	commentAccent   color.Color
	warningSurface  color.Color
	warningAccent   color.Color

	codeBackground color.Color
	codeForeground color.Color
	border         color.Color
}

func currentDetailPalette() detailPalette {
	variant := theme.VariantLight
	activeTheme := theme.DefaultTheme()
	if application := fyne.CurrentApp(); application != nil {
		variant = application.Settings().ThemeVariant()
		if application.Settings().Theme() != nil {
			activeTheme = application.Settings().Theme()
		}
	}
	return paletteForTheme(activeTheme, variant)
}

func paletteFor(variant fyne.ThemeVariant) detailPalette {
	if variant == theme.VariantDark {
		return paletteWithText(variant,
			color.NRGBA{R: 231, G: 239, B: 250, A: 255},
			color.NRGBA{R: 150, G: 162, B: 180, A: 255},
		)
	}
	return paletteWithText(variant,
		color.NRGBA{R: 35, G: 42, B: 54, A: 255},
		color.NRGBA{R: 104, G: 116, B: 132, A: 255},
	)
}

func paletteForTheme(activeTheme fyne.Theme, variant fyne.ThemeVariant) detailPalette {
	foreground := activeTheme.Color(theme.ColorNameForeground, variant)
	muted := activeTheme.Color(theme.ColorNameDisabled, variant)
	return paletteWithText(variant, foreground, muted)
}

func paletteWithText(variant fyne.ThemeVariant, foreground, muted color.Color) detailPalette {
	if variant == theme.VariantDark {
		return detailPalette{
			foreground:      foreground,
			muted:           muted,
			identitySurface: color.NRGBA{R: 32, G: 50, B: 80, A: 255},
			identityAccent:  color.NRGBA{R: 119, G: 166, B: 255, A: 255},
			valueSurface:    color.NRGBA{R: 55, G: 44, B: 83, A: 255},
			valueAccent:     color.NRGBA{R: 177, G: 143, B: 255, A: 255},
			metadataSurface: color.NRGBA{R: 67, G: 55, B: 31, A: 255},
			metadataAccent:  color.NRGBA{R: 235, G: 181, B: 79, A: 255},
			commentSurface:  color.NRGBA{R: 30, G: 65, B: 54, A: 255},
			commentAccent:   color.NRGBA{R: 105, G: 208, B: 162, A: 255},
			warningSurface:  color.NRGBA{R: 76, G: 39, B: 42, A: 255},
			warningAccent:   color.NRGBA{R: 255, G: 130, B: 130, A: 255},
			codeBackground:  color.NRGBA{R: 20, G: 25, B: 33, A: 255},
			codeForeground:  color.NRGBA{R: 231, G: 239, B: 250, A: 255},
			border:          color.NRGBA{R: 71, G: 82, B: 99, A: 255},
		}
	}

	return detailPalette{
		foreground:      foreground,
		muted:           muted,
		identitySurface: color.NRGBA{R: 235, G: 243, B: 255, A: 255},
		identityAccent:  color.NRGBA{R: 71, G: 120, B: 232, A: 255},
		valueSurface:    color.NRGBA{R: 245, G: 240, B: 255, A: 255},
		valueAccent:     color.NRGBA{R: 119, G: 84, B: 232, A: 255},
		metadataSurface: color.NRGBA{R: 255, G: 247, B: 229, A: 255},
		metadataAccent:  color.NRGBA{R: 197, G: 138, B: 32, A: 255},
		commentSurface:  color.NRGBA{R: 235, G: 249, B: 243, A: 255},
		commentAccent:   color.NRGBA{R: 47, G: 147, B: 110, A: 255},
		warningSurface:  color.NRGBA{R: 255, G: 240, B: 238, A: 255},
		warningAccent:   color.NRGBA{R: 209, G: 77, B: 77, A: 255},
		codeBackground:  color.NRGBA{R: 35, G: 43, B: 56, A: 255},
		codeForeground:  color.NRGBA{R: 235, G: 242, B: 250, A: 255},
		border:          color.NRGBA{R: 215, G: 221, B: 231, A: 255},
	}
}

func (palette detailPalette) surface(section detailSection) color.Color {
	switch section {
	case sectionValue:
		return palette.valueSurface
	case sectionMetadata:
		return palette.metadataSurface
	case sectionComments:
		return palette.commentSurface
	case sectionWarning:
		return palette.warningSurface
	default:
		return palette.identitySurface
	}
}

func (palette detailPalette) accent(section detailSection) color.Color {
	switch section {
	case sectionValue:
		return palette.valueAccent
	case sectionMetadata:
		return palette.metadataAccent
	case sectionComments:
		return palette.commentAccent
	case sectionWarning:
		return palette.warningAccent
	default:
		return palette.identityAccent
	}
}

func detailCard(palette detailPalette, section detailSection, title, subtitle string, content fyne.CanvasObject) fyne.CanvasObject {
	background := canvas.NewRectangle(palette.surface(section))
	background.CornerRadius = 10
	background.StrokeColor = palette.border
	background.StrokeWidth = 1

	accent := canvas.NewRectangle(palette.accent(section))
	accent.SetMinSize(fyne.NewSize(5, 1))

	heading := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	heading.SizeName = theme.SizeNameSubHeadingText
	objects := []fyne.CanvasObject{heading}
	if subtitle != "" {
		subheading := widget.NewLabel(subtitle)
		subheading.SizeName = theme.SizeNameCaptionText
		subheading.TextStyle = fyne.TextStyle{Italic: true}
		objects = append(objects, subheading)
	}
	objects = append(objects, content)
	body := container.NewVBox(objects...)
	inner := container.NewBorder(nil, nil, accent, nil, container.NewPadded(body))
	return container.NewStack(background, inner)
}

func detailGap() fyne.CanvasObject {
	gap := canvas.NewRectangle(color.Transparent)
	gap.SetMinSize(fyne.NewSize(1, 10))
	return gap
}

func statusChip(text string, background, foreground color.Color) fyne.CanvasObject {
	label := canvas.NewText(text, foreground)
	label.TextStyle = fyne.TextStyle{Bold: true}
	label.Alignment = fyne.TextAlignCenter
	fill := canvas.NewRectangle(background)
	fill.CornerRadius = 8
	return container.NewStack(fill, container.NewPadded(label))
}
