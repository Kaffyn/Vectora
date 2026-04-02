package domain

import "context"

// OSManager define a interface para operações específicas do sistema operacional.
// Cada implementação (Windows, Linux, macOS) fornecerá a lógica necessária.
type OSManager interface {
	// StartLLAMAServer inicia um servidor llama.cpp com um ID específico (ex: "text", "embedding").
	StartLLAMAServer(ctx context.Context, id string, modelPath string, port int, enableGPU bool) error
	// StopLLAMAServer encerra um servidor específico.
	StopLLAMAServer(id string) error
	// IsLLAMAServerRunning verifica se um servidor específico está rodando.
	IsLLAMAServerRunning(id string) bool
	// StopAllLLAMAServers encerra todos os servidores ativos.
	StopAllLLAMAServers() error
	// GetLlamaServerBinaryPath retorna o caminho para o binário do llama-server.
	GetLlamaServerBinaryPath() string
	// RunCommand executa um comando de sistema de forma síncrona.
	RunCommand(ctx context.Context, cmd string, args []string) (string, error)
}
