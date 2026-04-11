package ipc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// WorkspaceContext define o contexto de um tenant conectado.
// Ele amarra uma conexão IPC a um workspace específico no disco.
type WorkspaceContext struct {
	WorkspaceID   string
	WorkspaceRoot string
	ProjectName   string
	ContextCancel context.CancelFunc
}

// WorkspaceInitRequest representa a mensagem de inicialização enviada pelo cliente.
type WorkspaceInitRequest struct {
	WorkspaceRoot string `json:"workspace_root"`
	ProjectName   string `json:"project_name"`
}

// GenerateWorkspaceID cria um ID único e consistente baseado no caminho absoluto do workspace.
// Isso garante que o Projeto A sempre tenha o mesmo ID, independentemente da sessão.
func GenerateWorkspaceID(workspaceRoot string) string {
	hash := sha256.Sum256([]byte(workspaceRoot))
	return hex.EncodeToString(hash[:])
}

// WorkspaceInitResponse - resposta da inicialização do workspace
type WorkspaceInitResponse struct {
	WorkspaceID string `json:"workspace_id"`
	ProjectName string `json:"project_name"`
	Status      string `json:"status"`
}

// ValidateWorkspaceInit valida os campos da requisição de inicialização
func ValidateWorkspaceInit(req WorkspaceInitRequest) error {
	if req.WorkspaceRoot == "" {
		return fmt.Errorf("workspace_root não pode estar vazio")
	}

	if req.ProjectName == "" {
		return fmt.Errorf("project_name não pode estar vazio")
	}

	return nil
}
