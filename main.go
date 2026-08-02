package main

import (
	"os"

	"fyne.io/fyne/v2/app"

	"yamlviewer/internal/assets"
	"yamlviewer/internal/ui"
)

func main() {
	application := app.NewWithID("yamlviewer")
	application.SetIcon(assets.AppIcon())
	viewer := ui.New(application)
	if len(os.Args) > 1 {
		viewer.OpenPath(os.Args[1])
	} else {
		viewer.OpenLastPath()
	}
	viewer.ShowAndRun()
}
