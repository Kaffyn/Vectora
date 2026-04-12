package acp

import (
	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/Kaffyn/Vectora/core/tools/embedding"
)

// RegisterEmbeddingTools registers all embedding tools for ACP exposure.
// Phase 4G: When Vectora operates as an ACP agent (standalone mode),
// it can call embedding tools in addition to generic file/shell tools.
func (a *VectoraAgent) RegisterEmbeddingTools(router *llm.Router) {
	// Phase 4G: Core Embedding Tools
	// These tools are available when Vectora is used directly (ACP mode)

	// Embed tool - convert text to embeddings using Vectora's LLM
	embedTool := embedding.NewEmbedTool(
		router,
		a.vecStore,
		a.logger,
	)

	// Search database tool - semantic search in ChromemDB + BBoltStore
	searchTool := embedding.NewSearchDatabaseTool(
		router,
		a.vecStore,
		a.logger,
	)

	// Web search and embed tool - search web and vectorize end-to-end
	webSearchTool := embedding.NewWebSearchAndEmbedTool(
		router,
		a.vecStore,
		a.logger,
	)

	// Web fetch and embed tool - crawl internal links and vectorize
	webFetchTool := embedding.NewWebFetchAndEmbedTool(
		router,
		a.vecStore,
		a.logger,
	)

	// Store tools for ACP agent use
	a.embeddingTools = map[string]interface{}{
		"embed":                embedTool,
		"search_database":      searchTool,
		"web_search_and_embed": webSearchTool,
		"web_fetch_and_embed":  webFetchTool,
	}
}

// GetEmbeddingTool retrieves a registered embedding tool by name.
func (a *VectoraAgent) GetEmbeddingTool(name string) (interface{}, bool) {
	if a.embeddingTools == nil {
		return nil, false
	}
	tool, ok := a.embeddingTools[name]
	return tool, ok
}

// ListEmbeddingTools returns all registered embedding tools for ACP tooling.
func (a *VectoraAgent) ListEmbeddingTools() map[string]interface{} {
	if a.embeddingTools == nil {
		return make(map[string]interface{})
	}
	return a.embeddingTools
}
