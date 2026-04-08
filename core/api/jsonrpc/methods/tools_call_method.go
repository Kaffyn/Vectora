package methods

import (
	"context"
	"encoding/json"

	"vectora/core/api/shared"
	"vectora/core/engine"
)

// HandleToolsCall funciona isolado de protocolos parseando puramente bytes do stdin JSON-RPC MCP
func HandleToolsCall(ctx context.Context, deps *shared.CoreDeps, params json.RawMessage) (interface{}, error) {
	var req engine.ToolCallRequest
	
	// Simulação de unmarshall flexível do protocolo aberto (MCP tool call struct)
	var r map[string]interface{}
	if err := json.Unmarshal(params, &r); err != nil {
		return nil, err
	}
	req.Name = r["name"].(string)

	// Conexão direta com a inteligência do Vectora, isolado da camada STDIO
	result, err := deps.Engine.ExecuteTool(ctx, req)
	if err != nil {
		return nil, err
	}

	// Formatado de volta no Standard JSON-RPC esperado pela IDE
	return map[string]interface{}{
		"content": []map[string]string{{"type": "text", "text": result.Output}},
		"isError": result.Error != nil,
	}, nil
}
