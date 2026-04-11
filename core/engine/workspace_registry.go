package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

var registryMu sync.Mutex

// registerWorkspacePath records a workspace collection name → source path mapping
// in ~/.Vectora/workspaces.json for display by `workspace ls`.
func registerWorkspacePath(collectionName, sourcePath string) {
	registryMu.Lock()
	defer registryMu.Unlock()

	regPath := workspaceRegistryPath()

	reg := make(map[string]string)
	if data, err := os.ReadFile(regPath); err == nil {
		json.Unmarshal(data, &reg)
	}

	reg[collectionName] = sourcePath

	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return
	}

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(regPath), 0755)
	os.WriteFile(regPath, data, 0644)
}

func workspaceRegistryPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".Vectora", "workspaces.json")
}
