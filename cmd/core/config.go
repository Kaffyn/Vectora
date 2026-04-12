package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Vectora core configurations",
	Long:  "Read or modify global configuration keys stored in ~/.Vectora/.env",
	Run: func(cmd *cobra.Command, args []string) {
		runConfigInteractive()
	},
}

// validConfigKeys lists all recognized configuration keys.
var validConfigKeys = map[string]string{
	"GEMINI_API_KEY":          "Google Gemini API key",
	"ANTHROPIC_API_KEY":       "Anthropic Claude API key",
	"VOYAGE_API_KEY":          "Voyage AI API key for embeddings",
	"OPENAI_API_KEY":          "OpenAI API key",
	"OPENAI_BASE_URL":         "OpenAI Custom API Base URL",
	"QWEN_API_KEY":            "Qwen API key",
	"QWEN_BASE_URL":           "Qwen Custom API Base URL",
	"OPENROUTER_API_KEY":      "OpenRouter API key",
	"ANANNAS_API_KEY":         "Anannas API key",
	"DEFAULT_PROVIDER":        "Default LLM provider (gemini, claude, openai, qwen, openrouter, anannas)",
	"DEFAULT_FALLBACK_PROVIDER": "Fallback LLM provider (gemini, claude, etc.)",
	"DEFAULT_MODEL":           "Default model identifier",
	"DEFAULT_EMBEDDING_MODEL": "Default embedding model",
	"GEMINI_FALLBACK_MODEL":   "Fallback model for Gemini",
	"CLAUDE_FALLBACK_MODEL":   "Fallback model for Claude",
	"OPENAI_FALLBACK_MODEL":   "Fallback model for OpenAI",
	"DEEPSEEK_API_KEY":        "DeepSeek API key",
	"DEEPSEEK_BASE_URL":       "DeepSeek Custom API Base URL",
	"MISTRAL_API_KEY":         "Mistral API key",
	"MISTRAL_BASE_URL":        "Mistral Custom API Base URL",
	"GROK_API_KEY":            "xAI Grok API key",
	"GROK_BASE_URL":           "xAI Grok Custom API Base URL",
	"ZHIPU_API_KEY":           "Zhipu GLM-5 API key",
	"ZHIPU_BASE_URL":          "Zhipu GLM-5 Custom API Base URL",
	"LOG_LEVEL":               "Log verbosity (debug, info, warn, error)",
}

var configSetCmd = &cobra.Command{
	Use:   "set [KEY] [VALUE]",
	Short: "Set a configuration key",
	Long: `Set a configuration key in ~/.Vectora/.env

Valid keys:
  GEMINI_API_KEY           Google Gemini API key
  ANTHROPIC_API_KEY        Anthropic Claude API key
  VOYAGE_API_KEY           Voyage AI API key for embeddings
  OPENAI_API_KEY           OpenAI API key
  OPENAI_BASE_URL          OpenAI Custom API Base URL
  QWEN_API_KEY             Current Qwen API key
  QWEN_BASE_URL            Qwen Custom API Base URL
  OPENROUTER_API_KEY       OpenRouter API key
  ANANNAS_API_KEY          Anannas API key
  DEFAULT_PROVIDER         Default LLM provider
  DEFAULT_FALLBACK_PROVIDER Fallback LLM provider
  DEFAULT_MODEL            Default model identifier
  DEFAULT_EMBEDDING_MODEL  Default embedding model
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
			fmt.Println("Error: Configuration key cannot be empty.")
			return
		}

		// Validate value is not empty
		if value == "" {
			fmt.Println("Error: Configuration value cannot be empty.")
			return
		}

		// Warn on unrecognized keys
		if _, ok := validConfigKeys[key]; !ok {
			fmt.Printf("Warning: '%s' is not a recognized key.\n", key)
			fmt.Println("Valid keys:")
			for k, desc := range validConfigKeys {
				fmt.Printf("  %-30s %s\n", k, desc)
			}
			fmt.Println("\nProceeding anyway...")
		}

		envPath := getConfigPath()
		envMap, err := godotenv.Read(envPath)
		if err != nil {
			if os.IsNotExist(err) {
				envMap = make(map[string]string)
			} else {
				fmt.Println("Error reading .env:", err)
				return
			}
		}

		envMap[key] = value
		if err := godotenv.Write(envMap, envPath); err != nil {
			fmt.Println("Error writing .env:", err)
			return
		}

		fmt.Printf("Key %s has been set successfully.\n", key)
		fmt.Println("Restarting Core to apply changes...")
		runStop()
		spawnDetached()
		fmt.Println("Core restarted.")
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get [KEY]",
	Short: "Get a configuration key value",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		envPath := getConfigPath()
		envMap, err := godotenv.Read(envPath)
		if err != nil {
			fmt.Printf("Error reading configuration: %v\n", err)
			return
		}

		if val, ok := envMap[key]; ok {
			displayVal := val
			if len(val) > 20 {
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
			if len(v) > 10 {
				displayVal = v[:4] + "..." + v[len(v)-4:]
			}
			fmt.Printf("  %s=%s\n", k, displayVal)
		}
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
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
	envMap, err := godotenv.Read(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			envMap = make(map[string]string)
		} else {
			fmt.Println("Error reading .env:", err)
			return
		}
	}

	envMap[key] = value
	if err := godotenv.Write(envMap, envPath); err != nil {
		fmt.Println("Error writing .env:", err)
		return
	}

	fmt.Printf("\nKey %s has been set successfully.\n", key)
	fmt.Println("Changes will be applied when Core restarts.")
}

func getConfigPath() string {
	userProfile, _ := os.UserHomeDir()
	return filepath.Join(userProfile, ".Vectora", ".env")
}
