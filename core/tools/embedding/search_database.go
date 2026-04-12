package embedding

import (
	"context"
	"encoding/json"
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
	return "Perform semantic search in ChromemDB vector database with metadata filtering support"
}

// Schema returns JSON-Schema for tool parameters.
func (t *SearchDatabaseTool) Schema() json.RawMessage {
	return []byte(`{
  "type": "object",
  "properties": {
    "query": {
      "type": "string",
      "description": "Text query to search for (will be embedded)"
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
      "description": "Optional metadata filters (e.g., {\"source\": \"file.txt\"})",
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

	t.Logger.Debug("Searching database", slog.String("query", input.Query), slog.Int("top_k", input.TopK))

	// TODO: Implement semantic search
	// 1. Embed query using LLM provider
	// 2. Query ChromemDB with vector + topK
	// 3. Apply metadata filters if provided
	// 4. Return results with similarity scores

	return &tools.ToolResult{
		Output:  "search_database not yet implemented",
		IsError: true,
	}, nil
}
