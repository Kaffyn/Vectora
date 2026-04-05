package acp

import (
	"context"
	"encoding/json"
	"fmt"
)

type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, args json.RawMessage) (interface{}, error)
}

type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

func (r *Registry) Register(tool Tool) error {
	r.tools[tool.Name()] = tool
	return nil
}

func (r *Registry) Execute(ctx context.Context, toolName string, args json.RawMessage) (interface{}, error) {
	tool, ok := r.tools[toolName]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	return tool.Execute(ctx, args)
}

func (r *Registry) List() []string {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}
