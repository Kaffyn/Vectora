// cmd/acp is a standalone ACP server binary for testing and production use.
// It reads JSON-RPC 2.0 requests from stdin and writes responses to stdout.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Kaffyn/Vectora/core/api/acp"
	"github.com/Kaffyn/Vectora/core/api/mcp"
	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/engine"
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

	// Initialize Providers
	llmRouter := llm.NewRouter()

	// 1. Gemini (Default)
	if cfg.GeminiAPIKey != "" {
		gemini, err := llm.NewGeminiProvider(ctx, cfg.GeminiAPIKey)
		if err == nil {
			llmRouter.RegisterProvider("gemini", gemini, true)
		} else {
			fmt.Fprintf(os.Stderr, "Failed to init Gemini provider: %v\n", err)
		}
	}

	// 2. Claude
	if cfg.ClaudeAPIKey != "" {
		claude, err := llm.NewClaudeProvider(ctx, cfg.ClaudeAPIKey)
		if err == nil {
			llmRouter.RegisterProvider("claude", claude, false)
		} else {
			fmt.Fprintf(os.Stderr, "Failed to init Claude provider: %v\n", err)
		}
	}

	// 3. Voyage (Embedding Sniper)
	if cfg.VoyageAPIKey != "" {
		voyage, err := llm.NewVoyageProvider(ctx, cfg.VoyageAPIKey)
		if err == nil {
			llmRouter.RegisterProvider("voyage", voyage, false)
		} else {
			fmt.Fprintf(os.Stderr, "Failed to init Voyage provider: %v\n", err)
		}
	}

	// Initialize tools registry
	cwd, _ := os.Getwd()
	guardian := policies.NewGuardian(cwd)
	toolRegistry := tools.NewRegistry(cwd, guardian, kvStore)

	// Create centralized core engine
	eng := engine.NewEngine(vecStore, kvStore, llmRouter, toolRegistry, guardian, nil)

	// Start Dual Protocol (MCP Server over HTTP)
	mcpServer := mcp.NewServer(eng)
	go func() {
		if err := mcpServer.Start(":8080"); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		}
	}()

	fmt.Fprintf(os.Stderr, "Vectora Server started. Protocol: ACP (stdio) + MCP (HTTP :8080)\n")

	server := acp.NewServer(eng)
	if err := server.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "ACP server error: %v\n", err)
		os.Exit(1)
	}
}
