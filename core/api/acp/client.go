package acp

import (
	"log/slog"

	"github.com/coder/acp-go-sdk"
)

// IDEClientConnection fornece acesso aos métodos de requisição disponíveis no cliente IDE.
// O Core usa essa interface para fazer requisições bidirecionais ao IDE cliente
// (VS Code, Claude Code, etc) via a AgentSideConnection.
//
// Phase 7C: Cliente ACP para que o Core faça requisições ao IDE
// Exemplos de uso: pedir permissão, ler/escrever arquivos, criar terminais, etc.
type IDEClientConnection struct {
	conn   *acp.AgentSideConnection
	logger *slog.Logger
}

// NewIDEClientConnection cria um wrapper para acessar funcionalidades de cliente IDE.
// A AgentSideConnection é criada quando o VectoraAgent recebe a conexão do IDE.
func NewIDEClientConnection(conn *acp.AgentSideConnection, logger *slog.Logger) *IDEClientConnection {
	return &IDEClientConnection{
		conn:   conn,
		logger: logger,
	}
}

// GetConnection retorna a AgentSideConnection para acesso direto aos métodos do cliente.
// Métodos disponíveis:
// - RequestPermission() - Requisitar aprovação para ações
// - ReadTextFile() - Ler arquivo do IDE
// - WriteTextFile() - Escrever arquivo no IDE
// - CreateTerminal() - Criar terminal
// - TerminalOutput() - Obter output do terminal
// - WaitForTerminalExit() - Esperar término do terminal
// - ReleaseTerminal() - Liberar terminal
// - KillTerminalCommand() - Matar comando em terminal
func (c *IDEClientConnection) GetConnection() *acp.AgentSideConnection {
	return c.conn
}

// IsConnected verifica se a conexão com o IDE está ativa.
func (c *IDEClientConnection) IsConnected() bool {
	select {
	case <-c.conn.Done():
		return false
	default:
		return true
	}
}
