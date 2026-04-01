package desktop

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

// Preferences armazena as preferências do usuário
type Preferences struct {
	// Window
	WindowWidth  int    `json:"window_width"`
	WindowHeight int    `json:"window_height"`

	// Services
	IndexServiceAddr string `json:"index_service_addr"`
	DaemonAddr       string `json:"daemon_addr"`

	// Settings
	Theme         string `json:"theme"`
	FontSize      int    `json:"font_size"`
	DefaultWorkspace string `json:"default_workspace"`
	DefaultIndex  string `json:"default_index"`

	// Session
	LastActiveTab int    `json:"last_active_tab"`
	LastChat      string `json:"last_chat"`

	// Paths
	configPath string // internal
}

// LoadPreferences carrega as preferências do usuário
func LoadPreferences() *Preferences {
	prefs := &Preferences{
		WindowWidth:      1400,
		WindowHeight:     900,
		IndexServiceAddr: "localhost:3000",
		DaemonAddr:       "localhost:8765",
		Theme:            "light",
		FontSize:         12,
		LastActiveTab:    0,
	}

	// Determinar path do arquivo de configuração
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("[WARN] Não foi possível obter home dir: %v\n", err)
		return prefs
	}

	configDir := filepath.Join(homeDir, ".vectora")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Printf("[WARN] Não foi possível criar config dir: %v\n", err)
		return prefs
	}

	configPath := filepath.Join(configDir, "desktop_prefs.json")
	prefs.configPath = configPath

	// Carregar arquivo existente
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Arquivo não existe, salvar defaults
			prefs.Save()
			return prefs
		}
		log.Printf("[WARN] Erro ao ler preferences: %v\n", err)
		return prefs
	}

	// Parse JSON
	if err := json.Unmarshal(data, prefs); err != nil {
		log.Printf("[WARN] Erro ao desserializar preferences: %v\n", err)
		return prefs
	}

	prefs.configPath = configPath
	return prefs
}

// Save salva as preferências do usuário
func (p *Preferences) Save() error {
	if p.configPath == "" {
		return nil
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(p.configPath, data, 0600)
}
