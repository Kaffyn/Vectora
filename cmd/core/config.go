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
	options := []string{
		"Google Gemini (AIzaSy...)",
		"Anthropic Claude (sk-ant...)",
		"OpenAI (sk-...)",
		"DeepSeek (V3.2)",
		"Mistral (Large 3)",
		"xAI Grok (Grok-4)",
		"Alibaba Qwen (3.6-Plus)",
		"Zhipu GLM-5 (7-April-2026)",
		"Voyage AI (Embeddings)",
		"OpenRouter (Universal)",
		"---",
		"Set Default Provider",
		"Set Default Model",
		"Set Fallback Provider",
		"Set Fallback Model",
		"Cancel",
	}

	var choice string
	prompt := &survey.Select{
		Message: "Vectora configuration dashboard:",
		Options: options,
	}
	err := survey.AskOne(prompt, &choice)
	if err != nil || choice == "Cancel" || choice == "---" {
		return
	}

	switch choice {
	case "Google Gemini (AIzaSy...)":
		promptKeyAndSave("GEMINI_API_KEY", "Gemini API Key:", true)
	case "Anthropic Claude (sk-ant...)":
		promptKeyAndSave("ANTHROPIC_API_KEY", "Anthropic/Claude API Key:", true)
	case "OpenAI (sk-...)":
		promptKeyAndSave("OPENAI_API_KEY", "OpenAI API Key:", true)
		promptKeyAndSave("OPENAI_BASE_URL", "OpenAI Base URL (optional):", false)
	case "DeepSeek (V3.2)":
		promptKeyAndSave("DEEPSEEK_API_KEY", "DeepSeek API Key:", true)
		promptKeyAndSave("DEEPSEEK_BASE_URL", "DeepSeek Base URL (optional):", false)
	case "Mistral (Large 3)":
		promptKeyAndSave("MISTRAL_API_KEY", "Mistral API Key:", true)
		promptKeyAndSave("MISTRAL_BASE_URL", "Mistral Base URL (optional):", false)
	case "xAI Grok (Grok-4)":
		promptKeyAndSave("GROK_API_KEY", "xAI Grok API Key:", true)
		promptKeyAndSave("GROK_BASE_URL", "xAI Grok Base URL (optional):", false)
	case "Alibaba Qwen (3.6-Plus)":
		promptKeyAndSave("QWEN_API_KEY", "Qwen API Key:", true)
		promptKeyAndSave("QWEN_BASE_URL", "Qwen Base URL (optional):", false)
	case "Zhipu GLM-5 (7-April-2026)":
		promptKeyAndSave("ZHIPU_API_KEY", "Zhipu AI / GLM API Key:", true)
		promptKeyAndSave("ZHIPU_BASE_URL", "Zhipu Base URL (optional):", false)
	case "Voyage AI (Embeddings)":
		promptKeyAndSave("VOYAGE_API_KEY", "Voyage AI API Key:", true)
	case "OpenRouter (Universal)":
		promptKeyAndSave("OPENROUTER_API_KEY", "OpenRouter API Key:", true)
	case "Set Default Provider":
		selectAndSave("DEFAULT_PROVIDER", "Default provider:", []string{"gemini", "claude", "openai", "deepseek", "mistral", "grok", "qwen"})
	case "Set Default Model":
		promptKeyAndSave("DEFAULT_MODEL", "Default model ID:", false)
	case "Set Fallback Provider":
		selectAndSave("DEFAULT_FALLBACK_PROVIDER", "Fallback provider:", []string{"gemini", "claude", "openai"})
	case "Set Fallback Model":
		promptKeyAndSave("GEMINI_FALLBACK_MODEL", "Gemini fallback model (e.g. gemini-3-flash-preview):", false)
	}
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
