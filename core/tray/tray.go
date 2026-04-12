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

// ProviderInfo describes an LLM provider available in the tray menu
type ProviderInfo struct {
	ID      string                                                             // Internal identifier (gemini, claude, openai, etc.)
	Label   string                                                             // Display label (from i18n key)
	I18nKey string                                                             // i18n translation key
	GetKey  func(*infra.Config) string                                         // Function to get API key from config
	Setup   func(context.Context, string, *infra.Config) (llm.Provider, error) // Constructor
}

// AllProviders defines all 10+ LLM families supported by Vectora (AGENTS.md April 2026)
var AllProviders = []ProviderInfo{
	{
		ID:      "gemini",
		I18nKey: "tray_prov_gemini",
		GetKey: func(cfg *infra.Config) string {
			return cfg.GeminiAPIKey
		},
		Setup: func(ctx context.Context, key string, cfg *infra.Config) (llm.Provider, error) {
			return llm.NewGeminiProvider(ctx, key)
		},
	},
	{
		ID:      "claude",
		I18nKey: "tray_prov_claude",
		GetKey: func(cfg *infra.Config) string {
			return cfg.ClaudeAPIKey
		},
		Setup: func(ctx context.Context, key string, cfg *infra.Config) (llm.Provider, error) {
			return llm.NewClaudeProvider(ctx, key)
		},
	},
	{
		ID:      "openai",
		I18nKey: "tray_prov_openai",
		GetKey: func(cfg *infra.Config) string {
			return cfg.OpenAIAPIKey
		},
		Setup: func(ctx context.Context, key string, cfg *infra.Config) (llm.Provider, error) {
			return llm.NewOpenAIProvider(key, cfg.OpenAIBaseURL, "openai"), nil
		},
	},
	{
		ID:      "openrouter",
		I18nKey: "tray_prov_openrouter",
		GetKey: func(cfg *infra.Config) string {
			return cfg.OpenRouterAPIKey
		},
		Setup: func(ctx context.Context, key string, cfg *infra.Config) (llm.Provider, error) {
			return llm.NewGatewayProvider(key, "https://openrouter.ai/api/v1", "openrouter"), nil
		},
	},
	{
		ID:      "anannas",
		I18nKey: "tray_prov_anannas",
		GetKey: func(cfg *infra.Config) string {
			return cfg.AnannasAPIKey
		},
		Setup: func(ctx context.Context, key string, cfg *infra.Config) (llm.Provider, error) {
			return llm.NewGatewayProvider(key, "https://api.anannas.ai/v1", "anannas"), nil
		},
	},
	{
		ID:      "deepseek",
		I18nKey: "tray_prov_deepseek",
		GetKey: func(cfg *infra.Config) string {
			return cfg.DeepSeekAPIKey
		},
		Setup: func(ctx context.Context, key string, cfg *infra.Config) (llm.Provider, error) {
			baseURL := cfg.DeepSeekBaseURL
			if baseURL == "" {
				baseURL = "https://api.deepseek.com/v1"
			}
			return llm.NewOpenAIProvider(key, baseURL, "deepseek"), nil
		},
	},
	{
		ID:      "mistral",
		I18nKey: "tray_prov_mistral",
		GetKey: func(cfg *infra.Config) string {
			return cfg.MistralAPIKey
		},
		Setup: func(ctx context.Context, key string, cfg *infra.Config) (llm.Provider, error) {
			baseURL := cfg.MistralBaseURL
			if baseURL == "" {
				baseURL = "https://api.mistral.ai/v1"
			}
			return llm.NewOpenAIProvider(key, baseURL, "mistral"), nil
		},
	},
	{
		ID:      "grok",
		I18nKey: "tray_prov_grok",
		GetKey: func(cfg *infra.Config) string {
			return cfg.GrokAPIKey
		},
		Setup: func(ctx context.Context, key string, cfg *infra.Config) (llm.Provider, error) {
			baseURL := cfg.GrokBaseURL
			if baseURL == "" {
				baseURL = "https://api.x.ai/v1"
			}
			return llm.NewOpenAIProvider(key, baseURL, "grok"), nil
		},
	},
	{
		ID:      "zhipu",
		I18nKey: "tray_prov_zhipu",
		GetKey: func(cfg *infra.Config) string {
			return cfg.ZhipuAPIKey
		},
		Setup: func(ctx context.Context, key string, cfg *infra.Config) (llm.Provider, error) {
			baseURL := cfg.ZhipuBaseURL
			if baseURL == "" {
				baseURL = "https://open.bigmodel.cn/api/paas/v4"
			}
			return llm.NewOpenAIProvider(key, baseURL, "zhipu"), nil
		},
	},
}

var (
	mStatus       *systray.MenuItem
	mProv         *systray.MenuItem
	providerItems map[string]*systray.MenuItem // Dynamic provider menu items
	mLang         *systray.MenuItem
	mQuit         *systray.MenuItem

	mEn *systray.MenuItem
	mPt *systray.MenuItem
	mEs *systray.MenuItem
	mFr *systray.MenuItem

	ActiveProvider   llm.Provider
	ActiveProviderID string
)

// ReloadActiveProvider re-reads the .env configuration from disk and re-initializes
// the active LLM provider. This is critical for picking up API keys set by the CLI
// without requiring a full daemon restart.
func ReloadActiveProvider() {
	cfg := infra.LoadConfig()

	// If we already have an active provider ID, try to refresh it
	if ActiveProviderID != "" {
		for _, prov := range AllProviders {
			if prov.ID == ActiveProviderID {
				key := prov.GetKey(cfg)
				if key != "" {
					setProvider(prov, key, cfg)
					return
				}
			}
		}
	}

	// Otherwise, fallback to DEFAULT_PROVIDER or first available (same logic as onReady)
	if cfg.DefaultProvider != "" {
		for _, prov := range AllProviders {
			if prov.ID == cfg.DefaultProvider {
				key := prov.GetKey(cfg)
				if key != "" {
					ActiveProviderID = prov.ID
					setProvider(prov, key, cfg)
					return
				}
			}
		}
	}

	for _, prov := range AllProviders {
		key := prov.GetKey(cfg)
		if key != "" {
			ActiveProviderID = prov.ID
			setProvider(prov, key, cfg)
			return
		}
	}
}

// Setup configures and launches the systray.
func Setup() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(assets.IconData)
	systray.SetTitle("Vectora")
	systray.SetTooltip("Vectora - AI Assistant")

	// Status (informational only)
	mStatus = systray.AddMenuItem("", "")
	mStatus.Disable()
	systray.AddSeparator()

	// AI Provider selection (dynamic list of all available providers)
	mProv = systray.AddMenuItem("", "")
	providerItems = make(map[string]*systray.MenuItem)

	for _, prov := range AllProviders {
		item := mProv.AddSubMenuItemCheckbox("", "", false)
		providerItems[prov.ID] = item
	}
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
	ReloadActiveProvider()

	// Event loop
	go func() {
		for {
			select {
			// Provider selection
			case <-mQuit.ClickedCh:
				systray.Quit()
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

			// Dynamic provider selection
			default:
				cfg := infra.LoadConfig()
				for _, prov := range AllProviders {
					select {
					case <-providerItems[prov.ID].ClickedCh:
						// Uncheck all others, check this one
						for _, p := range AllProviders {
							if p.ID == prov.ID {
								providerItems[p.ID].Check()
							} else {
								providerItems[p.ID].Uncheck()
							}
						}
						ActiveProviderID = prov.ID
						key := prov.GetKey(cfg)
						setProvider(prov, key, cfg)
					default:
					}
				}
			}
		}
	}()
}

func setProvider(prov ProviderInfo, secret string, cfg *infra.Config) {
	if secret == "" {
		infra.NotifyOS("Vectora", "API key for "+prov.I18nKey+" not configured.")
		return
	}

	ctx := context.Background()
	p, err := prov.Setup(ctx, secret, cfg)
	if err != nil {
		infra.NotifyOS("Vectora", "Failed to initialize "+prov.ID+" provider: "+err.Error())
		return
	}

	ActiveProvider = p
	infra.NotifyOS("Vectora", prov.ID+" provider activated.")
}

func updateLabels() {
	mStatus.SetTitle(i18n.T("tray_status"))
	mProv.SetTitle(i18n.T("tray_provider"))

	for _, prov := range AllProviders {
		providerItems[prov.ID].SetTitle(i18n.T(prov.I18nKey))
	}

	mLang.SetTitle(i18n.T("tray_language"))
	mQuit.SetTitle(i18n.T("tray_quit"))
}

func onExit() {
	if infra.Logger() != nil {
		infra.Logger().Info("Core shutting down...")
	}
}
