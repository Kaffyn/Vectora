package ipc

import (
	"context"
	"encoding/json"

	"github.com/Kaffyn/vectora/internal/core"
	"github.com/Kaffyn/vectora/internal/db"
	"github.com/Kaffyn/vectora/internal/llm"
	"github.com/Kaffyn/vectora/internal/tools"
)

// ProviderFetcher is an injection to pull the current provider from the systray dynamically.
type ProviderFetcher func() llm.Provider

// RegisterRoutes couples Backend instances with mapped RPC Endpoints.
func RegisterRoutes(
	server *Server,
	kvStore db.KVStore,
	vecStore db.VectorStore,
	getProvider ProviderFetcher,
	msgService *llm.MessageService,
	memService *db.MemoryService,
	searchService *tools.SearchService,
) {

	// [1] Main RAG Query Route (Workspace)
	server.Register("workspace.query", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		var req core.QueryRequest
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}

		provider := getProvider()
		if provider == nil {
			return nil, ErrProviderNotConfig
		}

		p := core.NewPipeline(provider, vecStore, kvStore)
		
		res, err := p.Query(ctx, req)
		if err != nil {
			return nil, &IPCError{Code: "pipeline_execution_failed", Message: err.Error()}
		}

		return res, nil
	})

	// [2] History Retrieval Route (Sessions)
	server.Register("chat.history", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		var req struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}

		conv, err := msgService.GetConversation(ctx, req.ID)
		if err != nil {
			return nil, &IPCError{Code: "chat_not_found", Message: err.Error()}
		}

		return conv, nil
	})

	// [2.1] Listar Chats
	server.Register("chat.list", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		list, err := msgService.ListConversations(ctx)
		if err != nil {
			return nil, &IPCError{Code: "db_error", Message: err.Error()}
		}
		return list, nil
	})

	// [3] Active LLM Provider Status Route
	server.Register("provider.get", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		if getProvider() == nil {
			return map[string]any{"configured": false}, nil
		}
		return map[string]any{"configured": true}, nil
	})

	// [4] Semantic Search in Ad-hoc Memory
	server.Register("memory.search", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		var req struct {
			Query string `json:"query"`
			TopK  int    `json:"top_k"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}

		results, err := memService.SearchInsight(ctx, req.Query, req.TopK)
		if err != nil {
			return nil, &IPCError{Code: "memory_error", Message: err.Error()}
		}
		return results, nil
	})

	// [5] Busca Web Direta
	server.Register("search.web", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		var req struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}

		res, err := searchService.WebSearch(ctx, req.Query)
		if err != nil {
			return nil, &IPCError{Code: "search_error", Message: err.Error()}
		}
		return res, nil
	})
}
