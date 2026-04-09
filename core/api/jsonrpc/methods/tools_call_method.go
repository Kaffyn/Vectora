// Package methods provides JSON-RPC method handlers for the Vectora API.
package methods

import (
	"encoding/json"

	"github.com/Kaffyn/Vectora/core/api/jsonrpc"
	"github.com/Kaffyn/Vectora/core/api/shared"
	"github.com/Kaffyn/Vectora/core/engine"
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

	// Convert arguments back to JSON for the engine
	argsJSON, _ := json.Marshal(req.Arguments)

	result, err := deps.Engine.ExecuteTool(nil, engine.ToolCallRequest{
		Name:      req.Name,
		Arguments: argsJSON,
	})
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
