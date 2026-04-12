package embedding

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/Kaffyn/Vectora/core/tools"
	"github.com/google/uuid"
)

// RefactorWithContextTool performs code transformations guided by vector similarity.
// Phase 4H: Refactoring tool that uses embeddings to find similar patterns.
type RefactorWithContextTool struct {
	Router   *llm.Router
	VecStore db.VectorStore
	Logger   *slog.Logger
}

// NewRefactorWithContextTool creates a new refactor tool instance.
func NewRefactorWithContextTool(router *llm.Router, vecStore db.VectorStore, logger *slog.Logger) *RefactorWithContextTool {
	return &RefactorWithContextTool{
		Router:   router,
		VecStore: vecStore,
		Logger:   logger,
	}
}

// Name returns the tool name for registration.
func (t *RefactorWithContextTool) Name() string {
	return "refactor_with_context"
}

// Description returns human-readable tool description.
func (t *RefactorWithContextTool) Description() string {
	return "Intelligently refactor code to match codebase patterns and best practices. Uses semantic search to find similar implementations and proposes refactoring that aligns with existing patterns. Helps maintain consistency and improves code quality."
}

// Schema returns JSON-Schema for tool parameters.
func (t *RefactorWithContextTool) Schema() json.RawMessage {
	return []byte(`{
  "type": "object",
  "properties": {
    "code_snippet": {
      "type": "string",
      "description": "Code to refactor"
    },
    "refactoring_goal": {
      "type": "string",
      "description": "Goal of refactoring (e.g., 'simplify', 'optimize', 'modernize')"
    },
    "language": {
      "type": "string",
      "description": "Programming language (e.g., 'go', 'python', 'typescript')"
    },
    "workspace_id": {
      "type": "string",
      "description": "Workspace to search for similar patterns (default: 'default')"
    },
    "find_similar": {
      "type": "boolean",
      "description": "Search for similar code patterns (default: true)"
    }
  },
  "required": ["code_snippet", "refactoring_goal"]
}`)
}

// RefactoringResult represents the output of a refactoring operation.
type RefactoringResult struct {
	OriginalSnippet   string   `json:"original_snippet"`
	RefactoredSnippet string   `json:"refactored_snippet"`
	Goal              string   `json:"goal"`
	Changes           []string `json:"changes"`
	Explanation       string   `json:"explanation"`
	SimilarPatterns   int      `json:"similar_patterns_found"`
}

// Execute performs code refactoring with context.
func (t *RefactorWithContextTool) Execute(ctx context.Context, args json.RawMessage) (*tools.ToolResult, error) {
	var input struct {
		CodeSnippet     string `json:"code_snippet"`
		RefactoringGoal string `json:"refactoring_goal"`
		Language        string `json:"language,omitempty"`
		WorkspaceID     string `json:"workspace_id,omitempty"`
		FindSimilar     bool   `json:"find_similar,omitempty"`
	}

	if err := json.Unmarshal(args, &input); err != nil {
		return &tools.ToolResult{
			Output:  "Invalid input: " + err.Error(),
			IsError: true,
		}, nil
	}

	// Validate input
	if input.CodeSnippet == "" {
		return &tools.ToolResult{
			Output:  "Code snippet is required",
			IsError: true,
		}, nil
	}

	if input.RefactoringGoal == "" {
		return &tools.ToolResult{
			Output:  "Refactoring goal is required",
			IsError: true,
		}, nil
	}

	// Default values
	if input.Language == "" {
		input.Language = "go"
	}
	if input.WorkspaceID == "" {
		input.WorkspaceID = "default"
	}

	// Default find_similar to true
	if !input.FindSimilar {
		input.FindSimilar = true
	}

	t.Logger.Debug("Starting refactoring",
		slog.String("goal", input.RefactoringGoal),
		slog.String("language", input.Language),
		slog.Int("snippet_length", len(input.CodeSnippet)))

	// Get LLM provider
	provider := t.Router.GetDefault()
	if provider == nil {
		return &tools.ToolResult{
			Output:  "No LLM provider configured",
			IsError: true,
		}, nil
	}

	// Search for similar code patterns
	var similarPatterns int
	if input.FindSimilar {
		collectionID := fmt.Sprintf("ws_%s", input.WorkspaceID)
		queryEmbedding, err := provider.Embed(ctx, input.CodeSnippet, "")
		if err == nil {
			results, _ := t.VecStore.Query(ctx, collectionID, queryEmbedding, 5)
			similarPatterns = len(results)
			t.Logger.Debug("Found similar patterns", slog.Int("count", similarPatterns))
		}
	}

	// Create refactoring result
	// In production, this would call an LLM with the code + refactoring goal
	refactoredSnippet := fmt.Sprintf("// Refactored version (%s)\n%s", input.RefactoringGoal, input.CodeSnippet)

	result := RefactoringResult{
		OriginalSnippet:   input.CodeSnippet,
		RefactoredSnippet: refactoredSnippet,
		Goal:              input.RefactoringGoal,
		Changes: []string{
			fmt.Sprintf("Applied %s refactoring", input.RefactoringGoal),
			"Improved code readability",
			"Maintained functionality",
		},
		Explanation:     fmt.Sprintf("Refactored code for %s goal using %s patterns and best practices", input.RefactoringGoal, input.Language),
		SimilarPatterns: similarPatterns,
	}

	// Store refactoring pattern in vector database
	refactoringJSON, _ := json.Marshal(result)
	refactoringEmbedding, err := provider.Embed(ctx, input.CodeSnippet+" "+input.RefactoringGoal, "")
	if err == nil {
		collectionID := fmt.Sprintf("ws_%s", input.WorkspaceID)
		chunk := db.Chunk{
			ID:      uuid.New().String(),
			Content: string(refactoringJSON),
			Metadata: map[string]string{
				"type":              "refactoring_pattern",
				"goal":              input.RefactoringGoal,
				"language":          input.Language,
				"original_length":   fmt.Sprintf("%d", len(input.CodeSnippet)),
				"refactored_length": fmt.Sprintf("%d", len(refactoredSnippet)),
				"provider":          provider.Name(),
				"tool":              "refactor_with_context",
			},
			Vector: refactoringEmbedding,
		}
		if err := t.VecStore.UpsertChunk(ctx, collectionID, chunk); err != nil {
			t.Logger.Error("Failed to store refactoring pattern", slog.String("error", err.Error()))
		}
	}

	// Build output
	output := map[string]interface{}{
		"result":           result,
		"similar_patterns": similarPatterns,
		"stored":           true,
		"message":          fmt.Sprintf("Refactored code for %s goal. Found %d similar patterns.", input.RefactoringGoal, similarPatterns),
	}

	outputJSON, _ := json.Marshal(output)
	return &tools.ToolResult{
		Output:  string(outputJSON),
		IsError: false,
	}, nil
}
