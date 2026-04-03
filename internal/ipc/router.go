package ipc

import (
	"context"
	"encoding/json"

	"github.com/Kaffyn/vectora/internal/core"
	"github.com/Kaffyn/vectora/internal/db"
	"github.com/Kaffyn/vectora/internal/llm"
	"github.com/Kaffyn/vectora/internal/tools"

	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	// [6] i18n.get - Lê o CSV de traduções e retorna o objeto traduzido
	server.Register("i18n.get", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		var req struct {
			Locale string `json:"locale"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}

		// Localiza o arquivo CSV no diretório do Frontend
		csvPath := filepath.Join("internal", "app", "locales", req.Locale+".csv")
		content, err := os.ReadFile(csvPath)
		if err != nil {
			return nil, &IPCError{Code: "i18n_not_found", Message: fmt.Sprintf("Locale file %s not found", req.Locale)}
		}

		// Parser simplificado de CSV para JSON (Translations)
		lines := strings.Split(string(content), "\n")
		if len(lines) < 1 {
			return map[string]any{}, nil
		}
		headers := strings.Split(lines[0], ",")
		for i := range headers {
			headers[i] = strings.TrimSpace(headers[i])
		}

		translations := make(map[string]map[string]string)
		for i := 1; i < len(lines); i++ {
			cols := strings.Split(lines[i], ",")
			if len(cols) < 2 {
				continue
			}
			key := strings.TrimSpace(cols[0])
			if key == "" {
				continue
			}
			translations[key] = make(map[string]string)
			for j := 1; j < len(cols); j++ {
				if j < len(headers) {
					translations[key][headers[j]] = strings.TrimSpace(cols[j])
				}
			}
		}

		return translations, nil
	})

	// [7] app.health
	server.Register("app.health", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		return map[string]any{"status": "ok", "version": "1.0.0"}, nil
	})

	// [8] chat.create
	server.Register("chat.create", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		var req struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}
		conv, err := msgService.CreateConversation(ctx, req.ID, req.Title)
		if err != nil {
			return nil, &IPCError{Code: "db_error", Message: err.Error()}
		}
		return conv, nil
	})

	// [9] chat.delete
	server.Register("chat.delete", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		var req struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}
		if err := msgService.DeleteConversation(ctx, req.ID); err != nil {
			return nil, &IPCError{Code: "db_error", Message: err.Error()}
		}
		return map[string]bool{"success": true}, nil
	})

	// [10] chat.rename
	server.Register("chat.rename", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		var req struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}
		if err := msgService.RenameConversation(ctx, req.ID, req.Title); err != nil {
			return nil, &IPCError{Code: "db_error", Message: err.Error()}
		}
		return map[string]bool{"success": true}, nil
	})

	// [11] message.add
	server.Register("message.add", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		var req struct {
			ConversationID string   `json:"conversationId"`
			Role           llm.Role `json:"role"`
			Content        string   `json:"content"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}
		err := msgService.AddMessage(ctx, req.ConversationID, req.Role, req.Content)
		if err != nil {
			return nil, &IPCError{Code: "db_error", Message: err.Error()}
		}
		return map[string]bool{"success": true}, nil
	})

	// [12] provider.set
	server.Register("provider.set", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		// Apenas persiste o payload no bucket 'settings'
		if err := kvStore.Set(ctx, "settings", "provider", payload); err != nil {
			return nil, &IPCError{Code: "db_error", Message: err.Error()}
		}
		return map[string]bool{"success": true}, nil
	})
}
