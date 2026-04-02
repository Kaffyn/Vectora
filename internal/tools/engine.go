package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

type ToolResult struct {
	Output     string `json:"output"`
	SnapshotID string `json:"snapshot_id,omitempty"`
	IsError    bool   `json:"is_error"`
}

type Tool interface {
	Name() string
	Description() string
	Schema() json.RawMessage
	Execute(ctx context.Context, args map[string]any) (ToolResult, error)
}

type Registry struct {
	registry map[string]Tool
}

// Inicia um pool isolado de Ferramentas disponíveis pro LLM
func NewRegistry() *Registry {
	return &Registry{
		registry: make(map[string]Tool),
	}
}

func (r *Registry) Register(t Tool) {
	r.registry[t.Name()] = t
}

func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.registry[name]
	return t, ok
}

func (r *Registry) GetAll() []Tool {
	var all []Tool
	for _, t := range r.registry {
		all = append(all, t)
	}
	return all
}

// Transforma diretamente invocações puras de Texto JSON (que o LLM vomita)
// em Execuções Binárias seguras
func (r *Registry) ExecuteStringArgs(ctx context.Context, name string, argsJSON string) (ToolResult, error) {
	t, exists := r.registry[name]
	if !exists {
		return ToolResult{}, fmt.Errorf("agent_tool_err: ferramenta '%s' inexistente no arsenal local", name)
	}

	var args map[string]any
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return ToolResult{IsError: true, Output: "Argumentos inválidos ou malformados via JSON: " + err.Error()}, nil
	}

	return t.Execute(ctx, args)
}
