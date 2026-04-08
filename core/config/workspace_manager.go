package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

type WorkspaceManager struct {
	BaseWorkspacesDir string
}

func NewWorkspaceManager(homeDir string) *WorkspaceManager {
	return &WorkspaceManager{
		BaseWorkspacesDir: filepath.Join(homeDir, ".vectora", "workspaces"),
	}
}

// ResolveWorkspace calcula o ID e caminho de storage para um dado path de projeto.
func (wm *WorkspaceManager) ResolveWorkspace(projectPath string) (*WorkspaceContext, error) {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, err
	}

	// Verifica se o diretório existe
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("project path does not exist: %s", absPath)
	}

	// Gera ID único baseado no path absoluto
	hash := sha256.Sum256([]byte(absPath))
	id := hex.EncodeToString(hash[:16]) // Usar primeiros 16 chars para ser curto mas único o suficiente

	storagePath := filepath.Join(wm.BaseWorkspacesDir, id)

	// Garante que o diretório de storage exista
	if err := os.MkdirAll(storagePath, 0750); err != nil {
		return nil, err
	}

	return &WorkspaceContext{
		Path:        absPath,
		ID:          id,
		StoragePath: storagePath,
		IsActive:    true,
	}, nil
}
