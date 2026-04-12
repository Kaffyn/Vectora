package embedding

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/Kaffyn/Vectora/core/tools"
)

// WebFetchAndEmbedTool crawls URLs and vectorizes content with robots.txt compliance.
// Phase 4G: URL crawl + content extraction + chunking + embedding + storage.
type WebFetchAndEmbedTool struct {
	Router   *llm.Router
	VecStore db.VectorStore
	Logger   *slog.Logger
}

// NewWebFetchAndEmbedTool creates a new web fetch and embed tool instance.
func NewWebFetchAndEmbedTool(router *llm.Router, vecStore db.VectorStore, logger *slog.Logger) *WebFetchAndEmbedTool {
	return &WebFetchAndEmbedTool{
		Router:   router,
		VecStore: vecStore,
		Logger:   logger,
	}
}

// Name returns the tool name for registration.
func (t *WebFetchAndEmbedTool) Name() string {
	return "web_fetch_and_embed"
}

// Description returns human-readable tool description.
func (t *WebFetchAndEmbedTool) Description() string {
	return "Fetch URLs and crawl internal links with robots.txt compliance, vectorizing all content"
}

// Schema returns JSON-Schema for tool parameters.
func (t *WebFetchAndEmbedTool) Schema() json.RawMessage {
	return []byte(`{
  "type": "object",
  "properties": {
    "url": {
      "type": "string",
      "description": "Starting URL to crawl"
    },
    "max_depth": {
      "type": "integer",
      "description": "Maximum crawl depth (default: 2, max: 5)"
    },
    "max_pages": {
      "type": "integer",
      "description": "Maximum number of pages to crawl (default: 50, max: 500)"
    },
    "workspace_id": {
      "type": "string",
      "description": "Workspace to store embeddings (default: 'default')"
    },
    "css_selector": {
      "type": "string",
      "description": "CSS selector for content extraction (default: 'body')"
    }
  },
  "required": ["url"]
}`)
}

// Execute fetches and embeds web content.
func (t *WebFetchAndEmbedTool) Execute(ctx context.Context, args json.RawMessage) (*tools.ToolResult, error) {
	var input struct {
		URL         string `json:"url"`
		MaxDepth    int    `json:"max_depth,omitempty"`
		MaxPages    int    `json:"max_pages,omitempty"`
		WorkspaceID string `json:"workspace_id,omitempty"`
		CSSSelector string `json:"css_selector,omitempty"`
	}

	if err := json.Unmarshal(args, &input); err != nil {
		return &tools.ToolResult{
			Output:  "Invalid input: " + err.Error(),
			IsError: true,
		}, nil
	}

	// Validate input
	if input.URL == "" {
		return &tools.ToolResult{
			Output:  "URL is required",
			IsError: true,
		}, nil
	}

	// Default workspace
	if input.WorkspaceID == "" {
		input.WorkspaceID = "default"
	}

	// Default and validate crawl parameters
	if input.MaxDepth == 0 {
		input.MaxDepth = 2
	}
	if input.MaxDepth > 5 {
		input.MaxDepth = 5
	}

	if input.MaxPages == 0 {
		input.MaxPages = 50
	}
	if input.MaxPages > 500 {
		input.MaxPages = 500
	}

	if input.CSSSelector == "" {
		input.CSSSelector = "body"
	}

	t.Logger.Debug("Web fetch and embed",
		slog.String("url", input.URL),
		slog.Int("max_depth", input.MaxDepth),
		slog.Int("max_pages", input.MaxPages))

	// TODO: Implement web fetch and embed
	// 1. Fetch initial URL
	// 2. Check robots.txt for crawl rules
	// 3. Extract content using CSS selector
	// 4. Identify internal links
	// 5. Crawl recursively up to maxDepth
	// 6. Chunk content appropriately
	// 7. Embed each chunk using LLM provider
	// 8. Store in ChromemDB with URL + depth metadata
	// 9. Return summary of crawled pages and stored chunks

	return &tools.ToolResult{
		Output:  "web_fetch_and_embed not yet implemented",
		IsError: true,
	}, nil
}
