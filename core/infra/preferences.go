package infra

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// UserPreferences stores non-secret UI and usage settings.
type UserPreferences struct {
	ActiveModel      string `json:"active_model"`
	DefaultProvider  string `json:"default_provider"`
	Language         string `json:"language"`
	Theme            string `json:"theme"`
	LastWorkspace    string `json:"last_workspace"`
}

var (
	prefLock sync.RWMutex
	prefPath string
)

func init() {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		home, _ := os.UserHomeDir()
		prefPath = filepath.Join(home, ".Vectora", "preferences.json")
	} else {
		prefPath = filepath.Join(appData, "Vectora", "preferences.json")
	}
}

// LoadPreferences reads non-secret settings from %APPDATA%/Vectora/preferences.json
func LoadPreferences() *UserPreferences {
	prefLock.RLock()
	defer prefLock.RUnlock()

	data, err := os.ReadFile(prefPath)
	if err != nil {
		return &UserPreferences{
			Language: "en",
		}
	}

	var prefs UserPreferences
	if err := json.Unmarshal(data, &prefs); err != nil {
		return &UserPreferences{
			Language: "en",
		}
	}

	return &prefs
}

// SavePreferences persists non-secret settings to JSON.
func SavePreferences(prefs *UserPreferences) error {
	prefLock.Lock()
	defer prefLock.Unlock()

	if err := os.MkdirAll(filepath.Dir(prefPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(prefPath, data, 0644)
}
