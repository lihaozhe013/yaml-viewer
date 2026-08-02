package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// focusCancelEntry keeps the standard Fyne entry behavior while notifying the
// viewer when the editor loses focus.
type focusCancelEntry struct {
	widget.BaseWidget
	*widget.Entry
	onFocusLost           func()
	suppressNextFocusLost bool
}

func newFocusCancelEntry(onFocusLost func()) *focusCancelEntry {
	entry := &focusCancelEntry{
		Entry:       widget.NewMultiLineEntry(),
		onFocusLost: onFocusLost,
	}
	entry.BaseWidget.ExtendBaseWidget(entry)
	return entry
}

func (entry *focusCancelEntry) MinSize() fyne.Size {
	return entry.BaseWidget.MinSize()
}

func (entry *focusCancelEntry) Resize(size fyne.Size) {
	entry.BaseWidget.Resize(size)
}

func (entry *focusCancelEntry) Move(position fyne.Position) {
	entry.BaseWidget.Move(position)
}

func (entry *focusCancelEntry) Position() fyne.Position {
	return entry.BaseWidget.Position()
}

func (entry *focusCancelEntry) Size() fyne.Size {
	return entry.BaseWidget.Size()
}

func (entry *focusCancelEntry) Hide() {
	entry.BaseWidget.Hide()
}

func (entry *focusCancelEntry) Show() {
	entry.BaseWidget.Show()
}

func (entry *focusCancelEntry) Visible() bool {
	return entry.BaseWidget.Visible()
}

func (entry *focusCancelEntry) Refresh() {
	entry.BaseWidget.Refresh()
}

func (entry *focusCancelEntry) CreateRenderer() fyne.WidgetRenderer {
	return &focusCancelEntryRenderer{
		entry:    entry,
		delegate: entry.Entry.CreateRenderer(),
	}
}

func (entry *focusCancelEntry) FocusGained() {
	entry.Entry.FocusGained()
	entry.Refresh()
}

func (entry *focusCancelEntry) FocusLost() {
	entry.Entry.FocusLost()
	if entry.suppressNextFocusLost {
		entry.suppressNextFocusLost = false
		entry.Refresh()
		return
	}
	if entry.onFocusLost == nil {
		entry.Refresh()
		return
	}
	fyne.Do(func() {
		entry.Refresh()
		entry.onFocusLost()
	})
}

func (entry *focusCancelEntry) preserveNextFocusLoss() {
	entry.suppressNextFocusLost = true
}

func (entry *focusCancelEntry) Tapped(event *fyne.PointEvent) {
	entry.focus()
	entry.Entry.Tapped(event)
	entry.Refresh()
}

func (entry *focusCancelEntry) MouseDown(event *desktop.MouseEvent) {
	entry.focus()
	entry.Entry.MouseDown(event)
	entry.Refresh()
}

func (entry *focusCancelEntry) MouseUp(event *desktop.MouseEvent) {
	entry.Entry.MouseUp(event)
	entry.Refresh()
}

func (entry *focusCancelEntry) Dragged(event *fyne.DragEvent) {
	entry.Entry.Dragged(event)
	entry.Refresh()
}

func (entry *focusCancelEntry) DragEnd() {
	entry.Entry.DragEnd()
	entry.Refresh()
}

func (entry *focusCancelEntry) DoubleTapped(event *fyne.PointEvent) {
	entry.Entry.DoubleTapped(event)
	entry.Refresh()
}

func (entry *focusCancelEntry) TypedRune(r rune) {
	entry.Entry.TypedRune(r)
	entry.Refresh()
}

func (entry *focusCancelEntry) TypedKey(event *fyne.KeyEvent) {
	entry.Entry.TypedKey(event)
	entry.Refresh()
}

func (entry *focusCancelEntry) SetText(text string) {
	entry.Entry.SetText(text)
	entry.Refresh()
}

func (entry *focusCancelEntry) focus() {
	if application := fyne.CurrentApp(); application != nil {
		if canvas := application.Driver().CanvasForObject(entry); canvas != nil {
			canvas.Focus(entry)
		}
	}
}

type focusCancelEntryRenderer struct {
	entry    *focusCancelEntry
	delegate fyne.WidgetRenderer
}

func (renderer *focusCancelEntryRenderer) Destroy() {
	renderer.delegate.Destroy()
}

func (renderer *focusCancelEntryRenderer) Layout(size fyne.Size) {
	// Keep the embedded entry's geometry in sync because its renderer uses
	// that size when refreshing cursor and selection state.
	renderer.entry.Entry.Resize(size)
	renderer.delegate.Layout(size)
}

func (renderer *focusCancelEntryRenderer) MinSize() fyne.Size {
	return renderer.delegate.MinSize()
}

func (renderer *focusCancelEntryRenderer) Objects() []fyne.CanvasObject {
	return renderer.delegate.Objects()
}

func (renderer *focusCancelEntryRenderer) Refresh() {
	renderer.delegate.Refresh()
}

var (
	_ fyne.Widget         = (*focusCancelEntry)(nil)
	_ fyne.Focusable      = (*focusCancelEntry)(nil)
	_ fyne.Tappable       = (*focusCancelEntry)(nil)
	_ fyne.DoubleTappable = (*focusCancelEntry)(nil)
	_ fyne.Draggable      = (*focusCancelEntry)(nil)
	_ desktop.Mouseable   = (*focusCancelEntry)(nil)
)

// editorActionButton keeps editor actions from being mistaken for an outside
// click when the button clears the input focus before invoking its callback.
type editorActionButton struct {
	widget.BaseWidget
	button            *widget.Button
	preserveFocusLoss func()
}

func newEditorActionButton(label string, preserveFocusLoss func(), action func()) *editorActionButton {
	button := &editorActionButton{
		button:            widget.NewButton(label, action),
		preserveFocusLoss: preserveFocusLoss,
	}
	button.BaseWidget.ExtendBaseWidget(button)
	return button
}

func (button *editorActionButton) MinSize() fyne.Size {
	return button.BaseWidget.MinSize()
}

func (button *editorActionButton) Resize(size fyne.Size) {
	button.BaseWidget.Resize(size)
}

func (button *editorActionButton) Move(position fyne.Position) {
	button.BaseWidget.Move(position)
}

func (button *editorActionButton) Position() fyne.Position {
	return button.BaseWidget.Position()
}

func (button *editorActionButton) Size() fyne.Size {
	return button.BaseWidget.Size()
}

func (button *editorActionButton) Hide() {
	button.BaseWidget.Hide()
}

func (button *editorActionButton) Show() {
	button.BaseWidget.Show()
}

func (button *editorActionButton) Visible() bool {
	return button.BaseWidget.Visible()
}

func (button *editorActionButton) Refresh() {
	button.BaseWidget.Refresh()
}

func (button *editorActionButton) CreateRenderer() fyne.WidgetRenderer {
	return &editorActionButtonRenderer{
		button:   button,
		delegate: button.button.CreateRenderer(),
	}
}

func (button *editorActionButton) Tapped(event *fyne.PointEvent) {
	if button.button.Disabled() {
		return
	}
	if button.preserveFocusLoss != nil {
		button.preserveFocusLoss()
	}
	if application := fyne.CurrentApp(); application != nil {
		if canvas := application.Driver().CanvasForObject(button); canvas != nil {
			canvas.Focus(nil)
		}
	}
	button.button.Tapped(event)
}

func (button *editorActionButton) MouseIn(event *desktop.MouseEvent) {
	button.button.MouseIn(event)
	button.Refresh()
}

func (button *editorActionButton) MouseMoved(event *desktop.MouseEvent) {
	button.button.MouseMoved(event)
	button.Refresh()
}

func (button *editorActionButton) MouseOut() {
	button.button.MouseOut()
	button.Refresh()
}

func (button *editorActionButton) FocusGained() {
	button.button.FocusGained()
	button.Refresh()
}

func (button *editorActionButton) FocusLost() {
	button.button.FocusLost()
	button.Refresh()
}

func (button *editorActionButton) TypedRune(r rune) {
	button.button.TypedRune(r)
}

func (button *editorActionButton) TypedKey(event *fyne.KeyEvent) {
	button.button.TypedKey(event)
}

func (button *editorActionButton) Disable() {
	button.button.Disable()
	button.Refresh()
}

func (button *editorActionButton) Enable() {
	button.button.Enable()
	button.Refresh()
}

func (button *editorActionButton) Disabled() bool {
	return button.button.Disabled()
}

type editorActionButtonRenderer struct {
	button   *editorActionButton
	delegate fyne.WidgetRenderer
}

func (renderer *editorActionButtonRenderer) Destroy() {
	renderer.delegate.Destroy()
}

func (renderer *editorActionButtonRenderer) Layout(size fyne.Size) {
	renderer.button.button.Resize(size)
	renderer.delegate.Layout(size)
}

func (renderer *editorActionButtonRenderer) MinSize() fyne.Size {
	return renderer.delegate.MinSize()
}

func (renderer *editorActionButtonRenderer) Objects() []fyne.CanvasObject {
	return renderer.delegate.Objects()
}

func (renderer *editorActionButtonRenderer) Refresh() {
	renderer.delegate.Refresh()
}

var (
	_ fyne.Widget       = (*editorActionButton)(nil)
	_ fyne.Focusable    = (*editorActionButton)(nil)
	_ fyne.Tappable     = (*editorActionButton)(nil)
	_ fyne.Disableable  = (*editorActionButton)(nil)
	_ desktop.Hoverable = (*editorActionButton)(nil)
)
