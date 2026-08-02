// Package assets contains the application's bundled visual resources.
package assets

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed yaml-viewer.svg
var iconSVG []byte

// AppIcon returns the application icon as a Fyne resource. Keeping the SVG
// bundled lets each desktop driver render it at the appropriate size.
func AppIcon() fyne.Resource {
	return fyne.NewStaticResource("yaml-viewer.svg", iconSVG)
}
