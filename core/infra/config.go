// Package infra provides infrastructure utilities for Vectora.
//
// Configuration:
// Vectora uses a simple .env file for configuration, located at:
//   - Windows: %USERPROFILE%\.Vectora\.env
//   - Linux/macOS: ~/.Vectora/.env
//
// Supported environment variables:
//   - GEMINI_API_KEY: API key for Google Gemini (required for embeddings and chat)
//   - CLAUDE_API_KEY: API key for Anthropic Claude (optional, alternative provider)
//
// Example .env file:
//
//	GEMINI_API_KEY=AIzaSy...
//	CLAUDE_API_KEY=sk-ant-...
//
// Note: API keys are stored in plaintext. For production deployments with
// strict security requirements, consider using OS-level secret management
// (Windows Credential Manager, macOS Keychain, Linux libsecret) and
// loading keys from there instead.
package infra

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config holds the runtime configuration for Vectora.
// Supports all 10 LLM families from AGENTS.md (Standard April 2026):
// 1. OpenAI (GPT-5.4), 2. Anthropic (Claude 4.6), 3. Google (Gemini 3.1)
// 4. Alibaba (Qwen 3.6), 5. Voyage AI (Voyage-3), 6. Meta (Llama 4, Muse)
// 7. Microsoft (Phi-4), 8. DeepSeek (V3), 9. Mistral, 10. xAI (Grok)
// + Zhipu AI (GLM-5)
type Config struct {
	GeminiAPIKey          string
	ClaudeAPIKey          string
	VoyageAPIKey          string
	OpenAIAPIKey          string
	OpenAIBaseURL         string
	QwenAPIKey            string
	QwenBaseURL           string
	OpenRouterAPIKey      string
	AnannasAPIKey         string
	DeepSeekAPIKey        string
	DeepSeekBaseURL       string
	MistralAPIKey         string
	MistralBaseURL        string
	GrokAPIKey            string
	GrokBaseURL           string
	ZhipuAPIKey           string
	ZhipuBaseURL          string
	DefaultEmbeddingModel string
	DefaultModel          string
	DefaultProvider       string
}

// LoadConfig loads configuration from %USERPROFILE%\.Vectora\.env.
// If the file doesn't exist, it returns a Config with empty keys.
// This is the official configuration method for Vectora.
func LoadConfig() *Config {
	userProfile, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Could not find user home directory: %v", err)
	}

	envPath := filepath.Join(userProfile, ".Vectora", ".env")

	if _, err := os.Stat(filepath.Dir(envPath)); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(envPath), 0755)
	}

	if err := godotenv.Overload(envPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: error loading .env file: %v", err)
	}

	return &Config{
		GeminiAPIKey:          os.Getenv("GEMINI_API_KEY"),
		ClaudeAPIKey:          os.Getenv("CLAUDE_API_KEY"),
		VoyageAPIKey:          os.Getenv("VOYAGE_API_KEY"),
		OpenAIAPIKey:          os.Getenv("OPENAI_API_KEY"),
		OpenAIBaseURL:         os.Getenv("OPENAI_BASE_URL"),
		QwenAPIKey:            os.Getenv("QWEN_API_KEY"),
		QwenBaseURL:           os.Getenv("QWEN_BASE_URL"),
		OpenRouterAPIKey:      os.Getenv("OPENROUTER_API_KEY"),
		AnannasAPIKey:         os.Getenv("ANANNAS_API_KEY"),
		DeepSeekAPIKey:        os.Getenv("DEEPSEEK_API_KEY"),
		DeepSeekBaseURL:       os.Getenv("DEEPSEEK_BASE_URL"),
		MistralAPIKey:         os.Getenv("MISTRAL_API_KEY"),
		MistralBaseURL:        os.Getenv("MISTRAL_BASE_URL"),
		GrokAPIKey:            os.Getenv("GROK_API_KEY"),
		GrokBaseURL:           os.Getenv("GROK_BASE_URL"),
		ZhipuAPIKey:           os.Getenv("ZHIPU_API_KEY"),
		ZhipuBaseURL:          os.Getenv("ZHIPU_BASE_URL"),
		DefaultEmbeddingModel: os.Getenv("DEFAULT_EMBEDDING_MODEL"),
		DefaultModel:          os.Getenv("DEFAULT_MODEL"),
		DefaultProvider:       os.Getenv("DEFAULT_PROVIDER"),
	}
}

// SaveConfig persists API keys to the user's .env file.
// Only keys that are non-empty are written.
func SaveConfig(cfg *Config) error {
	userProfile, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	envPath := filepath.Join(userProfile, ".Vectora", ".env")

	content := ""
	if cfg.GeminiAPIKey != "" {
		content += fmt.Sprintf("GEMINI_API_KEY=%s\n", cfg.GeminiAPIKey)
	}
	if cfg.ClaudeAPIKey != "" {
		content += fmt.Sprintf("CLAUDE_API_KEY=%s\n", cfg.ClaudeAPIKey)
	}
	if cfg.VoyageAPIKey != "" {
		content += fmt.Sprintf("VOYAGE_API_KEY=%s\n", cfg.VoyageAPIKey)
	}
	if cfg.OpenAIAPIKey != "" {
		content += fmt.Sprintf("OPENAI_API_KEY=%s\n", cfg.OpenAIAPIKey)
	}
	if cfg.OpenAIBaseURL != "" {
		content += fmt.Sprintf("OPENAI_BASE_URL=%s\n", cfg.OpenAIBaseURL)
	}
	if cfg.QwenAPIKey != "" {
		content += fmt.Sprintf("QWEN_API_KEY=%s\n", cfg.QwenAPIKey)
	}
	if cfg.QwenBaseURL != "" {
		content += fmt.Sprintf("QWEN_BASE_URL=%s\n", cfg.QwenBaseURL)
	}
	if cfg.OpenRouterAPIKey != "" {
		content += fmt.Sprintf("OPENROUTER_API_KEY=%s\n", cfg.OpenRouterAPIKey)
	}
	if cfg.AnannasAPIKey != "" {
		content += fmt.Sprintf("ANANNAS_API_KEY=%s\n", cfg.AnannasAPIKey)
	}
	if cfg.DeepSeekAPIKey != "" {
		content += fmt.Sprintf("DEEPSEEK_API_KEY=%s\n", cfg.DeepSeekAPIKey)
	}
	if cfg.DeepSeekBaseURL != "" {
		content += fmt.Sprintf("DEEPSEEK_BASE_URL=%s\n", cfg.DeepSeekBaseURL)
	}
	if cfg.MistralAPIKey != "" {
		content += fmt.Sprintf("MISTRAL_API_KEY=%s\n", cfg.MistralAPIKey)
	}
	if cfg.MistralBaseURL != "" {
		content += fmt.Sprintf("MISTRAL_BASE_URL=%s\n", cfg.MistralBaseURL)
	}
	if cfg.GrokAPIKey != "" {
		content += fmt.Sprintf("GROK_API_KEY=%s\n", cfg.GrokAPIKey)
	}
	if cfg.GrokBaseURL != "" {
		content += fmt.Sprintf("GROK_BASE_URL=%s\n", cfg.GrokBaseURL)
	}
	if cfg.ZhipuAPIKey != "" {
		content += fmt.Sprintf("ZHIPU_API_KEY=%s\n", cfg.ZhipuAPIKey)
	}
	if cfg.ZhipuBaseURL != "" {
		content += fmt.Sprintf("ZHIPU_BASE_URL=%s\n", cfg.ZhipuBaseURL)
	}
	if cfg.DefaultEmbeddingModel != "" {
		content += fmt.Sprintf("DEFAULT_EMBEDDING_MODEL=%s\n", cfg.DefaultEmbeddingModel)
	}
	if cfg.DefaultModel != "" {
		content += fmt.Sprintf("DEFAULT_MODEL=%s\n", cfg.DefaultModel)
	}
	if cfg.DefaultProvider != "" {
		content += fmt.Sprintf("DEFAULT_PROVIDER=%s\n", cfg.DefaultProvider)
	}

	return os.WriteFile(envPath, []byte(content), 0600)
}
