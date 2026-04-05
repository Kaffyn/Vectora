package infra

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.MaxRAMDaemon != 4294967296 {
		t.Errorf("Default MaxRAMDaemon incorrect")
	}
	if cfg.MaxRAMIndexing != 536870912 {
		t.Errorf("Default MaxRAMIndexing incorrect")
	}
	if cfg.PreferredLLMProvider != "qwen_local" {
		t.Errorf("Default PreferredLLMProvider incorrect")
	}
}

func TestConfigValidate(t *testing.T) {
	cfg := &Config{
		MaxRAMDaemon:   4294967296,
		MaxRAMIndexing: 536870912,
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Valid config failed validation: %v", err)
	}
}

func TestConfigValidateInvalid(t *testing.T) {
	cfg := &Config{
		MaxRAMDaemon:   1000, // Less than 1GB
		MaxRAMIndexing: 536870912,
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Invalid config did not fail validation")
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	envContent := `GEMINI_API_KEY=test-key
MAX_RAM_DAEMON=2147483648
PREFERRED_LLM_PROVIDER=custom_llm
LOG_LEVEL=DEBUG
`

	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}

	cfg, err := LoadConfig(envFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.GeminiAPIKey != "test-key" {
		t.Errorf("GEMINI_API_KEY not loaded from .env")
	}
	if cfg.MaxRAMDaemon != 2147483648 {
		t.Errorf("MAX_RAM_DAEMON not loaded from .env")
	}
	if cfg.PreferredLLMProvider != "custom_llm" {
		t.Errorf("PREFERRED_LLM_PROVIDER not loaded from .env")
	}
	if cfg.LogLevel != "DEBUG" {
		t.Errorf("LOG_LEVEL not loaded from .env")
	}
}

func TestConfigString(t *testing.T) {
	cfg := &Config{
		MaxRAMDaemon:         4294967296,
		MaxRAMIndexing:       536870912,
		PreferredLLMProvider: "test_llm",
		LogLevel:             "INFO",
	}

	str := cfg.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
	if !contains(str, "MaxRAMDaemon") {
		t.Error("String() missing MaxRAMDaemon")
	}
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
