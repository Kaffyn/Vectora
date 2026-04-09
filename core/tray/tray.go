//go:build windows

package tray

import (
	"context"

	"github.com/Kaffyn/Vectora/assets"
	"github.com/Kaffyn/Vectora/core/i18n"
	"github.com/Kaffyn/Vectora/core/infra"
	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/getlantern/systray"
)

var (
	mStatus *systray.MenuItem
	mProv   *systray.MenuItem
	mGemini *systray.MenuItem
	mClaude *systray.MenuItem
	mLang   *systray.MenuItem
	mQuit   *systray.MenuItem

	mEn *systray.MenuItem
	mPt *systray.MenuItem
	mEs *systray.MenuItem
	mFr *systray.MenuItem

	ActiveProvider llm.Provider
)

// Setup configures and launches the systray.
// MVP: minimal tray — status, provider switch (Gemini/Claude), language, quit.
func Setup() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(assets.IconData)
	systray.SetTitle("Vectora")
	systray.SetTooltip("Vectora - Local AI Assistant (MVP)")

	// Status (informational only)
	mStatus = systray.AddMenuItem("", "")
	mStatus.Disable()
	systray.AddSeparator()

	// AI Provider selection (Gemini / Claude)
	mProv = systray.AddMenuItem("", "")
	mGemini = mProv.AddSubMenuItemCheckbox("", "", false)
	mClaude = mProv.AddSubMenuItemCheckbox("", "", false)
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
	} else if cfg.ClaudeAPIKey != "" {
		mClaude.Check()
		setProvider("claude", cfg.ClaudeAPIKey)
	} else {
		// Default to Gemini — user must configure API key
		mGemini.Check()
	}

	// Event loop
	go func() {
		for {
			select {
			case <-mGemini.ClickedCh:
				cfg := infra.LoadConfig()
				mGemini.Check()
				mClaude.Uncheck()
				setProvider("gemini", cfg.GeminiAPIKey)
			case <-mClaude.ClickedCh:
				cfg := infra.LoadConfig()
				mClaude.Check()
				mGemini.Uncheck()
				setProvider("claude", cfg.ClaudeAPIKey)
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
	// Shutdown previous provider if needed
	if ActiveProvider != nil && ActiveProvider.Name() == "claude" {
		if cProv, ok := ActiveProvider.(*llm.ClaudeProvider); ok {
			cProv.Close()
		}
	}

	if id == "gemini" {
		ctx := context.Background()
		prov, err := llm.NewGeminiProvider(ctx, secret)
		if err != nil {
			infra.NotifyOS("Vectora", "Gemini API key invalid or missing.")
			return
		}
		ActiveProvider = prov
		infra.NotifyOS("Vectora", "Gemini provider activated.")
	} else if id == "claude" {
		ctx := context.Background()
		prov, err := llm.NewClaudeProvider(ctx, secret)
		if err != nil {
			infra.NotifyOS("Vectora", "Claude API key invalid or missing.")
			return
		}
		ActiveProvider = prov
		infra.NotifyOS("Vectora", "Claude provider activated.")
	}
}

func updateLabels() {
	mStatus.SetTitle(i18n.T("tray_status"))
	mProv.SetTitle(i18n.T("tray_provider"))
	mGemini.SetTitle(i18n.T("tray_prov_gemini"))
	mClaude.SetTitle(i18n.T("tray_prov_claude"))
	mLang.SetTitle(i18n.T("tray_language"))
	mQuit.SetTitle(i18n.T("tray_quit"))
}

func onExit() {
	if ActiveProvider != nil && ActiveProvider.Name() == "claude" {
		if cProv, ok := ActiveProvider.(*llm.ClaudeProvider); ok {
			cProv.Close()
		}
	}
	if ActiveProvider != nil && ActiveProvider.Name() == "gemini" {
		// Gemini client cleanup if needed
	}
	if infra.Logger() != nil {
		infra.Logger().Info("Daemon shutting down...")
	}
}
