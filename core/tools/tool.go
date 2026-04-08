package tools

import (
	"context"
	"encoding/json"
)

// ToolResult é a resposta padronizada para o LLM
type ToolResult struct {
	Output   string                 `json:"output"`
	IsError  bool                   `json:"is_error"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Tool define o contrato básico para execução agêntica
type Tool interface {
	Name() string
	Description() string // Usado para gerar o JSON Schema do MCP
	Schema() string      // JSON Schema string dos argumentos esperados
	Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error)
}
