package mocks

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// MockDaemon represents a mock daemon for testing
type MockDaemon struct {
	listener    net.Listener
	mu          sync.RWMutex
	models      []ModelInfo
	indices     []IndexInfo
	config      map[string]interface{}
	chatHistory []ChatMessage
	activeModel string
	running     bool
	port        int
}

// ModelInfo represents model information
type ModelInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
	Size   string `json:"size"`
	Type   string `json:"type"`
}

// IndexInfo represents index information
type IndexInfo struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	DocumentCount int       `json:"document_count"`
	Size          int64     `json:"size"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Model   string `json:"model"`
}

// SystemHealth represents system health information
type SystemHealth struct {
	Status     string `json:"status"`
	Uptime     int64  `json:"uptime"`
	ModelsPath string `json:"models_path"`
	Version    string `json:"version"`
}

// NewMockDaemon creates a new mock daemon
func NewMockDaemon() *MockDaemon {
	md := &MockDaemon{
		models: []ModelInfo{
			{ID: "gpt-4", Name: "GPT-4", Active: true, Size: "100GB", Type: "chat"},
			{ID: "gpt-3.5", Name: "GPT-3.5", Active: false, Size: "50GB", Type: "chat"},
			{ID: "claude-3", Name: "Claude 3", Active: false, Size: "80GB", Type: "chat"},
		},
		indices: []IndexInfo{},
		config: map[string]interface{}{
			"theme":      "light",
			"font_size":  "medium",
			"auto_save":  true,
		},
		chatHistory: []ChatMessage{},
		activeModel: "gpt-4",
		port:        9999,
	}
	return md
}

// Start starts the mock daemon
func (md *MockDaemon) Start() error {
	md.mu.Lock()
	defer md.mu.Unlock()

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", md.port))
	if err != nil {
		return err
	}

	md.listener = listener
	md.running = true

	go md.acceptConnections()
	return nil
}

// Stop stops the mock daemon
func (md *MockDaemon) Stop() error {
	md.mu.Lock()
	defer md.mu.Unlock()

	md.running = false
	if md.listener != nil {
		return md.listener.Close()
	}
	return nil
}

// acceptConnections accepts incoming connections
func (md *MockDaemon) acceptConnections() {
	for md.running {
		conn, err := md.listener.Accept()
		if err != nil {
			if md.running {
				continue
			}
			return
		}
		go md.handleConnection(conn)
	}
}

// handleConnection handles a single client connection
func (md *MockDaemon) handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var request map[string]interface{}
		if err := decoder.Decode(&request); err != nil {
			return
		}

		method := request["method"].(string)
		response := md.handleRequest(method, request)

		encoder.Encode(response)
	}
}

// handleRequest processes a request and returns a response
func (md *MockDaemon) handleRequest(method string, request map[string]interface{}) map[string]interface{} {
	md.mu.RLock()
	defer md.mu.RUnlock()

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      request["id"],
	}

	switch method {
	case "chat.send":
		params := request["params"].(map[string]interface{})
		message := params["message"].(string)
		response["result"] = map[string]string{
			"response": fmt.Sprintf("Echo: %s", message),
			"model":    md.activeModel,
		}

	case "models.list":
		response["result"] = md.models

	case "models.active":
		response["result"] = map[string]string{
			"id":     md.activeModel,
			"active": md.activeModel,
		}

	case "models.set":
		params := request["params"].(map[string]interface{})
		modelID := params["model"].(string)
		md.setActiveModel(modelID)
		response["result"] = map[string]bool{"success": true}

	case "index.list":
		response["result"] = md.indices

	case "index.add":
		params := request["params"].(map[string]interface{})
		name := params["name"].(string)
		newIndex := IndexInfo{
			ID:        fmt.Sprintf("idx-%d", len(md.indices)),
			Name:      name,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		md.indices = append(md.indices, newIndex)
		response["result"] = map[string]bool{"success": true}

	case "index.remove":
		params := request["params"].(map[string]interface{})
		indexID := params["id"].(string)
		for i, idx := range md.indices {
			if idx.ID == indexID {
				md.indices = append(md.indices[:i], md.indices[i+1:]...)
				break
			}
		}
		response["result"] = map[string]bool{"success": true}

	case "system.health":
		response["result"] = SystemHealth{
			Status:     "running",
			Uptime:     3600,
			ModelsPath: "/home/user/.vectora/models",
			Version:    "1.0.0",
		}

	case "config.get":
		response["result"] = md.config

	case "config.set":
		params := request["params"].(map[string]interface{})
		for k, v := range params {
			if k != "id" && k != "jsonrpc" && k != "method" {
				md.config[k] = v
			}
		}
		response["result"] = map[string]bool{"success": true}

	default:
		response["error"] = map[string]interface{}{
			"code":    -32601,
			"message": "Method not found",
		}
	}

	return response
}

// setActiveModel sets the active model
func (md *MockDaemon) setActiveModel(modelID string) {
	for i := range md.models {
		md.models[i].Active = false
	}
	for i := range md.models {
		if md.models[i].ID == modelID {
			md.models[i].Active = true
			md.activeModel = modelID
			break
		}
	}
}

// Chat sends a chat message
func (md *MockDaemon) Chat(msg string) (string, error) {
	md.mu.RLock()
	defer md.mu.RUnlock()

	message := ChatMessage{
		Role:    "user",
		Content: msg,
		Model:   md.activeModel,
	}
	md.chatHistory = append(md.chatHistory, message)

	response := ChatMessage{
		Role:    "assistant",
		Content: fmt.Sprintf("Echo: %s", msg),
		Model:   md.activeModel,
	}
	md.chatHistory = append(md.chatHistory, response)

	return response.Content, nil
}

// ListModels returns the list of models
func (md *MockDaemon) ListModels() ([]ModelInfo, error) {
	md.mu.RLock()
	defer md.mu.RUnlock()
	return md.models, nil
}

// SetModel sets the active model
func (md *MockDaemon) SetModel(id string) error {
	md.mu.Lock()
	defer md.mu.Unlock()

	for i := range md.models {
		if md.models[i].ID == id {
			md.setActiveModel(id)
			return nil
		}
	}
	return fmt.Errorf("model not found: %s", id)
}

// IsRunning returns whether the daemon is running
func (md *MockDaemon) IsRunning() bool {
	md.mu.RLock()
	defer md.mu.RUnlock()
	return md.running
}
