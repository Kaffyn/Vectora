package tray

import (
	"os/exec"
	"runtime"

	"github.com/Kaffyn/vectora/assets"
	"github.com/Kaffyn/vectora/internal/i18n"
	"github.com/Kaffyn/vectora/internal/infra"
	"github.com/getlantern/systray"
)

var (
	mStatus *systray.MenuItem
	mWeb    *systray.MenuItem
	mCli    *systray.MenuItem
	mProv   *systray.MenuItem
	mGemini *systray.MenuItem
	mQwen   *systray.MenuItem
	mLang   *systray.MenuItem
	mSet    *systray.MenuItem
	mQuit   *systray.MenuItem

	mEn *systray.MenuItem
	mPt *systray.MenuItem
	mEs *systray.MenuItem
	mFr *systray.MenuItem
)

// Setup configures and launches the systray.
func Setup() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(assets.IconData)
	systray.SetTitle("Vectora")
	systray.SetTooltip("Vectora - Agentic System")

	mStatus = systray.AddMenuItem("", "")
	mStatus.Disable()
	systray.AddSeparator()

	mWeb = systray.AddMenuItem("", "")
	mCli = systray.AddMenuItem("", "")
	systray.AddSeparator()

	mProv = systray.AddMenuItem("", "")
	mGemini = mProv.AddSubMenuItemCheckbox("", "", true)
	mQwen = mProv.AddSubMenuItemCheckbox("", "", false)
	systray.AddSeparator()

	mLang = systray.AddMenuItem("", "")
	mEn = mLang.AddSubMenuItemCheckbox("English", "", false)
	mPt = mLang.AddSubMenuItemCheckbox("Português", "", false)
	mEs = mLang.AddSubMenuItemCheckbox("Español", "", false)
	mFr = mLang.AddSubMenuItemCheckbox("Français", "", false)
	
	switch i18n.GetCurrentLang() {
	case "en": mEn.Check()
	case "pt": mPt.Check()
	case "es": mEs.Check()
	case "fr": mFr.Check()
	}

	systray.AddSeparator()
	mSet = systray.AddMenuItem("", "")
	mQuit = systray.AddMenuItem("", "")

	updateLabels()

	// Event Loop
	go func() {
		for {
			select {
			case <-mWeb.ClickedCh:
				if infra.Logger != nil { infra.Logger.Info("Open Web UI") }
			case <-mCli.ClickedCh:
				if infra.Logger != nil { infra.Logger.Info("Open CLI") }
				openTerminal()
			case <-mGemini.ClickedCh:
				mGemini.Check(); mQwen.Uncheck()
				if infra.Logger != nil { infra.Logger.Info("Switched to Gemini") }
			case <-mQwen.ClickedCh:
				mQwen.Check(); mGemini.Uncheck()
				if infra.Logger != nil { infra.Logger.Info("Starting Qwen3 llama.cpp subprocess...") }
			case <-mEn.ClickedCh:
				mEn.Check(); mPt.Uncheck(); mEs.Uncheck(); mFr.Uncheck()
				i18n.SetLanguage("en"); updateLabels()
			case <-mPt.ClickedCh:
				mPt.Check(); mEn.Uncheck(); mEs.Uncheck(); mFr.Uncheck()
				i18n.SetLanguage("pt"); updateLabels()
			case <-mEs.ClickedCh:
				mEs.Check(); mEn.Uncheck(); mPt.Uncheck(); mFr.Uncheck()
				i18n.SetLanguage("es"); updateLabels()
			case <-mFr.ClickedCh:
				mFr.Check(); mEn.Uncheck(); mPt.Uncheck(); mFr.Uncheck()
				i18n.SetLanguage("fr"); updateLabels()
			case <-mSet.ClickedCh:
				if infra.Logger != nil { infra.Logger.Info("Open Settings (Abre app Wails na view Settings)") }
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

func updateLabels() {
	mStatus.SetTitle(i18n.T("tray_status"))
	mWeb.SetTitle(i18n.T("tray_open_web"))
	mCli.SetTitle(i18n.T("tray_open_cli"))
	
	mProv.SetTitle(i18n.T("tray_provider"))
	mGemini.SetTitle(i18n.T("tray_prov_gemini"))
	mQwen.SetTitle(i18n.T("tray_prov_qwen"))
	
	mLang.SetTitle(i18n.T("tray_language"))
	mSet.SetTitle(i18n.T("tray_settings"))
	mQuit.SetTitle(i18n.T("tray_quit"))
}

func openTerminal() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "start", "vectora-cli.exe")
	}
	if cmd != nil {
		cmd.Start()
	}
}

func onExit() {
	if infra.Logger != nil {
		infra.Logger.Info("Daemon shutting down...")
	}
}
