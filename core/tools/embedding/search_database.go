package embedding

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/Kaffyn/Vectora/core/tools"
)

// SearchDatabaseTool performs semantic search in ChromemDB with metadata filtering.
// Phase 4G: Semantic search in vector database.
type SearchDatabaseTool struct {
	Router   *llm.Router
	VecStore db.VectorStore
	Logger   *slog.Logger
}

// NewSearchDatabaseTool creates a new semantic search tool instance.
func NewSearchDatabaseTool(router *llm.Router, vecStore db.VectorStore, logger *slog.Logger) *SearchDatabaseTool {
	return &SearchDatabaseTool{
		Router:   router,
		VecStore: vecStore,
		Logger:   logger,
	}
}

// Name returns the tool name for registration.
func (t *SearchDatabaseTool) Name() string {
	return "search_database"
}

// Description returns human-readable tool description.
func (t *SearchDatabaseTool) Description() string {
	return "Perform semantic search across indexed codebase or documents. Finds code/text similar to your query using vector embeddings. Use this to discover patterns, find implementations, or locate related code without knowing exact names. Returns relevant code snippets with similarity scores."
}

// Schema returns JSON-Schema for tool parameters.
func (t *SearchDatabaseTool) Schema() json.RawMessage {
	return []byte(`{
  "type": "object",
  "properties": {
    "query": {
      "type": "string",
      "description": "Text query to search for (will be embedded using LLM provider)"
    },
    "workspace_id": {
      "type": "string",
      "description": "Workspace to search in (default: 'default')"
    },
    "top_k": {
      "type": "integer",
      "description": "Number of results to return (default: 10, max: 100)"
    },
    "metadata_filter": {
      "type": "object",
      "description": "Optional metadata filters for exact matching (e.g., {\"source\": \"file.txt\", \"provider\": \"claude\"})",
      "additionalProperties": {"type": "string"}
    }
  },
  "required": ["query"]
}`)
}

// Execute performs semantic search in the vector database.
func (t *SearchDatabaseTool) Execute(ctx context.Context, args json.RawMessage) (*tools.ToolResult, error) {
	var input struct {
		Query          string            `json:"query"`
		WorkspaceID    string            `json:"workspace_id,omitempty"`
		TopK           int               `json:"top_k,omitempty"`
		MetadataFilter map[string]string `json:"metadata_filter,omitempty"`
	}

	if err := json.Unmarshal(args, &input); err != nil {
		return &tools.ToolResult{
			Output:  "Invalid input: " + err.Error(),
			IsError: true,
		}, nil
	}

	// Validate input
	if input.Query == "" {
		return &tools.ToolResult{
			Output:  "Query is required",
			IsError: true,
		}, nil
	}

	// Default workspace
	if input.WorkspaceID == "" {
		input.WorkspaceID = "default"
	}

	// Default and validate topK
	if input.TopK == 0 {
		input.TopK = 10
	}
	if input.TopK > 100 {
		input.TopK = 100
	}

	t.Logger.Debug("Searching database",
		slog.String("query", input.Query),
		slog.String("workspace_id", input.WorkspaceID),
		slog.Int("top_k", input.TopK))

	// Get default LLM provider to embed the query
	provider := t.Router.GetDefault()
	if provider == nil {
		return &tools.ToolResult{
			Output:  "No LLM provider configured",
			IsError: true,
		}, nil
	}

	// Embed the query using the same provider
	queryEmbedding, err := provider.Embed(ctx, input.Query, "")
	if err != nil {
		return &tools.ToolResult{
			Output:  fmt.Sprintf("Query embedding failed: %v", err),
			IsError: true,
		}, nil
	}

	// Perform semantic search in ChromemDB
	collectionID := fmt.Sprintf("ws_%s", input.WorkspaceID)
	results, err := t.VecStore.Query(ctx, collectionID, queryEmbedding, input.TopK)
	if err != nil {
		return &tools.ToolResult{
			Output:  fmt.Sprintf("Database query failed: %v", err),
			IsError: true,
		}, nil
	}

	// Filter results by metadata if provided
	var filteredResults []db.ScoredChunk
	if len(input.MetadataFilter) > 0 {
		for _, result := range results {
			if matchesAllFilters(result.Chunk.Metadata, input.MetadataFilter) {
				filteredResults = append(filteredResults, result)
			}
		}
	} else {
		filteredResults = results
	}

	t.Logger.Info("Search complete",
		slog.Int("total_results", len(results)),
		slog.Int("filtered_results", len(filteredResults)))

	// Format results
	output := map[string]interface{}{
		"query":         input.Query,
		"workspace":     input.WorkspaceID,
		"results_count": len(filteredResults),
		"results":       make([]map[string]interface{}, 0),
	}

	resultsList := output["results"].([]map[string]interface{})
	for _, result := range filteredResults {
		resultsList = append(resultsList, map[string]interface{}{
			"chunk_id":      result.ID,
			"content":       result.Content,
			"similarity":    fmt.Sprintf("%.4f", result.Score),
			"metadata":      result.Metadata,
			"embedding_dim": len(result.Vector),
		})
	}
	output["results"] = resultsList

	result, _ := json.Marshal(output)
	return &tools.ToolResult{
		Output:  string(result),
		IsError: false,
	}, nil
}

// matchesAllFilters checks if metadata contains all filter key-value pairs.
func matchesAllFilters(metadata map[string]string, filters map[string]string) bool {
	for key, filterValue := range filters {
		if metaValue, exists := metadata[key]; !exists || metaValue != filterValue {
			return false
		}
	}
	return true
}
