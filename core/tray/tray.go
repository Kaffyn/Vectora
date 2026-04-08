//go:build windows

package tray

import (
	"fmt"
	"os"
	"path/filepath"

	"vectora/core/llm"
	"vectora/core/telemetry"

	"github.com/getlantern/systray"
)

var (
	mStatus *systray.MenuItem
	mQuit   *systray.MenuItem

	activeProvider llm.LLMProvider
)

// Setup configura e inicia o systray do Windows.
func Setup(router *llm.Router) {
	if router != nil {
		activeProvider = router.GetDefault()
	}
	logger := telemetry.GetLogger()
	if logger != nil {
		logger.Info("Starting systray...")
	}
	// Log para ficheiro para debug
	logPath := filepath.Join(os.Getenv("USERPROFILE"), ".vectora", "logs", "systray-debug.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		fmt.Fprintf(f, "systray.Run called at %s\n", "now")
		f.Close()
	}
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("Vectora")
	systray.SetTooltip("Vectora - Agentic System")

	mStatus = systray.AddMenuItem("Running", "Vectora daemon is active")
	mStatus.Disable()

	systray.AddSeparator()
	mQuit = systray.AddMenuItem("Quit", "Shutdown Vectora")

	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

func onExit() {
	// Cleanup
}
