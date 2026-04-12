package main

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Kaffyn/Vectora/core/infra"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Vectora core configurations",
	Long:  "Read or modify global configuration keys stored in official OS locations",
	Run: func(cmd *cobra.Command, args []string) {
		runConfigInteractive()
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get [KEY]",
	Short: "Get a configuration key value",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		envPath := getConfigPath()
		envMap, _ := godotenv.Read(envPath)
		if val, ok := envMap[key]; ok {
			displayVal := val
			if len(val) > 20 && stringsContainsSecret(key) {
				displayVal = val[:10] + "..." + val[len(val)-10:]
			}
			fmt.Printf("%s=%s\n", key, displayVal)
		} else {
			fmt.Printf("Key '%s' not found in configuration.\n", key)
		}
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration keys",
	Run: func(cmd *cobra.Command, args []string) {
		envPath := getConfigPath()
		envMap, err := godotenv.Read(envPath)
		if err != nil {
			fmt.Println("No configuration found or error reading it.")
			return
		}

		fmt.Println("Configuration Keys:")
		for k, v := range envMap {
			displayVal := v
			if len(v) > 20 && stringsContainsSecret(k) {
				displayVal = v[:10] + "..." + v[len(v)-10:]
			}
			fmt.Printf("  %s=%s\n", k, displayVal)
		}
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [KEY] [VALUE]",
	Short: "Set a configuration key",
	Long: `Set a configuration key in the global .env file.
Supported keys (among others):
  GEMINI_API_KEY           API key for Google Gemini
  ANTHROPIC_API_KEY        API key for Anthropic Claude
  OPENAI_API_KEY           API key for OpenAI
  DEFAULT_PROVIDER         Default AI provider (gemini, claude, openai, qwen, openrouter)
  DEFAULT_MODEL            Override the default model
  DEFAULT_FALLBACK_PROVIDER The provider to use if primary fails
  GEMINI_FALLBACK_MODEL    Fallback model for Gemini
  CLAUDE_FALLBACK_MODEL    Fallback model for Claude
  OPENAI_FALLBACK_MODEL    Fallback model for OpenAI
  LOG_LEVEL                Log verbosity (debug, info, warn, error)

Example: vectora config set GEMINI_API_KEY AIzaSy...`,
	Args: cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			runConfigInteractive()
			return
		}

		key := args[0]
		var value string
		if len(args) == 2 {
			value = args[1]
		} else {
			// Prompt for value if only key is provided
			prompt := &survey.Password{
				Message: fmt.Sprintf("Enter value for %s:", key),
			}
			survey.AskOne(prompt, &value)
		}

		// Validate key is not empty
		if key == "" {
			fmt.Println("Error: Key cannot be empty.")
			return
		}

		saveConfigValue(key, value)
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configSetCmd)
}

// Helpers for secret display
func stringsContainsSecret(s string) bool {
	s = stringsToLower(s)
	return stringsContains(s, "key") || stringsContains(s, "secret") || stringsContains(s, "token") || stringsContains(s, "pwd") || stringsContains(s, "password")
}

// Minimal string helpers since we removed 'strings' import to fix build (temp wait I'll restore it)
func stringsToLower(s string) string {
	res := ""
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			res += string(r + 32)
		} else {
			res += string(r)
		}
	}
	return res
}

func stringsContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func runConfigInteractive() {
	cfg := infra.LoadConfig()

	keyPreview := func(key string) string {
		if key == "" {
			return "not set"
		}
		prefixLen := 7
		if len(key) < prefixLen {
			prefixLen = len(key)
		}
		return key[:prefixLen] + "..."
	}

	// ── PASSO 1: Gateway ──────────────────────────────────────────────
	gatewayOptions := []string{}
	noneLabel := "Nenhum (direto)"
	if cfg.ActiveGateway == "" || cfg.ActiveGateway == "none" {
		noneLabel += " ★"
	}
	gatewayOptions = append(gatewayOptions, noneLabel)

	orLabel := fmt.Sprintf("OpenRouter [%s]", keyPreview(cfg.OpenRouterAPIKey))
	if cfg.ActiveGateway == "openrouter" {
		orLabel += " ★"
	}
	gatewayOptions = append(gatewayOptions, orLabel)

	anLabel := fmt.Sprintf("Anannas [%s]", keyPreview(cfg.AnannasAPIKey))
	if cfg.ActiveGateway == "anannas" {
		anLabel += " ★"
	}
	gatewayOptions = append(gatewayOptions, anLabel)
	gatewayOptions = append(gatewayOptions, "Cancelar")

	var gatewayChoice string
	if err := survey.AskOne(&survey.Select{
		Message: "1/3 — Gateway (camada de roteamento):",
		Options: gatewayOptions,
	}, &gatewayChoice); err != nil || gatewayChoice == "Cancelar" {
		return
	}

	switch {
	case stringsIndexOf(gatewayChoice, "Nenhum") != -1:
		saveEnvKey("ACTIVE_GATEWAY", "")
		fmt.Println("Gateway: desativado.")
	case stringsIndexOf(gatewayChoice, "OpenRouter") != -1:
		if cfg.OpenRouterAPIKey == "" {
			promptKeyAndSave("OPENROUTER_API_KEY", "OpenRouter API Key:", true)
		}
		saveEnvKey("ACTIVE_GATEWAY", "openrouter")
		fmt.Println("Gateway: OpenRouter ativo.")
	case stringsIndexOf(gatewayChoice, "Anannas") != -1:
		if cfg.AnannasAPIKey == "" {
			promptKeyAndSave("ANANNAS_API_KEY", "Anannas API Key:", true)
		}
		saveEnvKey("ACTIVE_GATEWAY", "anannas")
		fmt.Println("Gateway: Anannas ativo.")
	}

	// Recarrega cfg após salvar gateway
	cfg = infra.LoadConfig()

	// ── PASSO 2: Provider ─────────────────────────────────────────────
	type providerDef struct {
		label   string
		envKey  string
		baseURL string
		current string
	}
	providers := []providerDef{
		{"Google Gemini", "GEMINI_API_KEY", "", cfg.GeminiAPIKey},
		{"Anthropic Claude", "CLAUDE_API_KEY", "", cfg.ClaudeAPIKey},
		{"OpenAI", "OPENAI_API_KEY", "OPENAI_BASE_URL", cfg.OpenAIAPIKey},
		{"DeepSeek", "DEEPSEEK_API_KEY", "DEEPSEEK_BASE_URL", cfg.DeepSeekAPIKey},
		{"Mistral", "MISTRAL_API_KEY", "MISTRAL_BASE_URL", cfg.MistralAPIKey},
		{"xAI Grok", "GROK_API_KEY", "GROK_BASE_URL", cfg.GrokAPIKey},
		{"Alibaba Qwen", "QWEN_API_KEY", "QWEN_BASE_URL", cfg.QwenAPIKey},
		{"Zhipu GLM-5", "ZHIPU_API_KEY", "ZHIPU_BASE_URL", cfg.ZhipuAPIKey},
		{"Voyage AI (embedding only)", "VOYAGE_API_KEY", "", cfg.VoyageAPIKey},
	}

	providerLabels := []string{}
	for _, p := range providers {
		label := fmt.Sprintf("%s [%s]", p.label, keyPreview(p.current))
		providerLabels = append(providerLabels, label)
	}
	providerLabels = append(providerLabels, "Pular")

	var providerChoice string
	if err := survey.AskOne(&survey.Select{
		Message: "2/3 — Provider padrão:",
		Options: providerLabels,
	}, &providerChoice); err != nil || providerChoice == "Pular" {
		fmt.Println("Configuração salva.")
		return
	}

	// Encontra provider selecionado
	var selectedProvider *providerDef
	for i, label := range providerLabels {
		if label == providerChoice && i < len(providers) {
			p := providers[i]
			selectedProvider = &p
			break
		}
	}

	if selectedProvider != nil {
		if selectedProvider.current == "" {
			promptKeyAndSave(selectedProvider.envKey, selectedProvider.label+" API Key:", true)
			if selectedProvider.baseURL != "" {
				promptKeyAndSave(selectedProvider.baseURL, selectedProvider.label+" Base URL (opcional, deixe vazio para padrão):", false)
			}
		}

		// Deriva nome interno do provider para salvar DEFAULT_PROVIDER
		provInternal := map[string]string{
			"Google Gemini":              "gemini",
			"Anthropic Claude":           "claude",
			"OpenAI":                     "openai",
			"DeepSeek":                   "deepseek",
			"Mistral":                    "mistral",
			"xAI Grok":                   "grok",
			"Alibaba Qwen":               "qwen",
			"Zhipu GLM-5":                "zhipu",
			"Voyage AI (embedding only)": "voyage",
		}
		if internal, ok := provInternal[selectedProvider.label]; ok {
			saveEnvKey("DEFAULT_PROVIDER", internal)
			fmt.Printf("Provider padrão: %s\n", internal)
		}
	}

	// ── PASSO 3: Model ────────────────────────────────────────────────
	currentModel := ""
	envMap, _ := godotenv.Read(getConfigPath())
	if envMap != nil {
		currentModel = envMap["DEFAULT_MODEL"]
	}

	modelMsg := "3/3 — Model padrão (Enter para manter"
	if currentModel != "" {
		modelMsg += " '" + currentModel + "'"
	} else {
		modelMsg += " padrão do provider"
	}
	modelMsg += "):"

	var modelVal string
	if err := survey.AskOne(&survey.Input{Message: modelMsg}, &modelVal); err == nil && modelVal != "" {
		saveEnvKey("DEFAULT_MODEL", modelVal)
		fmt.Printf("Model padrão: %s\n", modelVal)
	}

	fmt.Println("\nConfiguração concluída.")
}

// saveEnvKey persiste uma chave no .env sem prompt interativo.
func saveEnvKey(key, value string) {
	envPath := getConfigPath()
	envMap, _ := godotenv.Read(envPath)
	if envMap == nil {
		envMap = make(map[string]string)
	}
	envMap[key] = value
	_ = godotenv.Write(envMap, envPath)
}

func stringsIndexOf(s, sep string) int {
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}

func promptKeyAndSave(key, message string, isSecret bool) {
	var val string
	var prompt survey.Prompt
	if isSecret {
		prompt = &survey.Password{Message: message}
	} else {
		prompt = &survey.Input{Message: message}
	}

	if err := survey.AskOne(prompt, &val); err != nil {
		return
	}
	if val != "" {
		saveConfigValue(key, val)
	}
}

func selectAndSave(key, message string, options []string) {
	var val string
	prompt := &survey.Select{
		Message: message,
		Options: options,
	}
	if err := survey.AskOne(prompt, &val); err != nil {
		return
	}
	saveConfigValue(key, val)
}

func saveConfigValue(key, value string) {
	envPath := getConfigPath()

	// Ensure directory exists
	dir := filepathDir(envPath)
	if _, err := osStat(dir); osIsNotExist(err) {
		_ = osMkdirAll(dir, 0755)
	}

	envMap := make(map[string]string)
	if _, err := osStat(envPath); err == nil {
		if m, err := godotenvRead(envPath); err == nil {
			envMap = m
		}
	}

	envMap[key] = value
	if err := godotenvWrite(envMap, envPath); err != nil {
		fmt.Println("Error writing .env:", err)
		return
	}

	fmt.Printf("\nKey %s has been set successfully.\n", key)
}

// Minimal path helpers to avoid import issues
func filepathDir(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' || p[i] == '\\' {
			return p[:i]
		}
	}
	return "."
}

func osStat(p string) (os.FileInfo, error) { return os.Stat(p) }
func osIsNotExist(err error) bool       { return os.IsNotExist(err) }
func osMkdirAll(p string, perm os.FileMode) error { return os.MkdirAll(p, perm) }
func godotenvRead(p string) (map[string]string, error) { return godotenv.Read(p) }
func godotenvWrite(m map[string]string, p string) error { return godotenv.Write(m, p) }

func getConfigPath() string {
	return infra.GetConfigPath()
}
