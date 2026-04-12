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

	formatLabel := func(base, key string) string {
		if key == "" {
			return base + " (not set)"
		}
		prefixLen := 7
		if len(key) < prefixLen {
			prefixLen = len(key)
		}
		return fmt.Sprintf("%s (%s...)", base, key[:prefixLen])
	}

	gatewayLabel := func(name, key string) string {
		active := cfg.ActiveGateway == name
		label := formatLabel(name, key)
		if active {
			label += " ★ ativo"
		}
		return label
	}

	options := []string{
		// ── GATEWAY (camada de roteamento, opcional) ──────────────────
		"── Gateway ──────────────────────────────────",
		gatewayLabel("Nenhum (direto)", ""),
		gatewayLabel("OpenRouter", cfg.OpenRouterAPIKey),
		gatewayLabel("Anannas", cfg.AnannasAPIKey),
		// ── PROVIDERS NATIVOS ─────────────────────────────────────────
		"── Providers ───────────────────────────────",
		formatLabel("Google Gemini", cfg.GeminiAPIKey),
		formatLabel("Anthropic Claude", cfg.ClaudeAPIKey),
		formatLabel("OpenAI", cfg.OpenAIAPIKey),
		formatLabel("DeepSeek", cfg.DeepSeekAPIKey),
		formatLabel("Mistral", cfg.MistralAPIKey),
		formatLabel("xAI Grok", cfg.GrokAPIKey),
		formatLabel("Alibaba Qwen", cfg.QwenAPIKey),
		formatLabel("Zhipu GLM-5", cfg.ZhipuAPIKey),
		// ── EMBEDDING ─────────────────────────────────────────────────
		"── Embedding only ──────────────────────────",
		formatLabel("Voyage AI", cfg.VoyageAPIKey),
		// ── CONFIGURAÇÕES ─────────────────────────────────────────────
		"── Configurações ───────────────────────────",
		"Set Default Provider",
		"Set Default Model",
		"Set Fallback Provider",
		"Cancel",
	}

	var choice string
	prompt := &survey.Select{
		Message: "Vectora — configuração:",
		Options: options,
	}
	err := survey.AskOne(prompt, &choice)
	if err != nil || choice == "Cancel" {
		return
	}

	// Ignorar separadores (linhas que começam com "──")
	if len(choice) >= 3 && choice[:3] == "── " {
		return
	}

	// Extrai base sem sufixos " (..." e " ★ ativo"
	baseChoice := choice
	if idx := stringsIndexOf(choice, " ("); idx != -1 {
		baseChoice = choice[:idx]
	}
	if idx := stringsIndexOf(baseChoice, " ★"); idx != -1 {
		baseChoice = baseChoice[:idx]
	}

	switch baseChoice {
	// ── Gateway
	case "Nenhum (direto)":
		saveEnvKey("ACTIVE_GATEWAY", "")
		fmt.Println("Gateway desativado — conexão direta com providers.")
	case "OpenRouter":
		promptKeyAndSave("OPENROUTER_API_KEY", "OpenRouter API Key:", true)
		saveEnvKey("ACTIVE_GATEWAY", "openrouter")
		fmt.Println("Gateway definido: OpenRouter. Todos os requests serão roteados por ele.")
	case "Anannas":
		promptKeyAndSave("ANANNAS_API_KEY", "Anannas API Key:", true)
		saveEnvKey("ACTIVE_GATEWAY", "anannas")
		fmt.Println("Gateway definido: Anannas.")

	// ── Providers nativos
	case "Google Gemini":
		promptKeyAndSave("GEMINI_API_KEY", "Gemini API Key:", true)
	case "Anthropic Claude":
		promptKeyAndSave("ANTHROPIC_API_KEY", "Anthropic/Claude API Key:", true)
	case "OpenAI":
		promptKeyAndSave("OPENAI_API_KEY", "OpenAI API Key:", true)
		promptKeyAndSave("OPENAI_BASE_URL", "OpenAI Base URL (opcional):", false)
	case "DeepSeek":
		promptKeyAndSave("DEEPSEEK_API_KEY", "DeepSeek API Key:", true)
		promptKeyAndSave("DEEPSEEK_BASE_URL", "DeepSeek Base URL (opcional):", false)
	case "Mistral":
		promptKeyAndSave("MISTRAL_API_KEY", "Mistral API Key:", true)
		promptKeyAndSave("MISTRAL_BASE_URL", "Mistral Base URL (opcional):", false)
	case "xAI Grok":
		promptKeyAndSave("GROK_API_KEY", "xAI Grok API Key:", true)
		promptKeyAndSave("GROK_BASE_URL", "xAI Grok Base URL (opcional):", false)
	case "Alibaba Qwen":
		promptKeyAndSave("QWEN_API_KEY", "Qwen API Key:", true)
		promptKeyAndSave("QWEN_BASE_URL", "Qwen Base URL (opcional):", false)
	case "Zhipu GLM-5":
		promptKeyAndSave("ZHIPU_API_KEY", "Zhipu AI / GLM API Key:", true)
		promptKeyAndSave("ZHIPU_BASE_URL", "Zhipu Base URL (opcional):", false)

	// ── Embedding
	case "Voyage AI":
		promptKeyAndSave("VOYAGE_API_KEY", "Voyage AI API Key (voyage-3, voyage-code-3 — embeddings only):", true)

	// ── Configurações
	case "Set Default Provider":
		selectAndSave("DEFAULT_PROVIDER", "Provider padrão (usado sem gateway):",
			[]string{"gemini", "claude", "openai", "deepseek", "mistral", "grok", "qwen", "zhipu", "voyage"})
	case "Set Default Model":
		promptKeyAndSave("DEFAULT_MODEL", "Model ID padrão (ex: gemini-3-flash-preview):", false)
	case "Set Fallback Provider":
		selectAndSave("DEFAULT_FALLBACK_PROVIDER", "Provider de fallback:",
			[]string{"gemini", "claude", "openai", "voyage"})
	}
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
