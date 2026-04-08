package methods

import (
	"context"
	"encoding/json"

	"github.com/Kaffyn/Vectora/core/api/shared"
	"github.com/Kaffyn/Vectora/core/engine"
)

func HandleToolsCall(ctx context.Context, deps *shared.CoreDeps, params json.RawMessage) (interface{}, error) {
	var r map[string]interface{}
	if err := json.Unmarshal(params, &r); err != nil {
		return nil, err
	}

	name, _ := r["name"].(string)
	argsBytes, _ := json.Marshal(r["arguments"])

	req := engine.ToolCallRequest{
		Name:      name,
		Arguments: json.RawMessage(argsBytes),
	}

	result, err := deps.Engine.ExecuteTool(ctx, req)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"content": []map[string]string{{"type": "text", "text": result.Output}},
		"isError": result.IsError,
	}, nil
}
