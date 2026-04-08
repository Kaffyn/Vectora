package handlers

import (
	"context"
	"encoding/json"

	"github.com/Kaffyn/Vectora/core/engine"
)

type ToolCallRequest struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ToolCallResponse struct {
	Content []map[string]string `json:"content"`
	IsError bool                `json:"isError"`
}

func HandleToolsCall(ctx context.Context, eng *engine.Engine, req ToolCallRequest) (*ToolCallResponse, error) {
	engineReq := engine.ToolCallRequest{
		Name:      req.Name,
		Arguments: req.Arguments,
	}
	result, err := eng.ExecuteTool(ctx, engineReq)
	if err != nil {
		return &ToolCallResponse{
			Content: []map[string]string{{"type": "text", "text": err.Error()}},
			IsError: true,
		}, nil
	}

	return &ToolCallResponse{
		Content: []map[string]string{{"type": "text", "text": result.Output}},
		IsError: result.IsError,
	}, nil
}

// Stubs para compilação pacifica:
type InitRequest struct{}
type InitResponse struct{}
type ToolsListResponse struct{}

func HandleInitialize(req InitRequest) InitResponse { return InitResponse{} }
func HandleToolsList() ToolsListResponse            { return ToolsListResponse{} }
