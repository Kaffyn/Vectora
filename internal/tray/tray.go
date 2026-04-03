package tray

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/Kaffyn/vectora/assets"
	"github.com/Kaffyn/vectora/internal/i18n"
	"github.com/Kaffyn/vectora/internal/infra"
	"github.com/Kaffyn/vectora/internal/llm"
	vecos "github.com/Kaffyn/vectora/internal/os"
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

	ActiveProvider llm.Provider
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
	mGemini = mProv.AddSubMenuItemCheckbox("", "", false)
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

	// Initial Config Check for Default LLM Provider
	cfg := infra.LoadConfig()
	if cfg.GeminiAPIKey != "" {
		mGemini.Check()
		go switchProvider("gemini", cfg.GeminiAPIKey)
	} else {
		mQwen.Check()
		go switchProvider("qwen", "")
	}

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
				switchProvider("gemini", cfg.GeminiAPIKey)
			case <-mQwen.ClickedCh:
				mQwen.Check(); mGemini.Uncheck()
				switchProvider("qwen", "")
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
				if infra.Logger != nil { infra.Logger.Info("Open Settings (Opens Wails app in Settings view)") }
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

func switchProvider(id, secret string) {
	// Cleanup before switching
	if ActiveProvider != nil && ActiveProvider.Name() == "qwen" {
		if qProv, ok := ActiveProvider.(*llm.QwenProvider); ok {
			qProv.Shutdown() // Terminate llama.cpp in the old provider
		}
	}

	if id == "gemini" {
		infra.Logger.Info("Initializing langchaingo binding with Gemini API...")
		prov, err := llm.NewGeminiProvider(context.Background(), secret)
		if err != nil {
			infra.NotifyOS("Vectora AI Setup", "Alert: Verify Google API Key Configuration.")
			return
		}
		
		// Persist key if login is successful
		if secret != "" {
			infra.SaveConfig(&infra.Config{GeminiAPIKey: secret})
		}

		ActiveProvider = prov
		infra.NotifyOS("Vectora Engine", "LangChain visually connected to Cloud Processor (Google).")
	} else if id == "qwen" {
		infra.Logger.Info("Waking Kernel for llama.cpp Local Sidecar...")
		
		// Search for qwen in the local models home.
		osMgr, _ := vecos.NewManager()
		base, _ := osMgr.GetAppDataDir()
		modelPath := filepath.Join(base, "qwen.gguf")

		prov, err := llm.NewQwenProvider(context.Background(), modelPath)
		if err != nil {
			msg := "Llama.cpp unavailable (Motor sidecar or GGUF 'qwen.gguf' not found in AppData)."
			infra.Logger.Warn(msg, "err", err)
			infra.NotifyOS("Local Failure", msg)
			return
		}
		ActiveProvider = prov
		infra.NotifyOS("Vectora Engine", "Native AI Engine (Qwen GGUF) activated successfully via dynamic pipes!")
	}
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
	if ActiveProvider != nil && ActiveProvider.Name() == "qwen" {
		if qProv, ok := ActiveProvider.(*llm.QwenProvider); ok {
			qProv.Shutdown()
		}
	}
	if infra.Logger != nil {
		infra.Logger.Info("Daemon shutting down...")
	}
}
