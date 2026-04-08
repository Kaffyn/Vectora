package engine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/ingestion"
	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/Kaffyn/Vectora/core/policies"
	"github.com/Kaffyn/Vectora/core/tools"
)

// Engine is the central orchestrator for Vectora.
type Engine struct {
	Storage  *db.ChromemStore
	KV       *db.BBoltStore
	LLM      *llm.Router
	Tools    *tools.Registry
	Guardian *policies.Guardian
	Indexer  *ingestion.Indexer
	Status   string
}

func NewEngine(
	vecStore *db.ChromemStore,
	kvStore *db.BBoltStore,
	llmRouter *llm.Router,
	toolsReg *tools.Registry,
	guardian *policies.Guardian,
	indexer *ingestion.Indexer,
) *Engine {
	return &Engine{
		Storage:  vecStore,
		KV:       kvStore,
		LLM:      llmRouter,
		Tools:    toolsReg,
		Guardian: guardian,
		Indexer:  indexer,
		Status:   "idle",
	}
}

// QueryChunk represents a streaming chunk of response.
type QueryChunk struct {
	Token   string
	Sources []db.ScoredChunk
	IsFinal bool
}

// ToolCallRequest is the request format for tool execution.
type ToolCallRequest struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ExecuteTool delegates to the tools registry with Guardian validation.
func (e *Engine) ExecuteTool(ctx context.Context, req ToolCallRequest) (*tools.ToolResult, error) {
	tool, ok := e.Tools.GetTool(req.Name)
	if !ok {
		return nil, fmt.Errorf("tool %s not found", req.Name)
	}
	return tool.Execute(ctx, req.Arguments)
}

// StreamQuery executes a RAG query and streams back results.
func (e *Engine) StreamQuery(ctx context.Context, query string, workspaceID string) (chan QueryChunk, error) {
	ch := make(chan QueryChunk)

	go func() {
		defer close(ch)

		provider := e.LLM.GetDefault()
		if provider == nil || !provider.IsConfigured() {
			ch <- QueryChunk{Token: "No LLM provider configured.", IsFinal: true}
			return
		}

		// 1. Embed query
		vector, err := provider.Embed(ctx, query)
		if err != nil {
			ch <- QueryChunk{Token: fmt.Sprintf("Embed error: %v", err), IsFinal: true}
			return
		}

		// 2. Retrieve from vector store
		chunks, err := e.Storage.Query(ctx, "ws_"+workspaceID, vector, 5)
		if err != nil {
			chunks = []db.ScoredChunk{}
		}

		// 3. Build context
		contextText := ""
		for _, doc := range chunks {
			if filename, ok := doc.Metadata["filename"]; ok {
				contextText += "File: " + filename + "\n"
			}
			contextText += doc.Content + "\n---\n"
		}

		// 4. Build tool definitions for LLM
		var toolDefs []llm.ToolDefinition
		for _, t := range e.Tools.GetAll() {
			toolDefs = append(toolDefs, llm.ToolDefinition{
				Name:        t.Name(),
				Description: t.Description(),
				Schema:      t.Schema(),
			})
		}

		// 5. Complete
		messages := []llm.Message{
			{Role: llm.RoleSystem, Content: "You are Vectora. Use the following context as your source of truth:\n" + contextText},
			{Role: llm.RoleUser, Content: query},
		}

		resp, err := provider.Complete(ctx, llm.CompletionRequest{
			Messages:    messages,
			MaxTokens:   1500,
			Temperature: 0.1,
			Tools:       toolDefs,
		})
		if err != nil {
			ch <- QueryChunk{Token: fmt.Sprintf("LLM error: %v", err), IsFinal: true}
			return
		}

		// 6. Handle tool calls
		if len(resp.ToolCalls) > 0 {
			for _, tc := range resp.ToolCalls {
				result, err := e.Tools.ExecuteStringArgs(ctx, tc.Name, tc.Args)
				if err != nil {
					ch <- QueryChunk{Token: fmt.Sprintf("\n[Tool %s error: %v]", tc.Name, err), IsFinal: false}
				} else {
					ch <- QueryChunk{Token: fmt.Sprintf("\n[Tool %s]: %s", tc.Name, result.Output), IsFinal: false}
				}
			}
		}

		ch <- QueryChunk{Token: resp.Content, Sources: chunks, IsFinal: true}
	}()

	return ch, nil
}

// ProcessQuery is the synchronous version of StreamQuery.
func (e *Engine) ProcessQuery(query string, workspaceID string) (string, error) {
	ctx := context.Background()
	ch, err := e.StreamQuery(ctx, query, workspaceID)
	if err != nil {
		return "", err
	}
	var result string
	for chunk := range ch {
		result += chunk.Token
	}
	return result, nil
}

// StartIndexation triggers the indexing pipeline.
func (e *Engine) StartIndexation(ctx context.Context, rootPath string) error {
	e.Status = "indexing"
	if e.Indexer == nil {
		e.Status = "idle"
		return fmt.Errorf("indexer not initialized")
	}
	err := e.Indexer.IndexDirectory(ctx, rootPath)
	e.Status = "idle"
	return err
}

// GetStatus returns the current engine status.
func (e *Engine) GetStatus() string {
	return e.Status
}
