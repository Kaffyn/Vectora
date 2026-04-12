package engine

import (
	"context"
	"encoding/json"
	"fmt"

	"strings"

	"github.com/Kaffyn/Vectora/core/api/acp"
	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/infra"
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
func (e *Engine) ExecuteTool(ctx context.Context, name string, args map[string]any) (acp.ToolResult, error) {
	tool, ok := e.Tools.GetTool(name)
	if !ok {
		return acp.ToolResult{Output: fmt.Sprintf("Tool %s not found", name), IsError: true}, nil
	}

	argsJSON, _ := json.Marshal(args)
	result, err := tool.Execute(ctx, argsJSON)
	if err != nil {
		return acp.ToolResult{Output: err.Error(), IsError: true}, nil
	}

	return acp.ToolResult{Output: result.Output, IsError: result.IsError}, nil
}

// Embed generates an embedding vector using the default provider.
func (e *Engine) Embed(ctx context.Context, text string, model string) ([]float32, error) {
	provider := e.LLM.GetDefault()
	if provider == nil {
		return nil, fmt.Errorf("no LLM provider configured")
	}
	return provider.Embed(ctx, text, model)
}

// Query is the entry point for the ACP server. It uses the agentic StreamQuery loop.
func (e *Engine) Query(ctx context.Context, query string, workspaceID string, model string, mode string, policy string) (string, error) {
	ch, err := e.StreamQuery(ctx, query, workspaceID, model, mode, policy)
	if err != nil {
		return "", err
	}

	var finalAnswer string
	for chunk := range ch {
		if chunk.IsFinal {
			finalAnswer = chunk.Token
		}
	}
	return finalAnswer, nil
}

// StreamQuery executes an agentic query with multi-turn tool support.
func (e *Engine) StreamQuery(ctx context.Context, query string, workspaceID string, model string, mode string, policy string) (chan QueryChunk, error) {
	ch := make(chan QueryChunk)

	go func() {
		defer close(ch)

		// Use preferences for model if not specified
		if model == "" {
			prefs := infra.LoadPreferences()
			model = prefs.ActiveModel
		}

		provider := e.LLM.GetDefault()

		// If a model is specified, try to find the provider for it
		if model != "" {
			if strings.HasPrefix(model, "claude") || strings.HasPrefix(model, "4.6") {
				if p, err := e.LLM.GetProvider("claude"); err == nil && p.IsConfigured() {
					provider = p
				}
			} else if strings.HasPrefix(model, "gemini") {
				if p, err := e.LLM.GetProvider("gemini"); err == nil && p.IsConfigured() {
					provider = p
				}
			}
		}

		if provider == nil || !provider.IsConfigured() {
			ch <- QueryChunk{Token: "No LLM provider configured.", IsFinal: true}
			return
		}

		// 1. Embed query for RAG
		vector, err := provider.Embed(ctx, query, model)
		if err != nil {
			ch <- QueryChunk{Token: fmt.Sprintf("Embed error: %v", err), IsFinal: true}
			return
		}

		// 2. Retrieve context
		chunks, err := e.Storage.Query(ctx, "ws_"+workspaceID, vector, 5)
		if err != nil {
			chunks = []db.ScoredChunk{}
		}

		contextText := ""
		for _, doc := range chunks {
			if filename, ok := doc.Metadata["filename"]; ok {
				contextText += "File: " + filename + "\n"
			}
			contextText += doc.Content + "\n---\n"
		}

		// 3. Prepare Tool Definitions
		var toolDefs []llm.ToolDefinition
		for _, t := range e.Tools.GetAll() {
			toolDefs = append(toolDefs, llm.ToolDefinition{
				Name:        t.Name(),
				Description: t.Description(),
				Schema:      t.Schema(),
			})
		}

		// 4. Agent Loop (Multi-turn)
		messages := []llm.Message{
			{Role: llm.RoleSystem, Content: "You are Vectora, an expert AI coding assistant. Use the provided tools to fulfill requests. Trust Folder: " + e.Tools.TrustFolder + "\n\nContext:\n" + contextText},
			{Role: llm.RoleUser, Content: query},
		}

		maxTurns := 5
		if mode == "planning" {
			maxTurns = 10
		}

		for turn := 0; turn < maxTurns; turn++ {
			resp, err := provider.Complete(ctx, llm.CompletionRequest{
				Messages:      messages,
				Model:         model,
				FallbackModel: e.LLM.GetFallbackModel(provider.Name()),
				MaxTokens:     4000,
				Temperature:   0.1,
				Tools:         toolDefs,
			})
			if err != nil {
				ch <- QueryChunk{Token: fmt.Sprintf("LLM error: %v", err), IsFinal: true}
				return
			}

			// Add assistant message (including tool calls) to history
			messages = append(messages, llm.Message{
				Role:      llm.RoleAssistant,
				Content:   resp.Content,
				ToolCalls: resp.ToolCalls,
			})

			if len(resp.ToolCalls) == 0 {
				// No more tools to call, return final answer
				ch <- QueryChunk{Token: resp.Content, Sources: chunks, IsFinal: true}
				return
			}

			// Execute tools and add results to history
			for _, tc := range resp.ToolCalls {
				ch <- QueryChunk{Token: fmt.Sprintf("\n[Executing %s...]", tc.Name), IsFinal: false}

				result, err := e.Tools.ExecuteStringArgs(ctx, tc.Name, tc.Args)
				output := ""
				if err != nil {
					output = fmt.Sprintf("Error: %v", err)
				} else {
					output = result.Output
				}

				messages = append(messages, llm.Message{
					Role:       llm.RoleTool,
					Content:    output,
					ToolCallID: tc.ID,
				})
			}
		}

		ch <- QueryChunk{Token: "Max tool turns reached.", IsFinal: true}
	}()

	return ch, nil
}

// ProcessQuery is the synchronous version of StreamQuery.
func (e *Engine) ProcessQuery(query string, workspaceID string) (string, error) {
	ctx := context.Background()
	ch, err := e.StreamQuery(ctx, query, workspaceID, "", "", "")
	if err != nil {
		return "", err
	}
	var result string
	for chunk := range ch {
		result += chunk.Token
	}
	return result, nil
}

// ReadFile reads a file via the read_file tool.
func (e *Engine) ReadFile(ctx context.Context, path string) (string, error) {
	res, err := e.ExecuteTool(ctx, "read_file", map[string]any{"path": path})
	if err != nil {
		return "", err
	}
	if res.IsError {
		return "", fmt.Errorf("%s", res.Output)
	}
	return res.Output, nil
}

// WriteFile writes to a file via the write_file tool.
func (e *Engine) WriteFile(ctx context.Context, path, content string) error {
	res, err := e.ExecuteTool(ctx, "write_file", map[string]any{"path": path, "content": content})
	if err != nil {
		return err
	}
	if res.IsError {
		return fmt.Errorf("%s", res.Output)
	}
	return nil
}

// RunCommand executes a shell command via the terminal_run tool.
func (e *Engine) RunCommand(ctx context.Context, cwd, command string) (string, error) {
	res, err := e.ExecuteTool(ctx, "terminal_run", map[string]any{"command": command})
	if err != nil {
		return "", err
	}
	if res.IsError {
		return res.Output, fmt.Errorf("command failed")
	}
	return res.Output, nil
}

// GrepSearch performs a regex search via the grep_search tool.
func (e *Engine) GrepSearch(ctx context.Context, root, pattern string) (string, error) {
	res, err := e.ExecuteTool(ctx, "grep_search", map[string]any{"pattern": pattern, "path": root})
	if err != nil {
		return "", err
	}
	return res.Output, nil
}

// Edit applies changes to a file via the edit tool.
func (e *Engine) Edit(ctx context.Context, path, instruction, content string) (string, error) {
	res, err := e.ExecuteTool(ctx, "edit", map[string]any{
		"path":        path,
		"instruction": instruction,
		"content":     content,
	})
	if err != nil {
		return "", err
	}
	return res.Output, nil
}

// WebSearch performs a web search via the google_search tool.
func (e *Engine) WebSearch(ctx context.Context, query string) (string, error) {
	res, err := e.ExecuteTool(ctx, "google_search", map[string]any{"query": query})
	if err != nil {
		return "", err
	}
	return res.Output, nil
}

// WebFetch fetches content via the web_fetch tool.
func (e *Engine) WebFetch(ctx context.Context, url string) (string, error) {
	res, err := e.ExecuteTool(ctx, "web_fetch", map[string]any{"url": url})
	if err != nil {
		return "", err
	}
	return res.Output, nil
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

// ListModels returns all available models for a given provider or the default one.
func (e *Engine) ListModels(ctx context.Context, providerName string) ([]string, error) {
	if providerName == "" {
		p := e.LLM.GetDefault()
		if p == nil {
			return nil, fmt.Errorf("no default provider configured")
		}
		providerName = p.Name()
	}
	return e.LLM.ListModels(ctx, providerName)
}

// CompleteCode suggests code based on prefix and suffix.
func (e *Engine) CompleteCode(ctx context.Context, path, prefix, suffix, language string) (string, error) {
	provider := e.LLM.GetDefault()
	if provider == nil || !provider.IsConfigured() {
		return "", fmt.Errorf("no LLM provider configured")
	}

	// Simple FIM (Fill-In-the-Middle) prompt
	prompt := fmt.Sprintf("Complete the code for the file: %s in language: %s.\n\nPREFIX:\n%s\n\nSUFFIX:\n%s\n\nReturn ONLY the code to be inserted between PREFIX and SUFFIX. Do not include markdown blocks or explanations.", path, language, prefix, suffix)

	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: "You are an expert software engineer specialized in code completion. Response must contain only the code to be inserted, without any preamble or backticks."},
		{Role: llm.RoleUser, Content: prompt},
	}

	resp, err := provider.Complete(ctx, llm.CompletionRequest{
		Messages:    messages,
		MaxTokens:   128, // Small for speed
		Temperature: 0.0, // Stable
	})
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}
