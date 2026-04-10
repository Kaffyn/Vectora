package acp

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// StatefulMockEngine implements the Engine interface with in-memory persistence.
// It is used for integration testing of the ACP server.
type StatefulMockEngine struct {
	mu       sync.RWMutex
	Files    map[string]string
	Sessions map[string][]string
	Delay    time.Duration
}

func NewStatefulMockEngine() *StatefulMockEngine {
	return &StatefulMockEngine{
		Files:    make(map[string]string),
		Sessions: make(map[string][]string),
		Delay:    100 * time.Millisecond,
	}
}

func (m *StatefulMockEngine) Embed(ctx context.Context, text string) ([]float32, error) {
	return make([]float32, 128), nil
}

func (m *StatefulMockEngine) Query(ctx context.Context, query string, workspaceID string, model string, mode string, policy string) (string, error) {
	time.Sleep(m.Delay)

	m.mu.Lock()
	m.Sessions[workspaceID] = append(m.Sessions[workspaceID], query)
	m.mu.Unlock()

	// Keyword-based response logic for smarter testing
	if strings.Contains(strings.ToLower(query), "hello") {
		return "Hello from Vectora Stateful Mock!", nil
	}

	if strings.Contains(strings.ToLower(query), "list files") {
		m.mu.RLock()
		defer m.mu.RUnlock()
		var names []string
		for k := range m.Files {
			names = append(names, k)
		}
		if len(names) == 0 {
			return "No files in memory.", nil
		}
		return "Files: " + strings.Join(names, ", "), nil
	}

	return fmt.Sprintf("Mock response to: %s", query), nil
}

func (m *StatefulMockEngine) ExecuteTool(ctx context.Context, name string, args map[string]any) (ToolResult, error) {
	switch name {
	case "read_file":
		content, err := m.ReadFile(ctx, args["path"].(string))
		if err != nil {
			return ToolResult{Output: err.Error(), IsError: true}, nil
		}
		return ToolResult{Output: content, IsError: false}, nil
	case "write_file":
		err := m.WriteFile(ctx, args["path"].(string), args["content"].(string))
		if err != nil {
			return ToolResult{Output: err.Error(), IsError: true}, nil
		}
		return ToolResult{Output: "Success", IsError: false}, nil
	}
	return ToolResult{Output: "Tool executed on mock", IsError: false}, nil
}

func (m *StatefulMockEngine) ReadFile(ctx context.Context, path string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	content, ok := m.Files[path]
	if !ok {
		return "", fmt.Errorf("file not found: %s", path)
	}
	return content, nil
}

func (m *StatefulMockEngine) WriteFile(ctx context.Context, path, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Files[path] = content
	return nil
}

func (m *StatefulMockEngine) CompleteCode(ctx context.Context, path, prefix, suffix, language string) (string, error) {
	return "// Mock completion", nil
}

func (m *StatefulMockEngine) RunCommand(ctx context.Context, cwd, command string) (string, error) {
	return fmt.Sprintf("Mock command output for: %s", command), nil
}

func (m *StatefulMockEngine) GrepSearch(ctx context.Context, root, pattern string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var results []string
	for path, content := range m.Files {
		if strings.Contains(content, pattern) {
			results = append(results, path)
		}
	}
	return strings.Join(results, "\n"), nil
}

func (m *StatefulMockEngine) Edit(ctx context.Context, path, instruction, content string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Files[path] = content // Simple overwrite for mock
	return "Changes applied to mock", nil
}

func (m *StatefulMockEngine) WebSearch(ctx context.Context, query string) (string, error) {
	return "Mock web results", nil
}

func (m *StatefulMockEngine) WebFetch(ctx context.Context, url string) (string, error) {
	return "Mock content from " + url, nil
}
