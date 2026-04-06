package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/Kaffyn/Vectora/desktop"
)

func main() {
	fyneApp := app.New()
	fyneApp.SetIcon(nil)

	// Initialize desktop application
	desktopApp, err := desktop.NewApp(fyneApp)
	if err != nil {
		log.Fatalf("Failed to initialize desktop app: %v", err)
	}

	// Create and show main window with title
	w := fyneApp.NewWindow("Vectora Desktop")
	w.SetContent(desktopApp.BuildUI())

	// Set window size to reasonable default
	w.Resize(fyne.NewSize(1200, 800))

	w.ShowAndRun()
}
