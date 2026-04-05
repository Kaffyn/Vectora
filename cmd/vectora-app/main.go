package main

import (
	"embed"

	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	vecos "github.com/Kaffyn/Vectora/internal/os"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:internal/app/out
var assets embed.FS

func main() {
	systemManager, _ := vecos.NewManager()
	if systemManager != nil && !systemManager.IsRunningAsAdmin() {
		// Attempt to restart as admin
		exe, _ := os.Executable()
		cwd, _ := os.Getwd()
		args := strings.Join(os.Args[1:], " ")

		psCmd := fmt.Sprintf("Start-Process -FilePath '%s' -Verb runas -WorkingDirectory '%s'", exe, cwd)
		if args != "" {
			psCmd += fmt.Sprintf(" -ArgumentList '%s'", args)
		}

		cmd := exec.Command("powershell", psCmd)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if err := cmd.Start(); err == nil {
			os.Exit(0)
		}
	}

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Vectora Desktop",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 9, G: 9, B: 11, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
