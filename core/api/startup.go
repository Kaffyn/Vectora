package api

import (
	"context"
	"log/slog"
	"os"

	"github.com/Kaffyn/Vectora/core/api/acp"
	"github.com/Kaffyn/Vectora/core/api/mcp"
	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/llm"
)

// ProtocolMode define em qual modo Vectora está operando
type ProtocolMode string

const (
	ModeACP ProtocolMode = "acp" // Agent Client Protocol (IDE agent)
	ModeMCP ProtocolMode = "mcp" // Model Context Protocol (sub-agent)
	ModeIPC ProtocolMode = "ipc" // Inter-Process Communication (local)
)

// InitializeProtocol detecta e inicializa o protocolo apropriado.
// Phase 7E: Wire-up dos protocolos ACP e MCP no startup do Core
func InitializeProtocol(
	ctx context.Context,
	mode ProtocolMode,
	name string,
	version string,
	kvStore db.KVStore,
	vecStore db.VectorStore,
	router *llm.Router,
	msgService *llm.MessageService,
	logger *slog.Logger,
) error {
	logger.Info("Inicializando protocolo",
		slog.String("mode", string(mode)),
		slog.String("name", name),
		slog.String("version", version),
	)

	switch mode {
	case ModeACP:
		return initializeACPAgent(ctx, name, version, kvStore, vecStore, router, msgService, logger)

	case ModeMCP:
		return initializeMCPServer(ctx, name, version, kvStore, vecStore, router, msgService, logger)

	case ModeIPC:
		logger.Info("Modo IPC - aguardando conexões em socket/pipe")
		// IPC é configurado em cmd/core/main.go
		return nil

	default:
		logger.Warn("Modo desconhecido, usando IPC padrão", slog.String("mode", string(mode)))
		return nil
	}
}

// initializeACPAgent inicia Vectora como agente ACP
func initializeACPAgent(
	ctx context.Context,
	name string,
	version string,
	kvStore db.KVStore,
	vecStore db.VectorStore,
	router *llm.Router,
	msgService *llm.MessageService,
	logger *slog.Logger,
) error {
	logger.Info("Inicializando como ACP Agent",
		slog.String("name", name),
	)

	agent := acp.NewVectoraAgent(
		name,
		version,
		kvStore,
		vecStore,
		router,
		msgService,
		logger,
	)

	return acp.StartACPAgent(ctx, agent, logger)
}

// initializeMCPServer inicia Vectora como servidor MCP
func initializeMCPServer(
	ctx context.Context,
	name string,
	version string,
	kvStore db.KVStore,
	vecStore db.VectorStore,
	router *llm.Router,
	msgService *llm.MessageService,
	logger *slog.Logger,
) error {
	logger.Info("Inicializando como MCP Server",
		slog.String("name", name),
	)

	_ = mcp.NewVectoraMCPServer(
		name,
		version,
		kvStore,
		vecStore,
		router,
		msgService,
		logger,
	)

	// TODO: Implementar comunicação MCP via stdio
	// Por enquanto, apenas criar o servidor
	logger.Debug("MCP Server criado com sucesso")

	// Aguardar indefinidamente
	<-ctx.Done()
	return ctx.Err()
}

// DetectProtocolMode detecta qual protocolo usar baseado em variáveis de ambiente
func DetectProtocolMode(logger *slog.Logger) ProtocolMode {
	// Verificar variável de ambiente VECTORA_PROTOCOL
	if protocol := os.Getenv("VECTORA_PROTOCOL"); protocol != "" {
		logger.Info("Protocolo detectado via VECTORA_PROTOCOL", slog.String("protocol", protocol))
		switch protocol {
		case "acp":
			return ModeACP
		case "mcp":
			return ModeMCP
		case "ipc":
			return ModeIPC
		}
	}

	// Verificar variáveis indicando invocação por ACP client
	if os.Getenv("VECTORA_ACP_AGENT") != "" {
		logger.Info("Detectado como ACP Agent")
		return ModeACP
	}

	// Verificar variáveis indicando invocação como MCP server
	if os.Getenv("VECTORA_MCP_SERVER") != "" {
		logger.Info("Detectado como MCP Server")
		return ModeMCP
	}

	// Padrão: modo IPC para operação local
	logger.Debug("Nenhum protocolo detectado, usando IPC padrão")
	return ModeIPC
}
