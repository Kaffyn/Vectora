//go:build windows

package tray

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/Kaffyn/Vectora/assets"
	"github.com/Kaffyn/Vectora/core/i18n"
	"github.com/Kaffyn/Vectora/core/infra"
	"github.com/Kaffyn/Vectora/core/llm"
	vecos "github.com/Kaffyn/Vectora/core/os"
	"github.com/getlantern/systray"
)

var (
	mStatus *systray.MenuItem
	mApp    *systray.MenuItem
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

	ActiveProvider llm.Provider
)

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

	mApp = systray.AddMenuItem("", "")
	mCli = systray.AddMenuItem("", "")
	systray.AddSeparator()

	mProv = systray.AddMenuItem("", "")
	mGemini = mProv.AddSubMenuItemCheckbox("", "", false)
	mQwen = mProv.AddSubMenuItemCheckbox("", "", false)
	systray.AddSeparator()

	mLang = systray.AddMenuItem("", "")
	mEn = mLang.AddSubMenuItemCheckbox("English", "", false)
	mPt = mLang.AddSubMenuItemCheckbox("Português", "", false)
	mEs = mLang.AddSubMenuItemCheckbox("Español", "", false)
	mFr = mLang.AddSubMenuItemCheckbox("Français", "", false)

	switch i18n.GetCurrentLang() {
	case "en":
		mEn.Check()
	case "pt":
		mPt.Check()
	case "es":
		mEs.Check()
	case "fr":
		mFr.Check()
	}

	systray.AddSeparator()
	mSet = systray.AddMenuItem("", "")
	mQuit = systray.AddMenuItem("", "")

	updateLabels()

	// Initial config check
	cfg := infra.LoadConfig()
	if cfg.GeminiAPIKey != "" {
		mGemini.Check()
		go switchProvider("gemini", cfg.GeminiAPIKey)
	} else {
		mQwen.Check()
		go switchProvider("qwen", "")
	}

	go func() {
		for {
			select {
			case <-mApp.ClickedCh:
				if infra.Logger != nil {
					infra.Logger.Info("Open Vectora Desktop Application")
				}
			case <-mCli.ClickedCh:
				openTerminal()
			case <-mGemini.ClickedCh:
				cfg := infra.LoadConfig()
				mGemini.Check()
				mQwen.Uncheck()
				switchProvider("gemini", cfg.GeminiAPIKey)
			case <-mQwen.ClickedCh:
				mQwen.Check()
				mGemini.Uncheck()
				switchProvider("qwen", "")
			case <-mEn.ClickedCh:
				mEn.Check()
				mPt.Uncheck()
				mEs.Uncheck()
				mFr.Uncheck()
				i18n.SetLanguage("en")
				updateLabels()
			case <-mPt.ClickedCh:
				mPt.Check()
				mEn.Uncheck()
				mEs.Uncheck()
				mFr.Uncheck()
				i18n.SetLanguage("pt")
				updateLabels()
			case <-mEs.ClickedCh:
				mEs.Check()
				mEn.Uncheck()
				mPt.Uncheck()
				mFr.Uncheck()
				i18n.SetLanguage("es")
				updateLabels()
			case <-mFr.ClickedCh:
				mFr.Check()
				mEn.Uncheck()
				mPt.Uncheck()
				mEs.Uncheck()
				i18n.SetLanguage("fr")
				updateLabels()
			case <-mSet.ClickedCh:
				if infra.Logger != nil {
					infra.Logger.Info("Open Settings")
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

func switchProvider(id, secret string) {
	if ActiveProvider != nil && ActiveProvider.Name() == "qwen" {
		if qProv, ok := ActiveProvider.(*llm.QwenProvider); ok {
			qProv.Shutdown()
		}
	}

	if id == "gemini" {
		infra.Logger.Info("Initializing Gemini provider...")
		ctx := context.Background()
		prov, err := llm.NewGeminiProvider(ctx, secret)
		if err != nil {
			infra.NotifyOS("Vectora AI Setup", "Alert: Verify Google API Key Configuration.")
			return
		}
		if secret != "" {
			infra.SaveConfig(&infra.Config{GeminiAPIKey: secret})
		}
		ActiveProvider = prov
		infra.NotifyOS("Vectora Engine", "Gemini connected.")
	} else if id == "qwen" {
		infra.Logger.Info("Starting Qwen local engine...")
		osMgr, _ := vecos.NewManager()
		base, _ := osMgr.GetAppDataDir()
		binPath := filepath.Join(base, "llama-server")
		modelPath := filepath.Join(base, "qwen.gguf")

		ctx := context.Background()
		prov, err := llm.NewQwenProvider(ctx, binPath, modelPath)
		if err != nil {
			msg := "Llama.cpp unavailable or GGUF not found in AppData."
			infra.Logger.Warn(msg, "err", err)
			infra.NotifyOS("Local Failure", msg)
			return
		}
		ActiveProvider = prov
		infra.NotifyOS("Vectora Engine", "Qwen GGUF activated.")
	}
}

func updateLabels() {
	mStatus.SetTitle(i18n.T("tray_status"))
	mApp.SetTitle(i18n.T("tray_open_app"))
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
		cmd = exec.Command("cmd", "/c", "start", "cmd", "/k", "vectora")
	}
	if cmd != nil {
		cmd.Start()
	}
}

func onExit() {
	if ActiveProvider != nil && ActiveProvider.Name() == "qwen" {
		if qProv, ok := ActiveProvider.(*llm.QwenProvider); ok {
			qProv.Shutdown()
		}
	}
	if infra.Logger != nil {
		infra.Logger.Info("Daemon shutting down...")
	}
}
