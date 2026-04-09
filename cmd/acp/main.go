// cmd/acp is a standalone ACP server binary for testing and production use.
// It reads JSON-RPC 2.0 requests from stdin and writes responses to stdout.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Kaffyn/Vectora/core/api/acp"
	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/infra"
	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/Kaffyn/Vectora/core/policies"
	"github.com/Kaffyn/Vectora/core/tools"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Load config for API keys
	cfg := infra.LoadConfig()
	if cfg.GeminiAPIKey == "" {
		fmt.Fprintf(os.Stderr, "WARNING: GEMINI_API_KEY not set. Embeddings will fail.\n")
	}

	// Initialize stores
	kvStore, err := db.NewKVStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init KV store: %v\n", err)
		os.Exit(1)
	}
	defer kvStore.Close()

	vecStore, err := db.NewVectorStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init vector store: %v\n", err)
		os.Exit(1)
	}

	// Initialize Gemini provider for embeddings and queries
	var provider llm.Provider
	if cfg.GeminiAPIKey != "" {
		provider, err = llm.NewGeminiProvider(ctx, cfg.GeminiAPIKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to init Gemini provider: %v\n", err)
		}
	}

	// Initialize tools registry
	cwd, _ := os.Getwd()
	guardian := policies.NewGuardian(cwd)
	toolRegistry := tools.NewRegistry(cwd, guardian, kvStore)

	// Create LLM router
	llmRouter := llm.NewRouter()
	if provider != nil {
		llmRouter.RegisterProvider("gemini", provider, true)
	}

	// Create engine
	eng := &ACP{
		kv:       kvStore,
		vec:      vecStore,
		llm:      llmRouter,
		tools:    toolRegistry,
		guardian: guardian,
		provider: provider,
		cwd:      cwd,
	}

	fmt.Fprintf(os.Stderr, "ACP Server started. Reading from stdin, writing to stdout.\n")

	server := acp.NewServer(eng)
	if err := server.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "ACP server error: %v\n", err)
		os.Exit(1)
	}
}

// ACP implements acp.Engine for the ACP server.
type ACP struct {
	kv       *db.BBoltStore
	vec      *db.ChromemStore
	llm      *llm.Router
	tools    *tools.Registry
	guardian *policies.Guardian
	provider llm.Provider
	cwd      string
}

func (a *ACP) Embed(ctx context.Context, text string) ([]float32, error) {
	if a.provider == nil {
		return nil, fmt.Errorf("no LLM provider configured")
	}
	return a.provider.Embed(ctx, text)
}

func (a *ACP) Query(ctx context.Context, query string, workspaceID string) (string, error) {
	if a.provider == nil {
		return fmt.Sprintf("No LLM provider configured. Query: %s", query), nil
	}

	// RAG: embed query, search vectors, build context, complete
	vector, err := a.provider.Embed(ctx, query)
	if err != nil {
		// Fallback: just query the LLM directly
		return a.completeWithProvider(ctx, query, "")
	}

	// Search vector store
	chunks, err := a.vec.Query(ctx, "ws_"+workspaceID, vector, 5)
	if err != nil {
		chunks = []db.ScoredChunk{}
	}

	// Build context
	contextText := ""
	for _, chunk := range chunks {
		if filename, ok := chunk.Metadata["filename"]; ok {
			contextText += "File: " + filename + "\n"
		}
		contextText += chunk.Content + "\n---\n"
	}

	return a.completeWithProvider(ctx, query, contextText)
}

func (a *ACP) completeWithProvider(ctx context.Context, query string, contextText string) (string, error) {
	systemPrompt := "You are Vectora, an AI coding assistant."
	if contextText != "" {
		systemPrompt += "\n\nUse the following context as your source of truth:\n" + contextText
	}

	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: systemPrompt},
		{Role: llm.RoleUser, Content: query},
	}

	resp, err := a.provider.Complete(ctx, llm.CompletionRequest{
		Messages:    messages,
		MaxTokens:   2000,
		Temperature: 0.2,
	})
	if err != nil {
		return "", fmt.Errorf("LLM error: %w", err)
	}

	return resp.Content, nil
}

func (a *ACP) ExecuteTool(ctx context.Context, name string, args map[string]any) (acp.ToolResult, error) {
	tool, ok := a.tools.GetTool(name)
	if !ok {
		return acp.ToolResult{Output: fmt.Sprintf("Tool '%s' not found", name), IsError: true}, nil
	}

	// Convert map to JSON
	argsJSON, _ := json.Marshal(args)
	result, err := tool.Execute(ctx, argsJSON)
	if err != nil {
		return acp.ToolResult{Output: err.Error(), IsError: true}, nil
	}

	return acp.ToolResult{Output: result.Output, IsError: result.IsError}, nil
}

func (a *ACP) ReadFile(ctx context.Context, path string) (string, error) {
	tool, _ := a.tools.GetTool("read_file")
	if tool == nil {
		return "", fmt.Errorf("read_file tool not available")
	}

	argsJSON, _ := json.Marshal(map[string]string{"path": path})
	result, err := tool.Execute(ctx, argsJSON)
	if err != nil {
		return "", err
	}
	if result.IsError {
		return "", fmt.Errorf(result.Output)
	}
	return result.Output, nil
}

func (a *ACP) WriteFile(ctx context.Context, path, content string) error {
	tool, _ := a.tools.GetTool("write_file")
	if tool == nil {
		return fmt.Errorf("write_file tool not available")
	}

	argsJSON, _ := json.Marshal(map[string]string{"path": path, "content": content})
	result, err := tool.Execute(ctx, argsJSON)
	if err != nil {
		return err
	}
	if result.IsError {
		return fmt.Errorf(result.Output)
	}
	return nil
}

func (a *ACP) RunCommand(ctx context.Context, cwd, command string) (string, error) {
	tool, _ := a.tools.GetTool("run_shell_command")
	if tool == nil {
		return "", fmt.Errorf("run_shell_command tool not available")
	}

	argsJSON, _ := json.Marshal(map[string]string{"command": command})
	result, err := tool.Execute(ctx, argsJSON)
	if err != nil {
		return "", err
	}
	if result.IsError {
		return result.Output, fmt.Errorf("command failed")
	}
	return result.Output, nil
}
