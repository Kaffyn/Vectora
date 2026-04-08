package handlers

import (
	"context"
	"encoding/json"
	"vectora/core/engine"
	"vectora/core/tools"
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
	// Delega para o executor de ferramentas do Engine
	result, err := eng.ExecuteTool(ctx, req.Name, req.Arguments)
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
func HandleToolsList() ToolsListResponse { return ToolsListResponse{} }
