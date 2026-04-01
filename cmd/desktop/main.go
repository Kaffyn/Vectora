package main

import (
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/Kaffyn/Vectora/desktop"
)

func main() {
	// Configurar logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Criar aplicação Fyne
	fyneApp := app.New()

	// Carregar preferences
	preferences := desktop.LoadPreferences()

	// Criar gerenciador de aplicação
	appManager, err := desktop.NewApp(fyneApp, preferences)
	if err != nil {
		log.Fatalf("[ERROR] Erro ao inicializar aplicação: %v\n", err)
	}

	// Criar window principal
	w := fyneApp.NewWindow("Vectora Desktop")
	w.SetContent(appManager.BuildUI())

	// Restaurar tamanho da window
	if preferences.WindowWidth > 0 && preferences.WindowHeight > 0 {
		w.Resize(fyne.NewSize(float32(preferences.WindowWidth), float32(preferences.WindowHeight)))
	} else {
		w.Resize(fyne.NewSize(1400, 900))
	}

	// Salvar tamanho ao fechar
	w.SetOnClosed(func() {
		preferences.WindowWidth = int(w.Canvas().Size().Width)
		preferences.WindowHeight = int(w.Canvas().Size().Height)
		preferences.Save()
		os.Exit(0)
	})

	w.ShowAndRun()
}
