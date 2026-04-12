package mcp

import (
	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/Kaffyn/Vectora/core/tools"
	"github.com/Kaffyn/Vectora/core/tools/embedding"
)

// RegisterEmbeddingTools registers all embedding tools for MCP exposure.
// These tools are unique to Vectora and not duplicated in other agents.
// They leverage Vectora's RAG capabilities (ChromemDB + vector search + LLM).
// The registry parameter is the Engine's tool registry that needs to be updated
// so tools are discoverable via tools/list and executable via tools/call.
func RegisterEmbeddingTools(vectoraMCPServer *VectoraMCPServer, router *llm.Router, toolsRegistry interface{}) {
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

	// Phase 4H: Planning Tools
	// Plan mode tool - structured planning with context awareness
	planTool := embedding.NewPlanModeTool(
		router,
		vectoraMCPServer.vecStore,
		vectoraMCPServer.logger,
	)

	// Refactor with context tool - code refactoring with pattern matching
	refactorTool := embedding.NewRefactorWithContextTool(
		router,
		vectoraMCPServer.vecStore,
		vectoraMCPServer.logger,
	)

	// Phase 4I: Analysis Tools
	// Analyze code patterns tool - pattern detection via embeddings
	patternsAnalysisTool := embedding.NewAnalyzeCodePatternsTool(
		router,
		vectoraMCPServer.vecStore,
		vectoraMCPServer.logger,
	)

	// Knowledge graph analysis tool - entity relationship extraction
	knowledgeGraphTool := embedding.NewKnowledgeGraphAnalysisTool(
		router,
		vectoraMCPServer.vecStore,
		vectoraMCPServer.logger,
	)

	// Documentation coverage analysis tool - doc quality metrics
	docCoverageTool := embedding.NewDocCoverageAnalysisTool(
		router,
		vectoraMCPServer.vecStore,
		vectoraMCPServer.logger,
	)

	// Phase 4J: Quality Tools
	// Test generation tool - generate test cases from code
	testGenTool := embedding.NewTestGenerationTool(
		router,
		vectoraMCPServer.vecStore,
		vectoraMCPServer.logger,
	)

	// Bug pattern detection tool - identify potential bugs
	bugDetectionTool := embedding.NewBugPatternDetectionTool(
		router,
		vectoraMCPServer.vecStore,
		vectoraMCPServer.logger,
	)

	// Store tools in MCP server for exposure via tools/list and tools/call
	// These will be returned by listTools() and invoked by callTool()
	vectoraMCPServer.embeddingTools = map[string]interface{}{
		"embed":                    embedTool,
		"search_database":          searchTool,
		"web_search_and_embed":     webSearchTool,
		"web_fetch_and_embed":      webFetchTool,
		"plan_mode":                planTool,
		"refactor_with_context":    refactorTool,
		"analyze_code_patterns":    patternsAnalysisTool,
		"knowledge_graph_analysis": knowledgeGraphTool,
		"doc_coverage_analysis":    docCoverageTool,
		"test_generation":          testGenTool,
		"bug_pattern_detection":    bugDetectionTool,
	}

	// Also register tools with the Engine's tool registry so they're discoverable
	// via tools/list and executable via tools/call (Option A solution)
	if registry, ok := toolsRegistry.(*tools.Registry); ok {
		registry.Register(embedTool)
		registry.Register(searchTool)
		registry.Register(webSearchTool)
		registry.Register(webFetchTool)
		registry.Register(planTool)
		registry.Register(refactorTool)
		registry.Register(patternsAnalysisTool)
		registry.Register(knowledgeGraphTool)
		registry.Register(docCoverageTool)
		registry.Register(testGenTool)
		registry.Register(bugDetectionTool)
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
