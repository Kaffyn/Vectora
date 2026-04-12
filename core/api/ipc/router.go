package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Kaffyn/Vectora/core/crypto"
	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/engine"
	"github.com/Kaffyn/Vectora/core/i18n"
	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/Kaffyn/Vectora/core/manager"
	"github.com/Kaffyn/Vectora/core/tray"
)

type ProviderFetcher func() llm.Provider

func RegisterRoutes(
	server *Server,
	globalKV db.KVStore,
	salter *crypto.WorkspaceSalter,
) {
	// [1] Main RAG Query
	server.Register("workspace.query", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		tenant := manager.TenantFromContext(ctx)
		if tenant == nil {
			return nil, errServer("mtp_error", "No tenant context found")
		}

		var req struct {
			Query          string `json:"query"`
			ConversationID string `json:"conversation_id,omitempty"`
			Model          string `json:"model,omitempty"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}

		// Perform agentic query through the Engine
		answer, modelUsed, err := tenant.Engine.Query(ctx, req.Query, req.ConversationID, req.Model, "chat", "")
		if err != nil {
			return nil, errServer("engine_query_failed", err.Error())
		}

		return map[string]any{
			"answer": answer,
			"model":  modelUsed,
		}, nil
	})

	// [1.1] File System Code Completion (Ghost Text)
	server.Register("fs/completion", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		tenant := manager.TenantFromContext(ctx)
		if tenant == nil {
			return nil, errServer("mtp_error", "No tenant context found")
		}

		var req struct {
			Path     string `json:"path"`
			Prefix   string `json:"prefix"`
			Suffix   string `json:"suffix"`
			Language string `json:"language"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}

		provider := tenant.Engine.LLM.GetDefault()
		if provider == nil || !provider.IsConfigured() {
			return nil, ErrProviderNotConfig
		}

		// Simplified FIM prompt
		prompt := fmt.Sprintf("Complete the code for the file: %s in language: %s.\n\nPREFIX:\n%s\n\nSUFFIX:\n%s\n\nReturn ONLY the code to be inserted.", req.Path, req.Language, req.Prefix, req.Suffix)

		resp, err := provider.Complete(ctx, llm.CompletionRequest{
			Messages: []llm.Message{
				{Role: llm.RoleSystem, Content: "You are an expert software engineer specialized in code completion. Response must contain only the code to be inserted."},
				{Role: llm.RoleUser, Content: prompt},
			},
			MaxTokens:   128,
			Temperature: 0.0,
		})
		if err != nil {
			return nil, errServer("llm_error", err.Error())
		}

		return map[string]any{"content": resp.Content}, nil
	})

	// [1.2] Session Management (ACP Standard)
	server.Register("session/new", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		tenant := manager.TenantFromContext(ctx)
		if tenant == nil {
			return nil, errServer("mtp_error", "No tenant context found")
		}

		// Map MTP tenant ID to ACP session ID
		return map[string]any{"sessionId": tenant.ID}, nil
	})

	server.Register("session/prompt", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		// Bridge session/prompt to workspace.query for consistency
		return server.handlers["workspace.query"](ctx, payload)
	})

	// [2] Chat History
	server.Register("chat.history", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		tenant := manager.TenantFromContext(ctx)
		if tenant == nil {
			return nil, errServer("mtp_error", "No tenant context found")
		}

		var req struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}

		msgService := llm.NewMessageService(tenant.KVStore)
		conv, err := msgService.GetConversation(ctx, req.ID)
		if err != nil {
			return nil, errServer("chat_not_found", err.Error())
		}
		return conv, nil
	})

	// [2.1] List Chats
	server.Register("chat.list", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		tenant := manager.TenantFromContext(ctx)
		if tenant == nil {
			return nil, errServer("mtp_error", "No tenant context found")
		}

		msgService := llm.NewMessageService(tenant.KVStore)
		list, err := msgService.ListConversations(ctx)
		if err != nil {
			return nil, errServer("db_error", err.Error())
		}
		return list, nil
	})

	// [3] Provider Status (Global)
	server.Register("provider.get", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		tenant := manager.TenantFromContext(ctx)
		if tenant == nil {
			return map[string]any{"configured": false}, nil
		}
		p := tenant.Engine.LLM.GetDefault()
		if p == nil {
			return map[string]any{"configured": false}, nil
		}
		return map[string]any{"configured": p.IsConfigured()}, nil
	})

	// [3.1] List Models (Global)
	server.Register("models.list", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		tenant := manager.TenantFromContext(ctx)
		if tenant == nil {
			return nil, errServer("mtp_error", "No tenant context found")
		}
		p := tenant.Engine.LLM.GetDefault()
		if p == nil || !p.IsConfigured() {
			return nil, ErrProviderNotConfig
		}
		list, err := p.ListModels(ctx)
		if err != nil {
			return nil, errServer("llm_error", err.Error())
		}
		return map[string]any{"models": list}, nil
	})

	// [3.2] Provider Reload (Triggered by CLI config)
	server.Register("provider.reload", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		tray.ReloadActiveProvider()
		return map[string]bool{"success": true}, nil
	})

	// [4] Memory Search (Global for now)
	server.Register("memory.search", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		// Note: MemoryService currently uses a global path.
		// Future phases may isolate this too.
		return nil, errServer("not_implemented", "Global memory search is currently disabled in MTP mode.")
	})

	// [5] i18n
	server.Register("i18n.get", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		var req struct {
			Locale string `json:"locale"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}
		if req.Locale != "" {
			i18n.SetLanguage(req.Locale)
		}
		return map[string]any{"lang": i18n.GetCurrentLang()}, nil
	})

	// [6] Health
	server.Register("app.health", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		return map[string]any{"status": "ok", "version": "0.1.0"}, nil
	})

	// [7] Create Chat
	server.Register("chat.create", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		tenant := manager.TenantFromContext(ctx)
		if tenant == nil {
			return nil, errServer("mtp_error", "No tenant context found")
		}

		var req struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}

		msgService := llm.NewMessageService(tenant.KVStore)
		conv, err := msgService.CreateConversation(ctx, req.ID, req.Title)
		if err != nil {
			return nil, errServer("db_error", err.Error())
		}
		return conv, nil
	})

	// [8] Delete Chat
	server.Register("chat.delete", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		tenant := manager.TenantFromContext(ctx)
		if tenant == nil {
			return nil, errServer("mtp_error", "No tenant context found")
		}

		var req struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}

		msgService := llm.NewMessageService(tenant.KVStore)
		if err := msgService.DeleteConversation(ctx, req.ID); err != nil {
			return nil, errServer("db_error", err.Error())
		}
		return map[string]bool{"success": true}, nil
	})

	// [9] Rename Chat
	server.Register("chat.rename", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		tenant := manager.TenantFromContext(ctx)
		if tenant == nil {
			return nil, errServer("mtp_error", "No tenant context found")
		}

		var req struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}

		msgService := llm.NewMessageService(tenant.KVStore)
		if err := msgService.RenameConversation(ctx, req.ID, req.Title); err != nil {
			return nil, errServer("db_error", err.Error())
		}
		return map[string]bool{"success": true}, nil
	})

	// [10] Add Message
	server.Register("message.add", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		tenant := manager.TenantFromContext(ctx)
		if tenant == nil {
			return nil, errServer("mtp_error", "No tenant context found")
		}

		var req struct {
			ConversationID string   `json:"conversationId"`
			Role           llm.Role `json:"role"`
			Content        string   `json:"content"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}

		msgService := llm.NewMessageService(tenant.KVStore)
		err := msgService.AddMessage(ctx, req.ConversationID, req.Role, req.Content)
		if err != nil {
			return nil, errServer("db_error", err.Error())
		}
		return map[string]bool{"success": true}, nil
	})

	// [11] Set Provider (Global Settings)
	server.Register("provider.set", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		if err := globalKV.Set(ctx, "settings", "provider", payload); err != nil {
			return nil, errServer("db_error", err.Error())
		}
		return map[string]bool{"success": true}, nil
	})

	// [13] File Read helper for translations
	server.Register("i18n.translations", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		csvPath := filepath.Join("core", "i18n", "translations.csv")
		content, err := os.ReadFile(csvPath)
		if err != nil {
			return nil, errServer("i18n_not_found", "Locale file not found")
		}

		lines := strings.Split(string(content), "\n")
		if len(lines) < 1 {
			return map[string]any{}, nil
		}
		headers := strings.Split(lines[0], ",")
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

	// [14] Start Embedding Background Job (Per Tenant)
	server.Register("workspace.embed.start", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		tenant := manager.TenantFromContext(ctx)
		if tenant == nil {
			return nil, errServer("mtp_error", "No tenant context found")
		}

		var req struct {
			RootPath string `json:"rootPath"`
			Include  string `json:"include"`
			Exclude  string `json:"exclude"`
			Force    bool   `json:"force"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}

		provider := tenant.Engine.LLM.GetDefault()
		if provider == nil || !provider.IsConfigured() {
			return nil, ErrProviderNotConfig
		}

		// Use tenant root if none provided
		rootPath := req.RootPath
		if rootPath == "" {
			rootPath = tenant.Root
		}

		go engine.RunEmbedJob(
			context.Background(),
			engine.EmbedJobConfig{
				RootPath:       rootPath,
				Include:        req.Include,
				Exclude:        req.Exclude,
				Workspace:      tenant.ID,
				Force:          req.Force,
				CollectionName: "ws_" + salter.HashPath(tenant.ID),
			},
			tenant.KVStore,
			tenant.VectorStore,
			provider,
			func(prog engine.EmbedProgress) {
				server.Broadcast("embed.progress", prog)
			},
		)

		return map[string]bool{"started": true}, nil
	})
}
