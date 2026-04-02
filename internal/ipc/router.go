package ipc

import (
	"context"
	"encoding/json"

	"github.com/Kaffyn/vectora/internal/core"
	"github.com/Kaffyn/vectora/internal/db"
	"github.com/Kaffyn/vectora/internal/llm"
)

// ProviderFetcher é uma injeção para podermos puxar o provider atual do systray dinamicamente
type ProviderFetcher func() llm.Provider

// RegisterRoutes acopla as instâncias do Backend com os Endpoints RPC mapeados.
// Isso deixa todo o core pronto para as chamadas do Wails (Web) e Bubbletea (CLI).
func RegisterRoutes(server *Server, kvStore db.KVStore, vecStore db.VectorStore, getProvider ProviderFetcher) {

	// [1] Rota de Query Principal RAG (Workspace)
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
		
		// Dispara RAG Pipeline puro.
		res, err := p.Query(ctx, req)
		if err != nil {
			return nil, &IPCError{
				Code:    "pipeline_execution_failed",
				Message: err.Error(),
			}
		}

		return res, nil
	})

	// [2] Rota de Retorno de Histórico Lineário (Sessões)
	server.Register("session.history", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		// Busca na persistência do bbbolt injetada
		histBytes, err := kvStore.Get(ctx, "sessions", "current_chat")
		if err != nil || histBytes == nil {
			// Retorna array vazio em caso nulo
			return []map[string]any{}, nil
		}

		var history []any
		if err := json.Unmarshal(histBytes, &history); err != nil {
			return nil, &IPCError{Code: "internal_error", Message: "Histórico corrompido"}
		}

		return history, nil
	})

	// [3] Rota Status do Provedor de LLM Ativo
	server.Register("provider.get", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		if getProvider() == nil {
			return map[string]any{"configured": false}, nil
		}

		return map[string]any{
			"configured": true,
		}, nil
	})

	// [4] Roteamento de Tool Revert (Undo Snapshot)
	server.Register("tool.undo", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		// Chamada pendente do Bridge Git.
		// Todo Frontend invoca isso debaixo da bolha do chat.
		return map[string]any{"restored": true}, nil
	})
}
