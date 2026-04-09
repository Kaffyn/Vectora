package acp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

// runStdioServer runs the JSON-RPC 2.0 server over stdin/stdout.
func runStdioServer(ctx context.Context, s *Server) error {
	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 10*1024*1024) // 10MB buffer
	scanner.Buffer(buf, len(buf))

	var idCounter int64

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var raw map[string]json.RawMessage
		if err := json.Unmarshal(line, &raw); err != nil {
			s.writeError(nil, -32700, "Parse error", err.Error())
			continue
		}

		// Check if it's a notification (no id)
		_, hasID := raw["id"]

		method, _ := raw["method"]
		methodStr := strings.Trim(string(method), "\"")

		params, _ := raw["params"]

		if hasID {
			var id int64
			if err := json.Unmarshal(raw["id"], &id); err == nil {
				atomic.StoreInt64(&idCounter, id)
			}
			go s.handleRequest(ctx, methodStr, params, raw["id"])
		} else {
			go s.handleNotification(ctx, methodStr, params)
		}
	}

	return scanner.Err()
}

func (s *Server) handleRequest(ctx context.Context, method string, params json.RawMessage, id json.RawMessage) {
	var result any
	var errMsg string

	switch method {
	case "initialize":
		result, errMsg = s.handleInitialize(ctx, params)
	case "session/new":
		result, errMsg = s.handleSessionNew(ctx, params)
	case "session/load":
		result, errMsg = s.handleSessionLoad(ctx, params)
	case "session/prompt":
		result, errMsg = s.handleSessionPrompt(ctx, params)
	case "fs/read_text_file":
		result, errMsg = s.handleFSRead(ctx, params)
	case "fs/write_text_file":
		result, errMsg = s.handleFSWrite(ctx, params)
	default:
		s.writeError(id, -32601, "Method not found", fmt.Sprintf("Method '%s' is not supported", method))
		return
	}

	if errMsg != "" {
		s.writeError(id, -32603, "Internal error", errMsg)
		return
	}

	s.writeResult(id, result)
}

func (s *Server) handleNotification(ctx context.Context, method string, params json.RawMessage) {
	switch method {
	case "session/cancel":
		s.handleSessionCancel(ctx, params)
	}
}

// ---- Initialize ----

func (s *Server) handleInitialize(ctx context.Context, params json.RawMessage) (any, string) {
	var req InitializeRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Sprintf("Invalid initialize params: %v", err)
	}

	negotiatedVersion := req.ProtocolVersion
	if negotiatedVersion > ProtocolVersion {
		negotiatedVersion = ProtocolVersion
	}

	resp := InitializeResponse{
		ProtocolVersion: negotiatedVersion,
		AgentInfo: &Info{
			Name:    "vectora",
			Title:   "Vectora AI Agent",
			Version: "0.1.0",
		},
		AgentCapabilities: &AgentCapabilities{
			LoadSession: true,
			PromptCaps: &PromptCapabilities{
				Image:           false,
				Audio:           false,
				EmbeddedContext: true,
			},
			MCPCaps: &MCPCapabilities{
				HTTP: false,
				SSE:  false,
			},
		},
		AuthMethods: []string{},
	}

	return resp, ""
}

// ---- Session New ----

func (s *Server) handleSessionNew(ctx context.Context, params json.RawMessage) (any, string) {
	var req SessionNewRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Sprintf("Invalid session/new params: %v", err)
	}

	sessionID := "sess_" + uuid.New().String()
	session := &Session{
		ID:           sessionID,
		CWD:          req.CWD,
		Updates:      make(chan SessionUpdate, 100),
		PermissionCh: make(chan PermissionResponse, 1),
	}

	s.sessions[sessionID] = session

	return SessionNewResponse{SessionID: sessionID}, ""
}

// ---- Session Load ----

func (s *Server) handleSessionLoad(ctx context.Context, params json.RawMessage) (any, string) {
	var req SessionLoadRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Sprintf("Invalid session/load params: %v", err)
	}

	// Check if session exists (from previous connection)
	if _, ok := s.sessions[req.SessionID]; ok {
		return nil, "" // null response — session replayed via updates
	}

	// For now, create a new session with the loaded ID
	// In production, this would replay conversation history
	session := &Session{
		ID:           req.SessionID,
		CWD:          req.CWD,
		Updates:      make(chan SessionUpdate, 100),
		PermissionCh: make(chan PermissionResponse, 1),
	}
	s.sessions[req.SessionID] = session

	return nil, ""
}

// ---- Session Prompt ----

func (s *Server) handleSessionPrompt(ctx context.Context, params json.RawMessage) (any, string) {
	var req SessionPromptRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Sprintf("Invalid session/prompt params: %v", err)
	}

	session, ok := s.sessions[req.SessionID]
	if !ok {
		return nil, fmt.Sprintf("Session '%s' not found", req.SessionID)
	}

	// Extract text from prompt blocks
	var queryText string
	for _, block := range req.Prompt {
		if block.Type == "text" {
			queryText += block.Text
		} else if block.Type == "resource" && block.Resource != nil {
			// Include resource content in query
			queryText += "\n\n[Resource: " + block.Resource.URI + "]\n" + block.Resource.Text
		}
	}

	// Send initial plan update
	s.sendUpdate(session, UpdateData{
		SessionUpdate: "plan",
		Entries: []PlanEntry{
			{Content: "Processing query...", Priority: "high", Status: "in_progress"},
		},
	})

	// Execute the query via engine
	answer, err := s.engine.Query(ctx, queryText, "default")
	if err != nil {
		s.sendUpdate(session, UpdateData{
			SessionUpdate: "plan",
			Entries: []PlanEntry{
				{Content: "Query failed: " + err.Error(), Priority: "high", Status: "failed"},
			},
		})
		return PromptResponse{StopReason: StopEndTurn}, ""
	}

	// Stream the response as text chunks
	chunkSize := 200
	for i := 0; i < len(answer); i += chunkSize {
		end := i + chunkSize
		if end > len(answer) {
			end = len(answer)
		}
		chunk := answer[i:end]

		s.sendUpdate(session, UpdateData{
			SessionUpdate: "agent_message_chunk",
			Content: []ToolContent{
				{
					Type: "content",
					Content: &ContentBlock{
						Type: "text",
						Text: chunk,
					},
				},
			},
		})
	}

	return PromptResponse{StopReason: StopEndTurn}, ""
}

// ---- Session Cancel ----

func (s *Server) handleSessionCancel(ctx context.Context, params json.RawMessage) {
	var req SessionCancelRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return
	}

	if session, ok := s.sessions[req.SessionID]; ok {
		session.PermissionCh <- PermissionResponse{Cancelled: true}
	}
}

// ---- File System ----

func (s *Server) handleFSRead(ctx context.Context, params json.RawMessage) (any, string) {
	var req FSReadRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Sprintf("Invalid fs/read_text_file params: %v", err)
	}

	content, err := s.engine.ReadFile(ctx, req.Path)
	if err != nil {
		return nil, fmt.Sprintf("Failed to read file: %v", err)
	}

	// Apply line/limit if specified
	if req.Line != nil || req.Limit != nil {
		lines := strings.Split(content, "\n")
		startLine := 0
		if req.Line != nil {
			startLine = *req.Line - 1 // 1-based
			if startLine < 0 {
				startLine = 0
			}
			if startLine >= len(lines) {
				return FSReadResponse{Content: ""}, ""
			}
		}

		endLine := len(lines)
		if req.Limit != nil {
			endLine = startLine + *req.Limit
			if endLine > len(lines) {
				endLine = len(lines)
			}
		}

		content = strings.Join(lines[startLine:endLine], "\n")
	}

	return FSReadResponse{Content: content}, ""
}

func (s *Server) handleFSWrite(ctx context.Context, params json.RawMessage) (any, string) {
	var req FSWriteRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Sprintf("Invalid fs/write_text_file params: %v", err)
	}

	err := s.engine.WriteFile(ctx, req.Path, req.Content)
	if err != nil {
		return nil, fmt.Sprintf("Failed to write file: %v", err)
	}

	return nil, ""
}

// ---- Helper Methods ----

func (s *Server) sendUpdate(session *Session, data UpdateData) {
	update := SessionUpdate{
		SessionID: session.ID,
		Update:    data,
	}
	select {
	case session.Updates <- update:
		// Also send to stdout for the client to receive
		s.sendSessionUpdateStdout(update)
	default:
		// Drop if buffer full
	}
}

// JSON-RPC response writing

type jsonResponse struct {
	JSONRPC string     `json:"jsonrpc"`
	ID      any        `json:"id"`
	Result  any        `json:"result,omitempty"`
	Error   *jsonError `json:"error,omitempty"`
}

type jsonError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

var outputMu sync.Mutex

func (s *Server) writeResult(id json.RawMessage, result any) {
	var idVal any
	if id != nil {
		json.Unmarshal(id, &idVal)
	}

	resp := jsonResponse{
		JSONRPC: "2.0",
		ID:      idVal,
		Result:  result,
	}

	outputMu.Lock()
	defer outputMu.Unlock()

	data, _ := json.Marshal(resp)
	fmt.Fprintln(os.Stdout, string(data))
}

func (s *Server) writeError(id json.RawMessage, code int, message string, data string) {
	var idVal any
	if id != nil {
		json.Unmarshal(id, &idVal)
	}

	resp := jsonResponse{
		JSONRPC: "2.0",
		ID:      idVal,
		Error: &jsonError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	outputMu.Lock()
	defer outputMu.Unlock()

	dataBytes, _ := json.Marshal(resp)
	fmt.Fprintln(os.Stdout, string(dataBytes))
}

// sendNotification writes a JSON-RPC notification to stdout (no id).
func sendNotification(method string, params any) {
	notif := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}

	outputMu.Lock()
	defer outputMu.Unlock()

	data, _ := json.Marshal(notif)
	fmt.Fprintln(os.Stdout, string(data))
}

// sendSessionUpdate sends a session/update notification to stdout.
func (s *Server) sendSessionUpdateStdout(update SessionUpdate) {
	sendNotification("session/update", update)
}

// sendPermissionRequest sends a session/request_permission to stdout and waits for response.
func (s *Server) sendPermissionRequest(ctx context.Context, session *Session, req RequestPermissionRequest) (*RequestPermissionResponse, error) {
	// Send the request
	outputMu.Lock()
	notifData := map[string]any{
		"jsonrpc": "2.0",
		"method":  "session/request_permission",
		"params":  req,
	}
	data, _ := json.Marshal(notifData)
	fmt.Fprintln(os.Stdout, string(data))
	outputMu.Unlock()

	// Wait for response
	select {
	case resp := <-session.PermissionCh:
		if resp.Cancelled {
			return &RequestPermissionResponse{
				Outcome: PermissionOutcome{Outcome: OutcomeCancelled},
			}, nil
		}
		return &RequestPermissionResponse{
			Outcome: PermissionOutcome{Outcome: OutcomeSelected, OptionID: resp.OptionID},
		}, nil
	case <-ctx.Done():
		return &RequestPermissionResponse{
			Outcome: PermissionOutcome{Outcome: OutcomeCancelled},
		}, ctx.Err()
	}
}
