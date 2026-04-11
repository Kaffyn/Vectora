package acp

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/coder/acp-go-sdk"
)

// StartACPAgent starts the Vectora Core as an ACP agent on stdio.
// This is called when Core is invoked as a sub-process by an ACP client
// (e.g., Claude Code, VS Code Extension).
func StartACPAgent(ctx context.Context, agent *VectoraAgent, logger *slog.Logger) error {
	logger.Info("Starting Vectora as ACP Agent",
		slog.String("agent_name", agent.name),
		slog.String("agent_version", agent.version),
	)

	// Create agent-side connection (stdio transport)
	// Note: NewAgentSideConnection expects (agent, peerInput, peerOutput)
	// peerInput = os.Stdout (writes to client's stdin)
	// peerOutput = os.Stdin (reads from client's stdout)
	conn := acp.NewAgentSideConnection(agent, os.Stdout, os.Stdin)

	// Set up the logger for the connection
	conn.SetLogger(logger)

	// The connection starts automatically and handles the ACP protocol.
	// We wait for it to complete (client disconnects) or context cancellation.
	select {
	case <-conn.Done():
		logger.Info("ACP agent connection closed")
		return nil
	case <-ctx.Done():
		logger.Info("ACP agent context cancelled")
		return ctx.Err()
	}
}

// TransportMode represents the protocol mode Vectora is running in
type TransportMode string

const (
	TransportModeIPC     TransportMode = "ipc"   // IPC mode for local clients
	TransportModeACP     TransportMode = "acp"   // ACP mode for IDE agents
	TransportModeMCP     TransportMode = "mcp"   // MCP mode for sub-agents
	TransportModeStdio   TransportMode = "stdio" // Generic stdio mode (detect later)
	TransportModeUnknown TransportMode = "unknown"
)

// DetectTransport attempts to detect which protocol mode Vectora should run in.
// Returns the detected transport mode.
func DetectTransport(logger *slog.Logger) TransportMode {
	// Check for explicit environment variable override
	if mode := os.Getenv("VECTORA_TRANSPORT"); mode != "" {
		logger.Info("Transport mode from environment", slog.String("mode", mode))
		switch mode {
		case "acp":
			return TransportModeACP
		case "mcp":
			return TransportModeMCP
		case "ipc":
			return TransportModeIPC
		case "stdio":
			return TransportModeStdio
		}
	}

	// Check for ACP_CLIENT environment variable (set by ACP clients)
	if os.Getenv("ACP_CLIENT") != "" {
		logger.Info("Detected ACP_CLIENT environment variable")
		return TransportModeACP
	}

	// Check for MCP environment variables
	if os.Getenv("MCP_SERVER") != "" {
		logger.Info("Detected MCP_SERVER environment variable")
		return TransportModeMCP
	}

	// Default to IPC mode for local operation
	logger.Debug("No explicit transport detected, defaulting to IPC mode")
	return TransportModeIPC
}

// ProtocolDetectionRequest is a minimal JSON-RPC request used to detect protocol
type ProtocolDetectionRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

// DetectStdioProtocol reads the first message and determines if it's ACP or MCP.
// This is called when VECTORA_TRANSPORT is set to "stdio".
// Note: This is more advanced and requires buffering/peeking at stdin,
// which is complex. For now, we rely on environment variables or flags.
func DetectStdioProtocol(logger *slog.Logger) TransportMode {
	// TODO: Implement protocol detection by reading first JSON-RPC message
	// and examining method name ("initialize" for ACP, etc.)
	// This requires careful handling of stdin buffering.

	logger.Warn("Stdio protocol detection not yet implemented, defaulting to ACP")
	return TransportModeACP
}
