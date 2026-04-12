package mcp

import (
	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/Kaffyn/Vectora/core/tools/embedding"
)

// RegisterEmbeddingTools registers all embedding tools for MCP exposure.
// These tools are unique to Vectora and not duplicated in other agents.
// They leverage Vectora's RAG capabilities (ChromemDB + vector search + LLM).
func RegisterEmbeddingTools(vectoraMCPServer *VectoraMCPServer, router *llm.Router) {
	// Phase 4G: Core Embedding Tools
	// Embed tool - convert text to embeddings using Vectora's LLM
	embedTool := embedding.NewEmbedTool(
		router,
		vectoraMCPServer.vecStore,
		vectoraMCPServer.logger,
	)

	// Search database tool - semantic search in ChromemDB + BBoltStore
	searchTool := embedding.NewSearchDatabaseTool(
		router,
		vectoraMCPServer.vecStore,
		vectoraMCPServer.logger,
	)

	// Web search and embed tool - search web and vectorize end-to-end
	webSearchTool := embedding.NewWebSearchAndEmbedTool(
		router,
		vectoraMCPServer.vecStore,
		vectoraMCPServer.logger,
	)

	// Web fetch and embed tool - crawl internal links and vectorize
	webFetchTool := embedding.NewWebFetchAndEmbedTool(
		router,
		vectoraMCPServer.vecStore,
		vectoraMCPServer.logger,
	)

	// Store tools in MCP server for exposure via tools/list and tools/call
	// These will be returned by listTools() and invoked by callTool()
	vectoraMCPServer.embeddingTools = map[string]interface{}{
		"embed":                 embedTool,
		"search_database":       searchTool,
		"web_search_and_embed":  webSearchTool,
		"web_fetch_and_embed":   webFetchTool,
	}
}

// GetEmbeddingTool retrieves a registered embedding tool by name.
func (s *VectoraMCPServer) GetEmbeddingTool(name string) (interface{}, bool) {
	if s.embeddingTools == nil {
		return nil, false
	}
	tool, ok := s.embeddingTools[name]
	return tool, ok
}

// ListEmbeddingTools returns all registered embedding tools for MCP tools/list.
func (s *VectoraMCPServer) ListEmbeddingTools() map[string]interface{} {
	if s.embeddingTools == nil {
		return make(map[string]interface{})
	}
	return s.embeddingTools
}
