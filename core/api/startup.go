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

// ProtocolMode defines which protocol mode Vectora is operating in.
type ProtocolMode string

const (
	ModeACP ProtocolMode = "acp" // Agent Client Protocol (IDE agent)
	ModeMCP ProtocolMode = "mcp" // Model Context Protocol (sub-agent)
	ModeIPC ProtocolMode = "ipc" // Inter-Process Communication (local)
)

// InitializeProtocol detects and initializes the appropriate protocol.
// Phase 7E: Protocol wire-up for ACP and MCP in Core startup.
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
	logger.Info("Initializing protocol",
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
		logger.Info("IPC mode - waiting for connections on socket/pipe")
		// IPC is configured in cmd/core/main.go
		return nil

	default:
		logger.Warn("Unknown mode, using default IPC", slog.String("mode", string(mode)))
		return nil
	}
}

// initializeACPAgent starts Vectora as an ACP agent.
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
	logger.Info("Initializing as ACP Agent",
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

// initializeMCPServer starts Vectora as an MCP server via stdio (JSON-RPC 2.0).
// Implements the Model Context Protocol over stdin/stdout to function as a sub-agent.
// When invoked by Claude Code, Antigravity, or other MCP clients, communicates via stdio.
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
	logger.Info("Initializing as MCP Server",
		slog.String("name", name),
		slog.String("version", version),
		slog.String("protocol", "stdio"),
	)

	// Create Engine wrapping all dependencies.
	// Engine.Tools exposes all available tools via the registry.
	// Engine.ExecuteTool() executes tools with Guardian validation.
	eng := mcp.NewVectoraMCPServer(
		name,
		version,
		kvStore,
		vecStore,
		router,
		msgService,
		logger,
	)

	// Create StdioServer for JSON-RPC 2.0 communication via stdin/stdout.
	// Implements full Model Context Protocol for sub-agent invocation.
	stdioServer := mcp.NewStdioServerFromMCP(eng, kvStore, vecStore, router, logger)

	logger.Info("MCP Server started via stdio",
		slog.String("protocol_version", "2024-11-05"),
	)
	logger.Debug("Waiting for JSON-RPC 2.0 messages on stdin")

	// Start MCP server and block until context is cancelled.
	return stdioServer.Start(ctx)
}

// DetectProtocolMode detects which protocol to use based on environment variables.
func DetectProtocolMode(logger *slog.Logger) ProtocolMode {
	// Check VECTORA_PROTOCOL environment variable.
	if protocol := os.Getenv("VECTORA_PROTOCOL"); protocol != "" {
		logger.Info("Protocol detected via VECTORA_PROTOCOL", slog.String("protocol", protocol))
		switch protocol {
		case "acp":
			return ModeACP
		case "mcp":
			return ModeMCP
		case "ipc":
			return ModeIPC
		}
	}

	// Check environment variables indicating invocation by ACP client.
	if os.Getenv("VECTORA_ACP_AGENT") != "" {
		logger.Info("Detected as ACP Agent")
		return ModeACP
	}

	// Check environment variables indicating invocation as MCP server.
	if os.Getenv("VECTORA_MCP_SERVER") != "" {
		logger.Info("Detected as MCP Server")
		return ModeMCP
	}

	// Default: IPC mode for local operation.
	logger.Debug("No protocol detected, using default IPC")
	return ModeIPC
}
