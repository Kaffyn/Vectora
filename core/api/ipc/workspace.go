package ipc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

// Nota: HandleWorkspaceInit será invocado diretamente no loop de conexão do servidor
// para garantir que o estado do tenant seja capturado antes de outros handlers.
