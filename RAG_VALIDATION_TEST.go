package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/Kaffyn/Vectora/internal/core"
)

func main() {
	// Setup logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create manager
	manager := core.NewManager(logger)
	ctx := context.Background()

	fmt.Println("=== RAG PIPELINE VALIDATION TEST ===\n")

	// Test 1: Create Workspace
	fmt.Println("[TEST 1] Creating Workspace...")
	ws, err := manager.CreateWorkspace(ctx, "Documentation", "Test documentation workspace")
	if err != nil {
		logger.Error("Failed to create workspace", "error", err)
		return
	}
	fmt.Printf("Created workspace: %s (%s)\n", ws.Name, ws.ID)
	fmt.Printf("Status: %v, CreatedAt: %v\n\n", ws.Status, ws.CreatedAt)

	// Test 2: Index Chunks
	fmt.Println("[TEST 2] Indexing Chunks...")
	chunks := []*core.Chunk{
		{
			ID:         "chunk_1",
			Content:    "Vectora is a RAG framework for AI applications",
			SourceFile: "docs.md",
			ChunkIndex: 0,
			Metadata:   map[string]string{"type": "documentation"},
		},
		{
			ID:         "chunk_2",
			Content:    "RAG combines retrieval and generation for better answers",
			SourceFile: "docs.md",
			ChunkIndex: 1,
			Metadata:   map[string]string{"type": "documentation"},
		},
		{
			ID:         "chunk_3",
			Content:    "Vectora supports Gemini and Qwen LLM providers",
			SourceFile: "config.md",
			ChunkIndex: 0,
			Metadata:   map[string]string{"type": "configuration"},
		},
	}

	for _, chunk := range chunks {
		err := manager.IndexChunk(ctx, ws.ID, chunk)
		if err != nil {
			logger.Error("Failed to index chunk", "id", chunk.ID, "error", err)
			return
		}
		fmt.Printf("Indexed chunk: %s\n", chunk.ID)
	}
	fmt.Println()

	// Test 3: List Workspaces
	fmt.Println("[TEST 3] Listing Workspaces...")
	workspaces, err := manager.ListWorkspaces(ctx)
	if err != nil {
		logger.Error("Failed to list workspaces", "error", err)
		return
	}
	fmt.Printf("Total workspaces: %d\n\n", len(workspaces))

	// Test 4: Query Workspace
	fmt.Println("[TEST 4] Querying Workspace...")
	result, err := manager.Query(ctx, ws.ID, "What is RAG and Vectora?")
	if err != nil {
		logger.Error("Failed to query", "error", err)
		return
	}
	fmt.Printf("Query ID: %s\n", result.ID)
	fmt.Printf("Model: %s\n", result.Model)
	fmt.Printf("Execution Time: %dms\n", result.ExecutionTime)
	fmt.Printf("Sources Found: %d\n", len(result.Sources))
	fmt.Printf("Answer: %s\n\n", result.Answer)

	// Test 5: Execute Tool
	fmt.Println("[TEST 5] Executing Tool (read_file)...")
	toolArgs := []byte(`{"path": "` + os.Args[0] + `"}`)
	result_tool, err := manager.ExecuteTool(ctx, "read_file", toolArgs)
	if err != nil {
		logger.Error("Failed to execute tool", "error", err)
	} else {
		fmt.Printf("Tool Result: %v\n\n", result_tool)
	}

	fmt.Println("=== ALL TESTS COMPLETED ===")
}
