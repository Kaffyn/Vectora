//go:build windows

package tray

import (
	"context"
	"path/filepath"

	"github.com/Kaffyn/Vectora/assets"
	"github.com/Kaffyn/Vectora/core/i18n"
	"github.com/Kaffyn/Vectora/core/infra"
	"github.com/Kaffyn/Vectora/core/llm"
	vecos "github.com/Kaffyn/Vectora/core/os"
	"github.com/getlantern/systray"
)

var (
	mStatus *systray.MenuItem
	mProv   *systray.MenuItem
	mGemini *systray.MenuItem
	mQwen   *systray.MenuItem
	mLang   *systray.MenuItem
	mQuit   *systray.MenuItem

	mEn *systray.MenuItem
	mPt *systray.MenuItem
	mEs *systray.MenuItem
	mFr *systray.MenuItem

	ActiveProvider llm.Provider
)

// Setup configures and launches the systray.
// MVP: minimal tray — status, provider switch, language, quit.
// No Desktop app, no CLI launcher, no settings dialog.
func Setup() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(assets.IconData)
	systray.SetTitle("Vectora")
	systray.SetTooltip("Vectora - Local AI Assistant (MVP)")

	// Status (disabled, informational only)
	mStatus = systray.AddMenuItem("", "")
	mStatus.Disable()
	systray.AddSeparator()

	// AI Provider selection
	mProv = systray.AddMenuItem("", "")
	mGemini = mProv.AddSubMenuItemCheckbox("", "", false)
	mQwen = mProv.AddSubMenuItemCheckbox("", "", false)
	systray.AddSeparator()

	// Language selection
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
	mQuit = systray.AddMenuItem("", "")

	updateLabels()

	// Initial config check for default provider
	cfg := infra.LoadConfig()
	if cfg.GeminiAPIKey != "" {
		mGemini.Check()
		setProvider("gemini", cfg.GeminiAPIKey)
	} else {
		mQwen.Check()
		setProvider("qwen", "")
	}

	// Event loop
	go func() {
		for {
			select {
			case <-mGemini.ClickedCh:
				cfg := infra.LoadConfig()
				mGemini.Check()
				mQwen.Uncheck()
				setProvider("gemini", cfg.GeminiAPIKey)
			case <-mQwen.ClickedCh:
				mQwen.Check()
				mGemini.Uncheck()
				setProvider("qwen", "")
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
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

func setProvider(id, secret string) {
	if ActiveProvider != nil && ActiveProvider.Name() == "qwen" {
		if qProv, ok := ActiveProvider.(*llm.QwenProvider); ok {
			qProv.Shutdown()
		}
	}

	if id == "gemini" {
		ctx := context.Background()
		prov, err := llm.NewGeminiProvider(ctx, secret)
		if err != nil {
			infra.NotifyOS("Vectora", "Gemini API key invalid or missing.")
			return
		}
		if secret != "" {
			infra.SaveConfig(&infra.Config{GeminiAPIKey: secret})
		}
		ActiveProvider = prov
		infra.NotifyOS("Vectora", "Gemini provider activated.")
	} else if id == "qwen" {
		// Qwen local requires llama.cpp binary + GGUF model
		// In MVP, this is optional — user must have llama-server + qwen.gguf in AppData
		osMgr, _ := vecos.NewManager()
		base, _ := osMgr.GetAppDataDir()
		binPath := filepath.Join(base, "llama-server")
		modelPath := filepath.Join(base, "qwen.gguf")

		ctx := context.Background()
		prov, err := llm.NewQwenProvider(ctx, binPath, modelPath)
		if err != nil {
			infra.NotifyOS("Vectora", "Qwen local: llama.cpp or model not found in AppData.")
			return
		}
		ActiveProvider = prov
		infra.NotifyOS("Vectora", "Qwen local provider activated.")
	}
}

func updateLabels() {
	mStatus.SetTitle(i18n.T("tray_status"))
	mProv.SetTitle(i18n.T("tray_provider"))
	mGemini.SetTitle(i18n.T("tray_prov_gemini"))
	mQwen.SetTitle(i18n.T("tray_prov_qwen"))
	mLang.SetTitle(i18n.T("tray_language"))
	mQuit.SetTitle(i18n.T("tray_quit"))
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
