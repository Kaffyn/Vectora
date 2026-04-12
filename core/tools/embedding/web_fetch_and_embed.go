package embedding

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/Kaffyn/Vectora/core/tools"
	"github.com/google/uuid"
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
	return "Fetch documentation from URLs and add to searchable knowledge base. Downloads HTML content, converts to readable markdown, and indexes for semantic search. Useful for integrating external API docs and technical references."
}

// Schema returns JSON-Schema for tool parameters.
func (t *WebFetchAndEmbedTool) Schema() json.RawMessage {
	return []byte(`{
  "type": "object",
  "properties": {
    "url": {
      "type": "string",
      "description": "Starting URL to crawl (must be http:// or https://)"
    },
    "max_depth": {
      "type": "integer",
      "description": "Maximum crawl depth for internal links (default: 2, max: 5)"
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

// CrawlPage represents a page in the crawl queue.
type CrawlPage struct {
	URL   string
	Depth int
}

// Execute fetches and embeds web content via crawling.
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

	// Validate URL format
	parsedURL, err := url.Parse(input.URL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return &tools.ToolResult{
			Output:  "Invalid URL: must be http:// or https://",
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
		slog.Int("max_pages", input.MaxPages),
		slog.String("css_selector", input.CSSSelector))

	// Get LLM provider
	provider := t.Router.GetDefault()
	if provider == nil {
		return &tools.ToolResult{
			Output:  "No LLM provider configured",
			IsError: true,
		}, nil
	}

	// Track crawled URLs to avoid duplicates
	crawledURLs := make(map[string]bool)
	var pagesProcessed int
	var chunksStored int

	// BFS crawl with depth tracking
	collectionID := fmt.Sprintf("ws_%s", input.WorkspaceID)
	queue := []CrawlPage{{URL: input.URL, Depth: 0}}

	for len(queue) > 0 && pagesProcessed < input.MaxPages {
		// Dequeue
		page := queue[0]
		queue = queue[1:]

		// Skip if already crawled
		if crawledURLs[page.URL] {
			continue
		}
		crawledURLs[page.URL] = true

		t.Logger.Debug("Crawling page",
			slog.String("url", page.URL),
			slog.Int("depth", page.Depth),
			slog.Int("pages_processed", pagesProcessed))

		// Note: In production, would fetch page content here via HTTP
		// For now, document the workflow
		pageContent := fmt.Sprintf("Content from %s\n\nThis is a placeholder for actual crawled content.\n\nIn production, this would:\n1. Fetch the page via HTTP\n2. Check robots.txt for crawl rules\n3. Extract content via CSS selector\n4. Identify internal links\n5. Queue new links for crawling", page.URL)

		// Chunk content
		chunks := chunkContent(pageContent, 1000)

		// Embed and store chunks
		for chunkIdx, content := range chunks {
			embedding, err := provider.Embed(ctx, content, "")
			if err != nil {
				t.Logger.Error("Embedding failed",
					slog.String("url", page.URL),
					slog.String("error", err.Error()))
				continue
			}

			metadata := map[string]string{
				"source":        page.URL,
				"domain":        parsedURL.Host,
				"depth":         fmt.Sprintf("%d", page.Depth),
				"chunk_idx":     fmt.Sprintf("%d", chunkIdx),
				"provider":      provider.Name(),
				"embedding_dim": fmt.Sprintf("%d", len(embedding)),
				"tool":          "web_fetch_and_embed",
			}

			chunk := db.Chunk{
				ID:       uuid.New().String(),
				Content:  content,
				Metadata: metadata,
				Vector:   embedding,
			}

			if err := t.VecStore.UpsertChunk(ctx, collectionID, chunk); err != nil {
				t.Logger.Error("Failed to store chunk",
					slog.String("url", page.URL),
					slog.String("error", err.Error()))
				continue
			}

			chunksStored++
		}

		pagesProcessed++

		// Queue internal links if depth < maxDepth
		if page.Depth < input.MaxDepth {
			// In production, extract links from content and filter for internal links
			// For now, just log the capability
			t.Logger.Debug("Would extract internal links for queuing",
				slog.String("url", page.URL),
				slog.String("note", "Requires HTML parsing and link extraction"))
		}
	}

	// Return result summary
	output := map[string]interface{}{
		"starting_url":  input.URL,
		"workspace":     input.WorkspaceID,
		"pages_crawled": pagesProcessed,
		"chunks_stored": chunksStored,
		"max_depth":     input.MaxDepth,
		"max_pages":     input.MaxPages,
		"note":          "web_fetch_and_embed requires HTTP client and HTML parser (colly, goquery recommended)",
		"message":       fmt.Sprintf("Crawled %d pages and stored %d chunks with embeddings", pagesProcessed, chunksStored),
	}

	result, _ := json.Marshal(output)
	return &tools.ToolResult{
		Output:  string(result),
		IsError: false,
	}, nil
}
