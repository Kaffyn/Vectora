package methods

import (
	"context"
	"encoding/json"
	"vectora/core/tools"
)

type ToolCallRequest struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func HandleToolsCall(ctx context.Context, registry *tools.Registry, req ToolCallRequest) (*tools.ToolResult, error) {
	tool, ok := registry.GetTool(req.Name)
	if !ok {
		return &tools.ToolResult{Output: "Tool not found", IsError: true}, nil
	}

	// Execução da ferramenta
	result, err := tool.Execute(ctx, req.Arguments)
	if err != nil {
		return &tools.ToolResult{Output: err.Error(), IsError: true}, nil
	}

	// Sanitização final (opcional, já feita nas tools)
	return result, nil
}
