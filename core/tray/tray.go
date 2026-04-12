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
	mModel        *systray.MenuItem
	modelItems    map[string]*systray.MenuItem // Dynamic model menu items
	mLang         *systray.MenuItem
	mQuit         *systray.MenuItem

	mEn *systray.MenuItem
	mPt *systray.MenuItem
	mEs *systray.MenuItem
	mFr *systray.MenuItem

	ActiveProvider   llm.Provider
	ActiveProviderID string
	ActiveModel      string
)

// ProviderModels defines available models per provider (AGENTS.md April 2026)
var ProviderModels = map[string][]string{
	"gemini":     {"gemini-3.1-pro-preview", "gemini-3-flash-preview", "gemma-4-31b"},
	"claude":     {"claude-4.6-sonnet", "claude-4.6-opus", "claude-4.5-haiku"},
	"openai":     {"gpt-5.4-pro", "gpt-5.4-mini", "gpt-5-o1"},
	"openrouter": {"google/gemini-3.1-pro", "anthropic/claude-4.6-sonnet", "meta-llama/llama-4-70b"},
	"anannas":    {"anthropic/claude-4.6-sonnet", "google/gemini-3.1-pro", "openai/gpt-5.4-pro"},
	"deepseek":   {"deepseek-v3.2", "deepseek-v3.2-speciale"},
	"mistral":    {"mistral-large-3", "mistral-small-4"},
	"grok":       {"grok-4.20", "grok-4.1"},
	"zhipu":      {"glm-5.1", "glm-5-flash"},
}

// ReloadActiveProvider re-reads the .env configuration from disk and re-initializes
// the active LLM provider. This is critical for picking up API keys set by the CLI
// without requiring a full daemon restart.
func ReloadActiveProvider() {
	cfg := infra.LoadConfig()

	// Load active model from config
	ActiveModel = cfg.GeminiFallbackModel // Default to Gemini fallback if nothing else
	if cfg.ActiveModel != "" {
		ActiveModel = cfg.ActiveModel
	}

	// If we already have an active provider ID, try to refresh it
	if ActiveProviderID != "" {
		for _, prov := range AllProviders {
			if prov.ID == ActiveProviderID {
				key := prov.GetKey(cfg)
				if key != "" {
					setProvider(prov, key, cfg)
					updateLabels()
					return
				}
			}
		}
	}

	// Otherwise, fallback to DEFAULT_PROVIDER
	if cfg.DefaultProvider != "" {
		for _, prov := range AllProviders {
			if prov.ID == cfg.DefaultProvider {
				key := prov.GetKey(cfg)
				if key != "" {
					ActiveProviderID = prov.ID
					setProvider(prov, key, cfg)
					updateLabels()
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
			updateLabels()
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

	// AI Provider selection
	mProv = systray.AddMenuItem("", "")
	providerItems = make(map[string]*systray.MenuItem)

	for _, prov := range AllProviders {
		item := mProv.AddSubMenuItemCheckbox("", "", prov.ID == ActiveProviderID)
		providerItems[prov.ID] = item
	}

	// Model selection (Dynamic based on provider)
	mModel = systray.AddMenuItem("Modelo", "")
	modelItems = make(map[string]*systray.MenuItem)
	// We'll populate this on demand or pre-populate all and hide/show
	for _, provModels := range ProviderModels {
		for _, model := range provModels {
			if _, exists := modelItems[model]; !exists {
				item := mModel.AddSubMenuItemCheckbox(model, "", false)
				item.Hide()
				modelItems[model] = item
			}
		}
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
				for id, item := range providerItems {
					select {
					case <-item.ClickedCh:
						// Find the provider info
						var selectedProv ProviderInfo
						for _, p := range AllProviders {
							if p.ID == id {
								selectedProv = p
								break
							}
						}

						ActiveProviderID = id
						cfg.DefaultProvider = id
						infra.SaveConfig(cfg)
						
						key := selectedProv.GetKey(cfg)
						setProvider(selectedProv, key, cfg)
						updateLabels()
					default:
					}
				}

				// Dynamic model selection
				for model, item := range modelItems {
					select {
					case <-item.ClickedCh:
						ActiveModel = model
						cfg.ActiveModel = model
						cfg.DefaultModel = model
						infra.SaveConfig(cfg)
						updateLabels()
						infra.NotifyOS("Vectora", "Model "+model+" selected.")
					default:
					}
				}
			}
		}
	}()
}

func setProvider(prov ProviderInfo, secret string, cfg *infra.Config) {
	if secret == "" {
		infra.NotifyOS("Vectora", "API key for "+prov.ID+" not configured.")
		return
	}

	ctx := context.Background()
	p, err := prov.Setup(ctx, secret, cfg)
	if err != nil {
		infra.NotifyOS("Vectora", "Failed to initialize "+prov.ID+" provider: "+err.Error())
		return
	}

	ActiveProvider = p
}

func updateLabels() {
	mStatus.SetTitle(i18n.T("tray_status"))
	mProv.SetTitle(i18n.T("tray_provider"))
	mModel.SetTitle("Modelo")

	for id, item := range providerItems {
		item.SetTitle(id) // Optional: use i18n
		if id == ActiveProviderID {
			item.Check()
		} else {
			item.Uncheck()
		}
	}

	// Update models visibility and checks
	validModels := ProviderModels[ActiveProviderID]
	for m, item := range modelItems {
		isVisible := false
		for _, vm := range validModels {
			if vm == m {
				isVisible = true
				break
			}
		}

		if isVisible {
			item.Show()
			if m == ActiveModel {
				item.Check()
			} else {
				item.Uncheck()
			}
		} else {
			item.Hide()
		}
	}

	mLang.SetTitle(i18n.T("tray_language"))
	mQuit.SetTitle(i18n.T("tray_quit"))
}

func onExit() {
	if infra.Logger() != nil {
		infra.Logger().Info("Core shutting down...")
	}
}
