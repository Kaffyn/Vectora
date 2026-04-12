package acp

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/coder/acp-go-sdk"
	"github.com/google/uuid"
)

// VectoraAgent implements the ACP Agent interface for when Vectora Core
// operates as an agent (e.g., invoked by Claude Code, VS Code Extension).
// Phase 4G: Includes embedding tools for RAG capabilities.
type VectoraAgent struct {
	name           string
	version        string
	kvStore        db.KVStore
	vecStore       db.VectorStore
	llmRouter      *llm.Router
	msgService     *llm.MessageService
	logger         *slog.Logger
	sessions       map[acp.SessionId]*AgentSession
	sessionLock    sync.RWMutex
	embeddingTools map[string]interface{} // Phase 4G: Registered embedding tools
}

// AgentSession tracks state for a single agent session.
type AgentSession struct {
	ID          acp.SessionId
	WorkspaceID string
	Messages    []acp.ContentBlock
	Context     context.Context
	Cancel      context.CancelFunc
}

// NewVectoraAgent creates a new ACP agent instance.
func NewVectoraAgent(
	name, version string,
	kvStore db.KVStore,
	vecStore db.VectorStore,
	llmRouter *llm.Router,
	msgService *llm.MessageService,
	logger *slog.Logger,
) *VectoraAgent {
	return &VectoraAgent{
		name:       name,
		version:    version,
		kvStore:    kvStore,
		vecStore:   vecStore,
		llmRouter:  llmRouter,
		msgService: msgService,
		logger:     logger,
		sessions:   make(map[acp.SessionId]*AgentSession),
	}
}

// Authenticate handles client authentication.
// For now, we accept all authenticated connections (no API keys required in this agent).
func (a *VectoraAgent) Authenticate(ctx context.Context, params acp.AuthenticateRequest) (acp.AuthenticateResponse, error) {
	a.logger.Debug("Authentication request received", slog.Any("method_id", params.MethodId))
	return acp.AuthenticateResponse{}, nil
}

// Initialize is called by the ACP client when establishing connection.
func (a *VectoraAgent) Initialize(ctx context.Context, params acp.InitializeRequest) (acp.InitializeResponse, error) {
	a.logger.Info("ACP agent initialized",
		slog.String("client_name", params.ClientInfo.Name),
		slog.String("client_version", params.ClientInfo.Version),
	)

	// Return agent capabilities and implementation info
	return acp.InitializeResponse{
		AgentInfo: &acp.Implementation{
			Name:    a.name,
			Version: a.version,
		},
		AgentCapabilities: acp.AgentCapabilities{
			LoadSession: false,
			McpCapabilities: acp.McpCapabilities{
				Http: false,
				Sse:  false,
			},
			PromptCapabilities: acp.PromptCapabilities{
				Audio:           false,
				Image:           false,
				EmbeddedContext: true,
			},
		},
		ProtocolVersion: acp.ProtocolVersionNumber,
		AuthMethods:     []acp.AuthMethod{},
	}, nil
}

// Cancel stops ongoing operations for a session.
func (a *VectoraAgent) Cancel(ctx context.Context, params acp.CancelNotification) error {
	a.sessionLock.RLock()
	session, exists := a.sessions[params.SessionId]
	a.sessionLock.RUnlock()

	if !exists {
		return fmt.Errorf("session not found: %s", params.SessionId)
	}

	// Cancel the session context to stop any in-progress operations
	if session.Cancel != nil {
		session.Cancel()
	}

	a.logger.Info("Session cancelled", slog.String("session_id", string(params.SessionId)))
	return nil
}

// NewSession creates a new agent session.
func (a *VectoraAgent) NewSession(ctx context.Context, params acp.NewSessionRequest) (acp.NewSessionResponse, error) {
	sessionID := acp.SessionId(uuid.New().String())

	// Create a cancellable context for this session
	sessionCtx, cancel := context.WithCancel(ctx)

	session := &AgentSession{
		ID:       sessionID,
		Messages: make([]acp.ContentBlock, 0),
		Context:  sessionCtx,
		Cancel:   cancel,
	}

	a.sessionLock.Lock()
	a.sessions[sessionID] = session
	a.sessionLock.Unlock()

	a.logger.Info("New session created", slog.String("session_id", string(sessionID)))

	return acp.NewSessionResponse{
		SessionId: sessionID,
	}, nil
}

// Prompt processes a user prompt in an agent session.
func (a *VectoraAgent) Prompt(ctx context.Context, params acp.PromptRequest) (acp.PromptResponse, error) {
	a.sessionLock.RLock()
	session, exists := a.sessions[params.SessionId]
	a.sessionLock.RUnlock()

	if !exists {
		return acp.PromptResponse{}, fmt.Errorf("session not found: %s", params.SessionId)
	}

	// Append the user prompt to session messages
	session.Messages = append(session.Messages, params.Prompt...)

	// Extract text from content blocks for LLM
	var userText string
	for _, block := range params.Prompt {
		if block.Text != nil {
			userText += block.Text.Text + "\n"
		}
	}

	a.logger.Debug("User prompt received",
		slog.String("session_id", string(params.SessionId)),
		slog.String("prompt", userText),
	)

	// TODO: Forward prompt to LLM via msgService and collect response
	// For now, just acknowledge the prompt was received

	return acp.PromptResponse{
		StopReason: acp.StopReasonEndTurn,
	}, nil
}

// SetSessionMode updates the mode of an agent session.
func (a *VectoraAgent) SetSessionMode(ctx context.Context, params acp.SetSessionModeRequest) (acp.SetSessionModeResponse, error) {
	a.sessionLock.RLock()
	_, exists := a.sessions[params.SessionId]
	a.sessionLock.RUnlock()

	if !exists {
		return acp.SetSessionModeResponse{}, fmt.Errorf("session not found: %s", params.SessionId)
	}

	a.logger.Debug("Session mode set",
		slog.String("session_id", string(params.SessionId)),
		slog.String("mode_id", string(params.ModeId)),
	)

	return acp.SetSessionModeResponse{}, nil
}
