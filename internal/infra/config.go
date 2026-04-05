package infra

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for Vectora daemon
type Config struct {
	GeminiAPIKey         string
	MaxRAMDaemon         int64
	MaxRAMIndexing       int64
	PreferredLLMProvider string
	IPCPipeWindows       string
	IPCSocketUnix        string
	LogLevel             string
	LogFormat            string
	DatabasePath         string
	WorkspacePath        string
}

// LoadConfig loads configuration from environment and .env file
func LoadConfig(envPath string) (*Config, error) {
	cfg := &Config{
		MaxRAMDaemon:         4294967296, // 4GB
		MaxRAMIndexing:       536870912,  // 512MB
		PreferredLLMProvider: "qwen_local",
		LogLevel:             "INFO",
		LogFormat:            "json",
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	cfg.DatabasePath = fmt.Sprintf("%s/.Vectora/db/vectora.db", homeDir)
	cfg.WorkspacePath = fmt.Sprintf("%s/.Vectora/workspaces", homeDir)

	// Load from .env if exists
	if envPath == "" {
		envPath = fmt.Sprintf("%s/.Vectora/.env", homeDir)
	}

	if _, err := os.Stat(envPath); err == nil {
		if err := loadEnvFile(envPath, cfg); err != nil {
			return nil, fmt.Errorf("failed to load .env: %w", err)
		}
	}

	// Override with environment variables
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		cfg.GeminiAPIKey = key
	}
	if maxRam := os.Getenv("MAX_RAM_DAEMON"); maxRam != "" {
		if val, err := strconv.ParseInt(maxRam, 10, 64); err == nil {
			cfg.MaxRAMDaemon = val
		}
	}
	if maxRam := os.Getenv("MAX_RAM_INDEXING"); maxRam != "" {
		if val, err := strconv.ParseInt(maxRam, 10, 64); err == nil {
			cfg.MaxRAMIndexing = val
		}
	}
	if llm := os.Getenv("PREFERRED_LLM_PROVIDER"); llm != "" {
		cfg.PreferredLLMProvider = llm
	}
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.LogLevel = level
	}

	return cfg, nil
}

// loadEnvFile parses a simple .env file
func loadEnvFile(path string, cfg *Config) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)

		switch key {
		case "GEMINI_API_KEY":
			cfg.GeminiAPIKey = value
		case "MAX_RAM_DAEMON":
			if val, err := strconv.ParseInt(value, 10, 64); err == nil {
				cfg.MaxRAMDaemon = val
			}
		case "MAX_RAM_INDEXING":
			if val, err := strconv.ParseInt(value, 10, 64); err == nil {
				cfg.MaxRAMIndexing = val
			}
		case "PREFERRED_LLM_PROVIDER":
			cfg.PreferredLLMProvider = value
		case "LOG_LEVEL":
			cfg.LogLevel = value
		case "LOG_FORMAT":
			cfg.LogFormat = value
		}
	}

	return scanner.Err()
}

// Validate checks configuration validity
func (c *Config) Validate() error {
	if c.MaxRAMDaemon < 1073741824 { // 1GB min
		return fmt.Errorf("MAX_RAM_DAEMON too low: %d bytes (minimum 1GB)", c.MaxRAMDaemon)
	}
	if c.MaxRAMIndexing < 134217728 { // 128MB min
		return fmt.Errorf("MAX_RAM_INDEXING too low: %d bytes (minimum 128MB)", c.MaxRAMIndexing)
	}
	return nil
}

// GetString returns a string representation of the config
func (c *Config) String() string {
	return fmt.Sprintf(
		"Config{MaxRAMDaemon=%d, MaxRAMIndexing=%d, LLM=%s, LogLevel=%s}",
		c.MaxRAMDaemon, c.MaxRAMIndexing, c.PreferredLLMProvider, c.LogLevel,
	)
}
