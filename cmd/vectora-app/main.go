package main

import (
	"context"
	"embed"
	"log"
	"net/http"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/Kaffyn/Vectora/internal/infra"
)

//go:embed all:internal/app/out
var assets embed.FS

type App struct {
	ctx    context.Context
	logger *infra.Logger
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) CallIPC(method string, payload string) (string, error) {
	// Bridge para IPC client
	log.Printf("IPC Call: %s with payload: %s", method, payload)
	return `{"success":true}`, nil
}

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "Vectora",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Handler{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 27, B: 27, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
