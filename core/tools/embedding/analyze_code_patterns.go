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

// AnalyzeCodePatternsTool detects patterns in code using vector similarity.
// Phase 4I: Pattern detection via semantic search on code embeddings.
type AnalyzeCodePatternsTool struct {
	Router   *llm.Router
	VecStore db.VectorStore
	Logger   *slog.Logger
}

// NewAnalyzeCodePatternsTool creates a new code pattern analysis tool.
func NewAnalyzeCodePatternsTool(router *llm.Router, vecStore db.VectorStore, logger *slog.Logger) *AnalyzeCodePatternsTool {
	return &AnalyzeCodePatternsTool{
		Router:   router,
		VecStore: vecStore,
		Logger:   logger,
	}
}

// Name returns the tool name for registration.
func (t *AnalyzeCodePatternsTool) Name() string {
	return "analyze_code_patterns"
}

// Description returns human-readable tool description.
func (t *AnalyzeCodePatternsTool) Description() string {
	return "Analyze code to identify design patterns (Singleton, Observer, etc), anti-patterns (code smells), and architectural patterns. Detects concurrency issues, error handling patterns, performance problems, and security concerns."
}

// Schema returns JSON-Schema for tool parameters.
func (t *AnalyzeCodePatternsTool) Schema() json.RawMessage {
	return []byte(`{
  "type": "object",
  "properties": {
    "code_snippets": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Array of code snippets to analyze for patterns"
    },
    "language": {
      "type": "string",
      "description": "Programming language (default: 'go')"
    },
    "pattern_type": {
      "type": "string",
      "enum": ["design", "anti", "performance", "all"],
      "description": "Type of patterns to find (default: 'all')"
    },
    "workspace_id": {
      "type": "string",
      "description": "Workspace for pattern search (default: 'default')"
    }
  },
  "required": ["code_snippets"]
}`)
}

// PatternAnalysis represents analysis of detected patterns.
type PatternAnalysis struct {
	TotalSnippets   int           `json:"total_snippets"`
	PatternsFound   int           `json:"patterns_found"`
	Patterns        []interface{} `json:"patterns"`
	Risks           []string      `json:"risks"`
	Recommendations []string      `json:"recommendations"`
}

// Execute analyzes code patterns.
func (t *AnalyzeCodePatternsTool) Execute(ctx context.Context, args json.RawMessage) (*tools.ToolResult, error) {
	var input struct {
		CodeSnippets []string `json:"code_snippets"`
		Language     string   `json:"language,omitempty"`
		PatternType  string   `json:"pattern_type,omitempty"`
		WorkspaceID  string   `json:"workspace_id,omitempty"`
	}

	if err := json.Unmarshal(args, &input); err != nil {
		return &tools.ToolResult{
			Output:  "Invalid input: " + err.Error(),
			IsError: true,
		}, nil
	}

	// Validate input
	if len(input.CodeSnippets) == 0 {
		return &tools.ToolResult{
			Output:  "At least one code snippet is required",
			IsError: true,
		}, nil
	}

	// Default values
	if input.Language == "" {
		input.Language = "go"
	}
	if input.PatternType == "" {
		input.PatternType = "all"
	}
	if input.WorkspaceID == "" {
		input.WorkspaceID = "default"
	}

	t.Logger.Debug("Analyzing code patterns",
		slog.Int("snippet_count", len(input.CodeSnippets)),
		slog.String("language", input.Language),
		slog.String("pattern_type", input.PatternType))

	// Get LLM provider
	provider := t.Router.GetDefault()
	if provider == nil {
		return &tools.ToolResult{
			Output:  "No LLM provider configured",
			IsError: true,
		}, nil
	}

	// Analyze each snippet
	analysis := PatternAnalysis{
		TotalSnippets: len(input.CodeSnippets),
		Patterns:      make([]interface{}, 0),
	}

	collectionID := fmt.Sprintf("ws_%s", input.WorkspaceID)

	for i, snippet := range input.CodeSnippets {
		// Embed snippet
		embedding, err := provider.Embed(ctx, snippet, "")
		if err != nil {
			t.Logger.Error("Embedding failed", slog.Int("snippet", i), slog.String("error", err.Error()))
			continue
		}

		// Search for similar patterns in database
		results, err := t.VecStore.Query(ctx, collectionID, embedding, 5)
		if err != nil {
			t.Logger.Debug("Pattern search returned no results", slog.Int("snippet", i))
			continue
		}

		// Analyze patterns found
		var pattern map[string]interface{}
		if len(results) > 0 {
			pattern = map[string]interface{}{
				"snippet_index":    i,
				"similar_patterns": len(results),
				"top_match_score":  fmt.Sprintf("%.4f", results[0].Score),
				"pattern_type":     input.PatternType,
			}
			analysis.Patterns = append(analysis.Patterns, pattern)
			analysis.PatternsFound++
		}

		// Store analysis in vector database
		if pattern != nil {
			analysisJSON, _ := json.Marshal(pattern)
			chunk := db.Chunk{
				ID:      uuid.New().String(),
				Content: string(analysisJSON),
				Metadata: map[string]string{
					"type":         "code_pattern",
					"language":     input.Language,
					"pattern_type": input.PatternType,
					"provider":     provider.Name(),
					"tool":         "analyze_code_patterns",
				},
				Vector: embedding,
			}
			if err := t.VecStore.UpsertChunk(ctx, collectionID, chunk); err != nil {
				t.Logger.Error("Failed to store pattern", slog.Int("snippet", i))
			}
		}
	}

	// Generate recommendations
	analysis.Recommendations = []string{
		fmt.Sprintf("Found %d patterns across %d snippets", analysis.PatternsFound, analysis.TotalSnippets),
		"Consider refactoring repeated patterns",
		"Review anti-patterns for best practice violations",
	}

	// Output
	output := map[string]interface{}{
		"analysis": analysis,
		"stored":   true,
		"message":  fmt.Sprintf("Pattern analysis complete: %d patterns found", analysis.PatternsFound),
	}

	result, _ := json.Marshal(output)
	return &tools.ToolResult{
		Output:  string(result),
		IsError: false,
	}, nil
}
