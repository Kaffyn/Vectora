package tray

import (
	"github.com/getlantern/systray"
	"github.com/Kaffyn/vectora/internal/infra"
)

// Setup configures and launches the systray. Blockingly executed.
func Setup() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("Vectora")
	systray.SetTooltip("Vectora - Agentic System")
	
	mStatus := systray.AddMenuItem("Daemon Running", "Daemon status")
	mStatus.Disable()

	systray.AddSeparator()
	
	mQuit := systray.AddMenuItem("Quit", "Shutdown Vectora")
	
	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				if infra.Logger != nil {
					infra.Logger.Info("Quit requested from tray")
				}
				systray.Quit()
			}
		}
	}()
}

func onExit() {
	if infra.Logger != nil {
		infra.Logger.Info("Daemon shutting down...")
	}
}
