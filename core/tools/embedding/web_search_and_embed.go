package embedding

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/Kaffyn/Vectora/core/tools"
)

// WebSearchAndEmbedTool performs web search and vectorizes results end-to-end.
// Phase 4G: Web search + content fetch + embedding + storage.
type WebSearchAndEmbedTool struct {
	Router   *llm.Router
	VecStore db.VectorStore
	Logger   *slog.Logger
}

// NewWebSearchAndEmbedTool creates a new web search and embed tool instance.
func NewWebSearchAndEmbedTool(router *llm.Router, vecStore db.VectorStore, logger *slog.Logger) *WebSearchAndEmbedTool {
	return &WebSearchAndEmbedTool{
		Router:   router,
		VecStore: vecStore,
		Logger:   logger,
	}
}

// Name returns the tool name for registration.
func (t *WebSearchAndEmbedTool) Name() string {
	return "web_search_and_embed"
}

// Description returns human-readable tool description.
func (t *WebSearchAndEmbedTool) Description() string {
	return "Search the web for content and automatically vectorize results, storing in ChromemDB"
}

// Schema returns JSON-Schema for tool parameters.
func (t *WebSearchAndEmbedTool) Schema() json.RawMessage {
	return []byte(`{
  "type": "object",
  "properties": {
    "query": {
      "type": "string",
      "description": "Web search query"
    },
    "max_results": {
      "type": "integer",
      "description": "Maximum number of results to fetch and embed (default: 5, max: 20)"
    },
    "workspace_id": {
      "type": "string",
      "description": "Workspace to store embeddings (default: 'default')"
    }
  },
  "required": ["query"]
}`)
}

// Execute performs web search and embeds results.
func (t *WebSearchAndEmbedTool) Execute(ctx context.Context, args json.RawMessage) (*tools.ToolResult, error) {
	var input struct {
		Query       string `json:"query"`
		MaxResults  int    `json:"max_results,omitempty"`
		WorkspaceID string `json:"workspace_id,omitempty"`
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

	// Default and validate maxResults
	if input.MaxResults == 0 {
		input.MaxResults = 5
	}
	if input.MaxResults > 20 {
		input.MaxResults = 20
	}

	t.Logger.Debug("Web search and embed",
		slog.String("query", input.Query),
		slog.Int("max_results", input.MaxResults))

	// TODO: Implement web search and embed
	// 1. Perform web search using DuckDuckGo or similar
	// 2. Fetch content from top results
	// 3. Chunk content as needed
	// 4. Embed each chunk using LLM provider
	// 5. Store in ChromemDB with source metadata (URL, title, date)
	// 6. Return summary of stored chunks

	return &tools.ToolResult{
		Output:  "web_search_and_embed not yet implemented",
		IsError: true,
	}, nil
}
