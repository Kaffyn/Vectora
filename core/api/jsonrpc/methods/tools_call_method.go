// Package methods provides JSON-RPC method handlers for the Vectora API.
package methods

import (
	"context"
	"encoding/json"

	"github.com/Kaffyn/Vectora/core/api/jsonrpc"
	"github.com/Kaffyn/Vectora/core/api/shared"
)

// HandleToolsCall processes a tool call via the shared CoreDeps.
func HandleToolsCall(deps *shared.CoreDeps, params json.RawMessage) (interface{}, error) {
	var req struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, jsonrpc.NewError(-32602, "Invalid params")
	}

	result, err := deps.Engine.ExecuteTool(context.Background(), req.Name, req.Arguments)
	if err != nil {
		return nil, jsonrpc.NewError(-32000, err.Error())
	}

	if result.IsError {
		return nil, jsonrpc.NewErrorWithData(-32001, result.Output, map[string]string{
			"tool": req.Name,
		})
	}

	return map[string]interface{}{
		"content": []map[string]string{{"type": "text", "text": result.Output}},
		"isError": false,
	}, nil
}
