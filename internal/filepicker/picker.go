// Package filepicker provides the platform-native file selection used by the
// application.
package filepicker

import (
	"errors"
	"fmt"

	nativeDialog "github.com/hajimehoshi/dialog"
)

// Picker selects a single file from the host operating system.
type Picker interface {
	Open(startDir string) (string, error)
}

// NativePicker uses the operating system's file picker rather than a widget
// rendered by the application's UI toolkit.
type NativePicker struct{}

// NewNative returns a picker backed by the host operating system.
func NewNative() Picker {
	return NativePicker{}
}

// Open displays the native open-file dialog. An empty start directory lets the
// operating system choose its default location.
func (NativePicker) Open(startDir string) (path string, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			path = ""
			err = fmt.Errorf("native file picker unavailable: %v", recovered)
		}
	}()

	builder := nativeDialog.File().Title("Open YAML file")
	if startDir != "" {
		builder.SetStartDir(startDir)
	}
	return builder.Load()
}

// IsCancelled reports whether the user closed the picker without selecting a
// file.
func IsCancelled(err error) bool {
	return errors.Is(err, nativeDialog.ErrCancelled)
}
