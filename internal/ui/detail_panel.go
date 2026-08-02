package ui

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"yamlviewer/internal/display"
	"yamlviewer/internal/yamlmodel"
)

func (viewer *Viewer) updateInspector(node *yamlmodel.Node) {
	viewer.inspector.RemoveAll()
	viewer.valueEditor = nil
	palette := currentDetailPalette()

	if viewer.lastError != nil {
		viewer.inspector.Add(detailCard(
			palette,
			sectionWarning,
			"Load error",
			"The previous document is still open",
			errorContent(viewer.lastError.Error(), palette, func() { viewer.copy(viewer.lastError.Error()) }),
		))
		viewer.inspector.Add(detailGap())
	}

	if node == nil {
		viewer.inspector.Add(emptyDetailState(palette))
		viewer.inspector.Refresh()
		return
	}

	viewer.inspector.Add(detailHeader(viewer, node, palette))
	viewer.inspector.Add(detailGap())
	valueSubtitle := "Read-only YAML representation"
	value := valueContent(viewer, node, palette)
	if viewer.editingNode == node.ID && node.Kind == yamlmodel.ScalarNode {
		valueSubtitle = "Edit one YAML scalar value, then apply the change"
		value = editableValueContent(viewer, node, palette)
	}
	viewer.inspector.Add(detailCard(palette, sectionValue, "Value", valueSubtitle, value))
	viewer.inspector.Add(detailGap())
	viewer.inspector.Add(detailCard(palette, sectionIdentity, "Identity", "Display and location information", identityContent(node)))
	viewer.inspector.Add(detailGap())
	viewer.inspector.Add(detailCard(palette, sectionMetadata, "YAML metadata", "Parser information preserved from the source", metadataContent(node)))

	if hasComments(node) {
		viewer.inspector.Add(detailGap())
		viewer.inspector.Add(detailCard(palette, sectionComments, "Comments", "Comments attached to this YAML node", commentsContent(node)))
	}
	viewer.inspector.Refresh()
}

func emptyDetailState(palette detailPalette) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Select a node", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.SizeName = theme.SizeNameSubHeadingText
	description := widget.NewLabel("Choose a YAML node from the hierarchy to inspect its details.")
	description.Wrapping = fyne.TextWrapWord
	return detailCard(palette, sectionIdentity, "", "", container.NewVBox(title, description))
}

func detailHeader(viewer *Viewer, node *yamlmodel.Node, palette detailPalette) fyne.CanvasObject {
	name := widget.NewLabelWithStyle(display.NodeDisplayName(node), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	name.SizeName = theme.SizeNameHeadingText

	chipText := strings.ToUpper(string(node.Kind))
	chip := statusChip(chipText, palette.identityAccent, colorForChip(palette, sectionIdentity))
	titleRow := container.NewHBox(name, chip)
	if node.Duplicate {
		titleRow.Add(statusChip("DUPLICATE KEY", palette.warningAccent, colorForChip(palette, sectionWarning)))
	}

	path := widget.NewLabel(fmt.Sprintf("%s", node.Path))
	path.TextStyle = fyne.TextStyle{Monospace: true}
	path.Wrapping = fyne.TextWrapWord
	path.Selectable = true

	actions := container.NewHBox(
		widget.NewButton("Copy key", func() { viewer.copy(node.Key) }),
		widget.NewButton("Copy path", func() { viewer.copy(node.Path) }),
	)
	content := container.NewBorder(nil, actions, nil, nil, container.NewVBox(titleRow, path))
	return detailCard(palette, sectionIdentity, "", "", content)
}

func identityContent(node *yamlmodel.Node) fyne.CanvasObject {
	return container.NewGridWithColumns(2,
		detailField("Display name", display.NodeDisplayName(node), false),
		detailField("Raw key", valueOrDash(node.Key), true),
		detailField("Path", node.Path, true),
		detailField("Source", fmt.Sprintf("line %d, column %d", node.Line, node.Column), false),
	)
}

func valueContent(viewer *Viewer, node *yamlmodel.Node, palette detailPalette) fyne.CanvasObject {
	value := valueForDisplay(node)
	grid := widget.NewTextGridFromString(value)
	grid.ShowLineNumbers = false
	style := &widget.CustomTextGridStyle{
		TextStyle: fyne.TextStyle{Monospace: true},
		FGColor:   palette.codeForeground,
		BGColor:   palette.codeBackground,
	}
	for index := range grid.Rows {
		grid.Rows[index].Style = style
	}
	gridBackground := canvas.NewRectangle(palette.codeBackground)
	gridBackground.CornerRadius = 6
	gridBackground.SetMinSize(fyne.NewSize(1, 72))
	var code fyne.CanvasObject = container.NewStack(gridBackground, container.NewScroll(grid))
	if node.Kind == yamlmodel.ScalarNode {
		code = newTappableValue(code, viewer.beginEditValue)
	}
	return container.NewVBox(code, widget.NewButton("Copy value", func() { viewer.copy(value) }))
}

func editableValueContent(viewer *Viewer, node *yamlmodel.Node, palette detailPalette) fyne.CanvasObject {
	entry := newFocusCancelEntry(nil)
	entry.onFocusLost = func() {
		if viewer.valueEditor == entry {
			viewer.cancelEditValue(node)
		}
	}
	entry.SetText(valueForEdit(node))
	entry.Wrapping = fyne.TextWrapWord
	entry.SetMinRowsVisible(4)
	viewer.valueEditor = entry

	validation := widget.NewLabel("")
	validation.Wrapping = fyne.TextWrapWord
	validation.TextStyle = fyne.TextStyle{Italic: true}

	apply := func() {
		if err := viewer.commitScalarEdit(node, entry.Text); err != nil {
			validation.SetText("Invalid value: " + err.Error())
			return
		}
	}
	cancel := func() { viewer.cancelEditValue(node) }
	entry.OnSubmitted = func(string) { apply() }
	preserveFocusLoss := entry.preserveNextFocusLoss
	applyButton := newEditorActionButton("Apply", preserveFocusLoss, apply)
	applyButton.button.Importance = widget.HighImportance

	return container.NewVBox(
		widget.NewLabel("Enter one YAML scalar literal. Use quotes for an explicit string or null for a null value."),
		entry,
		validation,
		container.NewHBox(
			newEditorActionButton("Cancel", preserveFocusLoss, cancel),
			applyButton,
			newEditorActionButton("Copy value", preserveFocusLoss, func() { viewer.copy(entry.Text) }),
		),
		readOnlyText(valueForDisplay(node), palette, false),
	)
}

type tappableValue struct {
	widget.BaseWidget
	content  fyne.CanvasObject
	onTapped func()
}

func newTappableValue(content fyne.CanvasObject, onTapped func()) *tappableValue {
	value := &tappableValue{content: content, onTapped: onTapped}
	value.ExtendBaseWidget(value)
	return value
}

func (value *tappableValue) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(value.content)
}

func (value *tappableValue) Tapped(*fyne.PointEvent) {
	if value.onTapped != nil {
		value.onTapped()
	}
}

func (value *tappableValue) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

var (
	_ fyne.Widget        = (*tappableValue)(nil)
	_ fyne.Tappable      = (*tappableValue)(nil)
	_ desktop.Cursorable = (*tappableValue)(nil)
)

func valueForEdit(node *yamlmodel.Node) string {
	if node == nil {
		return ""
	}
	if node.YAML != "" {
		return node.YAML
	}
	return valueForDisplay(node)
}

func metadataContent(node *yamlmodel.Node) fyne.CanvasObject {
	return container.NewGridWithColumns(2,
		detailField("Type", string(node.Kind), false),
		detailField("Tag", valueOrDash(node.Tag), true),
		detailField("Scalar style", valueOrDash(node.Style), false),
		detailField("Children", fmt.Sprintf("%d", len(node.Children)), false),
		detailField("Anchor", valueOrDash(node.Anchor), true),
		detailField("Alias", valueOrDash(node.Alias), true),
	)
}

func commentsContent(node *yamlmodel.Node) fyne.CanvasObject {
	comments := make([]fyne.CanvasObject, 0, 3)
	if node.Comments.Head != "" {
		comments = append(comments, commentField("Head", node.Comments.Head))
	}
	if node.Comments.Line != "" {
		comments = append(comments, commentField("Line", node.Comments.Line))
	}
	if node.Comments.Foot != "" {
		comments = append(comments, commentField("Foot", node.Comments.Foot))
	}
	return container.NewVBox(comments...)
}

func errorContent(message string, palette detailPalette, copyError func()) fyne.CanvasObject {
	return container.NewVBox(
		readOnlyText(message, palette, true),
		widget.NewButton("Copy error", copyError),
	)
}

func detailField(label, value string, monospace bool) fyne.CanvasObject {
	caption := widget.NewLabel(label)
	caption.SizeName = theme.SizeNameCaptionText
	caption.TextStyle = fyne.TextStyle{Bold: true}
	content := widget.NewLabel(value)
	content.Wrapping = fyne.TextWrapWord
	content.Selectable = true
	if monospace {
		content.TextStyle = fyne.TextStyle{Monospace: true}
	}
	return container.NewVBox(caption, content)
}

func commentField(label, value string) fyne.CanvasObject {
	caption := widget.NewLabelWithStyle(label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	caption.SizeName = theme.SizeNameCaptionText
	text := widget.NewLabel(value)
	text.Wrapping = fyne.TextWrapWord
	text.Selectable = true
	return container.NewVBox(caption, text)
}

func readOnlyText(value string, palette detailPalette, errorStyle bool) fyne.CanvasObject {
	grid := widget.NewTextGridFromString(value)
	grid.ShowLineNumbers = false
	foreground := palette.codeForeground
	background := palette.codeBackground
	if errorStyle {
		foreground = palette.warningAccent
		background = palette.warningSurface
	}
	style := &widget.CustomTextGridStyle{
		TextStyle: fyne.TextStyle{Monospace: true},
		FGColor:   foreground,
		BGColor:   background,
	}
	for index := range grid.Rows {
		grid.Rows[index].Style = style
	}
	fill := canvas.NewRectangle(background)
	fill.CornerRadius = 6
	fill.SetMinSize(fyne.NewSize(1, 56))
	preview := container.NewStack(fill, container.NewScroll(grid))
	return preview
}

func valueForDisplay(node *yamlmodel.Node) string {
	if node == nil {
		return ""
	}
	if node.Kind == yamlmodel.AliasNode {
		return "*" + node.Alias
	}
	if node.Kind == yamlmodel.MappingNode || node.Kind == yamlmodel.SequenceNode {
		if node.YAML != "" {
			return node.YAML
		}
	}
	if node.Value != "" {
		return node.Value
	}
	if node.YAML != "" {
		return node.YAML
	}
	return "null / empty"
}

func hasComments(node *yamlmodel.Node) bool {
	return node != nil && (node.Comments.Head != "" || node.Comments.Line != "" || node.Comments.Foot != "")
}

func colorForChip(palette detailPalette, section detailSection) color.Color {
	if section == sectionWarning {
		return palette.warningSurface
	}
	return palette.identitySurface
}
