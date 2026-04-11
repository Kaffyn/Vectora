// Package acp implements the Agent Client Protocol (ACP) server.
// ACP is a JSON-RPC 2.0 protocol over stdio for communication between
// code editors (VS Code, JetBrains, Zed) and AI coding agents.
//
// Protocol flow:
//  1. initialize — negotiate version and capabilities
//  2. authenticate — (optional) agent authentication
//  3. session/new — create a new session
//  4. session/prompt — send a prompt, receive streaming updates
//  5. session/update — agent → client: streaming chunks, tool calls, plans
//  6. session/request_permission — agent → client: ask for tool approval
//
// Spec: https://agentclientprotocol.com
package acp

import (
	"context"
	"fmt"
	"io"
	"os"
)

// ProtocolVersion is the current ACP major version we support.
const ProtocolVersion = 1

// ---- Initialization ----

type InitializeRequest struct {
	ProtocolVersion    int                 `json:"protocolVersion"`
	ClientCapabilities *ClientCapabilities `json:"clientCapabilities,omitempty"`
	ClientInfo         *Info               `json:"clientInfo,omitempty"`
}

type ClientCapabilities struct {
	FS       *FSCapabilities `json:"fs,omitempty"`
	Terminal bool            `json:"terminal,omitempty"`
}

type FSCapabilities struct {
	ReadTextFile  bool `json:"readTextFile,omitempty"`
	WriteTextFile bool `json:"writeTextFile,omitempty"`
}

type Info struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	Version string `json:"version"`
}

type InitializeResponse struct {
	ProtocolVersion   int                `json:"protocolVersion"`
	AgentCapabilities *AgentCapabilities `json:"agentCapabilities,omitempty"`
	AgentInfo         *Info              `json:"agentInfo"`
	AuthMethods       []string           `json:"authMethods,omitempty"`
}

type AgentCapabilities struct {
	LoadSession bool                `json:"loadSession,omitempty"`
	PromptCaps  *PromptCapabilities `json:"promptCapabilities,omitempty"`
	MCPCaps     *MCPCapabilities    `json:"mcpCapabilities,omitempty"`
}

type PromptCapabilities struct {
	Image           bool `json:"image,omitempty"`
	Audio           bool `json:"audio,omitempty"`
	EmbeddedContext bool `json:"embeddedContext,omitempty"`
}

type MCPCapabilities struct {
	HTTP bool `json:"http,omitempty"`
	SSE  bool `json:"sse,omitempty"`
}

// ---- Session ----

type SessionNewRequest struct {
	CWD        string      `json:"cwd"`
	MCPServers []MCPServer `json:"mcpServers,omitempty"`
}

type SessionNewResponse struct {
	SessionID string `json:"sessionId"`
}

type MCPServer struct {
	Name    string   `json:"name"`
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
	Env     []string `json:"env,omitempty"`
}

type SessionLoadRequest struct {
	SessionID  string      `json:"sessionId"`
	CWD        string      `json:"cwd"`
	MCPServers []MCPServer `json:"mcpServers,omitempty"`
}

type SessionCancelRequest struct {
	SessionID string `json:"sessionId"`
}

// ---- Prompt ----

type SessionPromptRequest struct {
	SessionID string         `json:"sessionId"`
	Prompt    []ContentBlock `json:"prompt"`
	Model     string         `json:"model,omitempty"`
	Mode      string         `json:"mode,omitempty"`
	Policy    string         `json:"policy,omitempty"`
}

type PromptResponse struct {
	StopReason string `json:"stopReason"`
}

// Stop reasons
const (
	StopEndTurn   = "end_turn"
	StopMaxTokens = "max_tokens"
	StopMaxTurns  = "max_turn_requests"
	StopRefusal   = "refusal"
	StopCancelled = "cancelled"
)

// ---- Content Blocks ----

type ContentBlock struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	Resource *Resource `json:"resource,omitempty"`
}

type Resource struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
}

// ---- Session Updates ----

type SessionUpdate struct {
	SessionID string     `json:"sessionId"`
	Update    UpdateData `json:"update"`
}

type UpdateData struct {
	SessionUpdate string `json:"sessionUpdate"`

	// plan
	Entries []PlanEntry `json:"entries,omitempty"`

	// tool_call / tool_call_update
	ToolCallID string         `json:"toolCallId,omitempty"`
	Title      string         `json:"title,omitempty"`
	Kind       string         `json:"kind,omitempty"`
	Status     string         `json:"status,omitempty"`
	Content    []ToolContent  `json:"content,omitempty"`
	Locations  []ToolLocation `json:"locations,omitempty"`
	RawInput   map[string]any `json:"rawInput,omitempty"`
	RawOutput  map[string]any `json:"rawOutput,omitempty"`
}

type PlanEntry struct {
	Content  string `json:"content"`
	Priority string `json:"priority,omitempty"`
	Status   string `json:"status,omitempty"`
}

type ToolContent struct {
	Type    string        `json:"type"`
	Content *ContentBlock `json:"content,omitempty"`
}

type ToolLocation struct {
	Path string `json:"path"`
	Line *int   `json:"line,omitempty"`
}

// Tool call statuses
const (
	ToolStatusPending    = "pending"
	ToolStatusInProgress = "in_progress"
	ToolStatusCompleted  = "completed"
	ToolStatusFailed     = "failed"
)

// Tool kinds
const (
	ToolKindRead    = "read"
	ToolKindEdit    = "edit"
	ToolKindDelete  = "delete"
	ToolKindMove    = "move"
	ToolKindSearch  = "search"
	ToolKindExecute = "execute"
	ToolKindThink   = "think"
	ToolKindFetch   = "fetch"
	ToolKindOther   = "other"
)

// ---- Permission ----

type RequestPermissionRequest struct {
	SessionID string             `json:"sessionId"`
	ToolCall  PermissionToolCall `json:"toolCall"`
	Options   []PermissionOption `json:"options"`
}

type PermissionToolCall struct {
	ToolCallID string `json:"toolCallId"`
}

type PermissionOption struct {
	OptionID string `json:"optionId"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
}

// Permission kinds
const (
	PermAllowOnce    = "allow_once"
	PermAllowAlways  = "allow_always"
	PermRejectOnce   = "reject_once"
	PermRejectAlways = "reject_always"
)

type RequestPermissionResponse struct {
	Outcome PermissionOutcome `json:"outcome"`
}

type PermissionOutcome struct {
	Outcome  string `json:"outcome"`
	OptionID string `json:"optionId,omitempty"`
}

const (
	OutcomeSelected  = "selected"
	OutcomeCancelled = "cancelled"
)

// ---- File System ----

type FSReadRequest struct {
	SessionID string `json:"sessionId"`
	Path      string `json:"path"`
	Line      *int   `json:"line,omitempty"`
	Limit     *int   `json:"limit,omitempty"`
}

type FSReadResponse struct {
	Content string `json:"content"`
}

type FSWriteRequest struct {
	SessionID string `json:"sessionId"`
	Path      string `json:"path"`
	Content   string `json:"content"`
}

type FSCompletionRequest struct {
	SessionID string `json:"sessionId"`
	Path      string `json:"path"`
	Prefix    string `json:"prefix"`
	Suffix    string `json:"suffix,omitempty"`
	Language  string `json:"language,omitempty"`
}

type FSCompletionResponse struct {
	Content string `json:"content"`
}

// ---- Terminal ----

type TerminalCreateRequest struct {
	SessionID string   `json:"sessionId"`
	CWD       string   `json:"cwd"`
	Command   string   `json:"command"`
	Args      []string `json:"args,omitempty"`
	Env       []string `json:"env,omitempty"`
}

type TerminalCreateResponse struct {
	TerminalID string `json:"terminalId"`
}

type TerminalOutputRequest struct {
	SessionID  string `json:"sessionId"`
	TerminalID string `json:"terminalId"`
}

type TerminalOutputResponse struct {
	Content string `json:"content"`
}

type TerminalReleaseRequest struct {
	SessionID  string `json:"sessionId"`
	TerminalID string `json:"terminalId"`
}

type TerminalWaitForExitRequest struct {
	SessionID  string `json:"sessionId"`
	TerminalID string `json:"terminalId"`
}

type TerminalKillRequest struct {
	SessionID  string `json:"sessionId"`
	TerminalID string `json:"terminalId"`
}

// ---- Server ----

// Server implements the ACP JSON-RPC server over stdio.
type Server struct {
	engine   Engine
	sessions map[string]*Session
	out      io.Writer
}

// Engine is the interface the ACP server uses to execute operations.
// This is implemented by the vectora core engine.
type Engine interface {
	// Embed generates an embedding vector for the given text.
	Embed(ctx context.Context, text string) ([]float32, error)
	// Query performs an agentic query with specific model, mode and policy.
	Query(ctx context.Context, query string, workspaceID string, model string, mode string, policy string) (string, error)
	// ExecuteTool runs a tool by name with the given arguments.
	ExecuteTool(ctx context.Context, name string, args map[string]any) (ToolResult, error)
	// ReadFile reads a file at the given path.
	ReadFile(ctx context.Context, path string) (string, error)
	// WriteFile writes content to a file at the given path.
	WriteFile(ctx context.Context, path, content string) error
	// CompleteCode suggests code based on prefix and suffix.
	CompleteCode(ctx context.Context, path, prefix, suffix, language string) (string, error)
	// RunCommand executes a shell command.
	RunCommand(ctx context.Context, cwd, command string) (string, error)
	// GrepSearch performs a regex search across the codebase.
	GrepSearch(ctx context.Context, root, pattern string) (string, error)
	// Edit applies changes to a file using the edit tool.
	Edit(ctx context.Context, path, instruction, content string) (string, error)
	// WebSearch performs a web search.
	WebSearch(ctx context.Context, query string) (string, error)
	// WebFetch fetches content from a URL.
	WebFetch(ctx context.Context, url string) (string, error)
}

type ToolResult struct {
	Output  string `json:"output"`
	IsError bool   `json:"is_error"`
}

// Session represents an active ACP conversation session.
type Session struct {
	ID           string
	CWD          string
	Updates      chan SessionUpdate
	PermissionCh chan PermissionResponse
}

type PermissionResponse struct {
	OptionID  string
	Cancelled bool
}

// NewServer creates a new ACP server.
func NewServer(engine Engine) *Server {
	return &Server{
		engine:   engine,
		sessions: make(map[string]*Session),
		out:      os.Stdout,
	}
}

// NOTE: The Server.Run method has been replaced by StartACPAgent which uses
// the Coder ACP SDK. The custom Server implementation is kept for reference
// but is no longer used in Phase 7+.
// To use the new Coder SDK-based ACP, use StartACPAgent(ctx, agent, logger)
// instead of Server.Run(ctx).

// Run starts the ACP server reading from stdin and writing to stdout.
// This blocks until the connection is closed.
// DEPRECATED: Use StartACPAgent instead.
func (s *Server) Run(ctx context.Context) error {
	// RunStream is no longer implemented. Use StartACPAgent(ctx, agent, logger) instead.
	return fmt.Errorf("deprecated: use StartACPAgent instead")
}
