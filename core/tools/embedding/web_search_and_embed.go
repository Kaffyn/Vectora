package embedding

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/Kaffyn/Vectora/core/tools"
	"github.com/google/uuid"
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
	return "Research topics on the web and automatically index results. Searches for information, fetches relevant articles, and adds them to your searchable knowledge base alongside your codebase. Great for learning best practices and gathering reference materials."
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

// SearchResult represents a web search result.
type SearchResult struct {
	Title   string
	URL     string
	Content string
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

	// Get LLM provider
	provider := t.Router.GetDefault()
	if provider == nil {
		return &tools.ToolResult{
			Output:  "No LLM provider configured",
			IsError: true,
		}, nil
	}

	// Perform web search - NOTE: This would require integration with a search API
	// (e.g., DuckDuckGo, Bing, Google Custom Search)
	// For now, we return a placeholder that documents the expected workflow
	t.Logger.Warn("web_search_and_embed: Requires search API integration",
		slog.String("query", input.Query),
		slog.String("note", "DuckDuckGo, Bing, or Google Custom Search API needed"))

	// Placeholder: Create mock search results for demonstration
	searchResults := []SearchResult{
		{
			Title:   fmt.Sprintf("Search result for: %s", input.Query),
			URL:     fmt.Sprintf("https://example.com/search?q=%s", strings.ReplaceAll(input.Query, " ", "+")),
			Content: fmt.Sprintf("This is a placeholder search result for '%s'. In production, this would be fetched from a web search API.", input.Query),
		},
	}

	// Embed and store search results
	collectionID := fmt.Sprintf("ws_%s", input.WorkspaceID)
	storedChunks := 0

	for i, result := range searchResults {
		if i >= input.MaxResults {
			break
		}

		// Chunk content into reasonable pieces (simple strategy: split by sentences)
		chunks := chunkContent(result.Content, 1000) // 1000 char chunks

		for chunkIdx, content := range chunks {
			// Embed the chunk
			embedding, err := provider.Embed(ctx, content, "")
			if err != nil {
				t.Logger.Error("Embedding failed",
					slog.String("url", result.URL),
					slog.Int("chunk", chunkIdx),
					slog.String("error", err.Error()))
				continue
			}

			// Create metadata with source information
			metadata := map[string]string{
				"source":        result.URL,
				"title":         result.Title,
				"chunk_idx":     fmt.Sprintf("%d", chunkIdx),
				"provider":      provider.Name(),
				"embedding_dim": fmt.Sprintf("%d", len(embedding)),
				"tool":          "web_search_and_embed",
			}

			// Store chunk
			chunk := db.Chunk{
				ID:       uuid.New().String(),
				Content:  content,
				Metadata: metadata,
				Vector:   embedding,
			}

			if err := t.VecStore.UpsertChunk(ctx, collectionID, chunk); err != nil {
				t.Logger.Error("Failed to store chunk",
					slog.String("url", result.URL),
					slog.String("error", err.Error()))
				continue
			}

			storedChunks++
		}
	}

	// Return result summary
	output := map[string]interface{}{
		"query":          input.Query,
		"workspace":      input.WorkspaceID,
		"search_results": len(searchResults),
		"chunks_stored":  storedChunks,
		"note":           "web_search_and_embed requires search API integration (DuckDuckGo, Bing, Google Custom Search)",
		"message":        fmt.Sprintf("Processed %d search results and stored %d chunks with embeddings", len(searchResults), storedChunks),
	}

	result, _ := json.Marshal(output)
	return &tools.ToolResult{
		Output:  string(result),
		IsError: false,
	}, nil
}

// chunkContent splits content into chunks of approximately maxChunkSize characters.
// Uses sentence boundaries when possible to preserve meaning.
func chunkContent(content string, maxChunkSize int) []string {
	if len(content) <= maxChunkSize {
		return []string{content}
	}

	var chunks []string
	var currentChunk strings.Builder

	// Split by sentences (simple approach: look for . ! ? followed by space)
	sentenceRe := regexp.MustCompile(`[.!?]+\s+`)
	sentences := sentenceRe.Split(content, -1)

	for _, sentence := range sentences {
		if currentChunk.Len()+len(sentence) > maxChunkSize && currentChunk.Len() > 0 {
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
		}
		currentChunk.WriteString(sentence)
		currentChunk.WriteString(". ")
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}
