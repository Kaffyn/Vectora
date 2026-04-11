package ipc

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/Kaffyn/Vectora/core/crypto"
	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/engine"
	"github.com/Kaffyn/Vectora/core/i18n"
	"github.com/Kaffyn/Vectora/core/llm"
)

type ProviderFetcher func() llm.Provider

func RegisterRoutes(
	server *Server,
	kvStore db.KVStore,
	vecStore db.VectorStore,
	getProvider ProviderFetcher,
	msgService *llm.MessageService,
	memService *db.MemoryService,
	salter *crypto.WorkspaceSalter,
) {
	// [1] Main RAG Query
	server.Register("workspace.query", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		var req struct {
			WorkspaceID    string `json:"workspace_id"`
			Query          string `json:"query"`
			ConversationID string `json:"conversation_id,omitempty"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}

		provider := getProvider()
		if provider == nil || !provider.IsConfigured() {
			return nil, ErrProviderNotConfig
		}

		vector, err := provider.Embed(ctx, req.Query)
		if err != nil {
			return nil, errServer("embed_failed", err.Error())
		}

		// Hash workspace ID with installation salt for per-machine uniqueness
		collectionID := "ws_" + salter.HashPath(req.WorkspaceID)
		chunks, err := vecStore.Query(ctx, collectionID, vector, 5)
		if err != nil {
			chunks = []db.ScoredChunk{}
		}

		contextText := ""
		for _, doc := range chunks {
			if filename, ok := doc.Metadata["filename"]; ok {
				contextText += "File: " + filename + "\n"
			}
			contextText += doc.Content + "\n---\n"
		}

		var messages []llm.Message

		if req.ConversationID != "" {
			// Load or create conversation
			conv, err := msgService.GetConversation(ctx, req.ConversationID)
			if err != nil {
				// First turn: create it
				msgService.CreateConversation(ctx, req.ConversationID, "chat")
				messages = append(messages, llm.Message{
					Role:    llm.RoleSystem,
					Content: "You are Vectora. Use the following context:\n" + contextText,
				})
			} else {
				// Inject system prompt first, then history
				messages = append(messages, llm.Message{
					Role:    llm.RoleSystem,
					Content: "You are Vectora. Use the following context:\n" + contextText,
				})
				for _, m := range conv.Messages {
					messages = append(messages, llm.Message{Role: m.Role, Content: m.Content})
				}
			}
			msgService.AddMessage(ctx, req.ConversationID, llm.RoleUser, req.Query)
		} else {
			messages = append(messages, llm.Message{
				Role:    llm.RoleSystem,
				Content: "You are Vectora. Use the following context:\n" + contextText,
			})
		}

		messages = append(messages, llm.Message{Role: llm.RoleUser, Content: req.Query})

		resp, err := provider.Complete(ctx, llm.CompletionRequest{
			Messages:    messages,
			MaxTokens:   1500,
			Temperature: 0.1,
		})
		if err != nil {
			return nil, errServer("llm_failed", err.Error())
		}

		if req.ConversationID != "" {
			msgService.AddMessage(ctx, req.ConversationID, llm.RoleAssistant, resp.Content)
		}

		return map[string]any{"answer": resp.Content, "sources": chunks}, nil
	})

	// [2] Chat History
	server.Register("chat.history", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		var req struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}

		conv, err := msgService.GetConversation(ctx, req.ID)
		if err != nil {
			return nil, errServer("chat_not_found", err.Error())
		}
		return conv, nil
	})

	// [2.1] List Chats
	server.Register("chat.list", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		list, err := msgService.ListConversations(ctx)
		if err != nil {
			return nil, errServer("db_error", err.Error())
		}
		return list, nil
	})

	// [3] Provider Status
	server.Register("provider.get", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		p := getProvider()
		if p == nil {
			return map[string]any{"configured": false}, nil
		}
		return map[string]any{"configured": p.IsConfigured()}, nil
	})

	// [4] Memory Search
	server.Register("memory.search", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		var req struct {
			Query string `json:"query"`
			TopK  int    `json:"top_k"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}

		results, err := memService.SearchInsight(ctx, req.Query, nil, req.TopK)
		if err != nil {
			return nil, errServer("memory_error", err.Error())
		}
		return results, nil
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
		var req struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}
		conv, err := msgService.CreateConversation(ctx, req.ID, req.Title)
		if err != nil {
			return nil, errServer("db_error", err.Error())
		}
		return conv, nil
	})

	// [8] Delete Chat
	server.Register("chat.delete", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		var req struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}
		if err := msgService.DeleteConversation(ctx, req.ID); err != nil {
			return nil, errServer("db_error", err.Error())
		}
		return map[string]bool{"success": true}, nil
	})

	// [9] Rename Chat
	server.Register("chat.rename", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		var req struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}
		if err := msgService.RenameConversation(ctx, req.ID, req.Title); err != nil {
			return nil, errServer("db_error", err.Error())
		}
		return map[string]bool{"success": true}, nil
	})

	// [10] Add Message
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
			return nil, errServer("db_error", err.Error())
		}
		return map[string]bool{"success": true}, nil
	})

	// [11] Set Provider
	server.Register("provider.set", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		if err := kvStore.Set(ctx, "settings", "provider", payload); err != nil {
			return nil, errServer("db_error", err.Error())
		}
		return map[string]bool{"success": true}, nil
	})

	// [12] File Read helper for translations
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

	// [13] Start Embedding Background Job
	server.Register("workspace.embed.start", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		var req struct {
			RootPath  string `json:"rootPath"`
			Include   string `json:"include"`
			Exclude   string `json:"exclude"`
			Workspace string `json:"workspace"`
			Force     bool   `json:"force"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, ErrIPCPayloadInvalid
		}

		provider := getProvider()
		if provider == nil || !provider.IsConfigured() {
			return nil, ErrProviderNotConfig
		}

		// Import statement required below but since this is within RegisterRoutes we ensure we import "github.com/Kaffyn/Vectora/core/engine"
		// at the top of router.go. I'll modify the imports via another call if needed.
		go func() {
			engine.RunEmbedJob(
				context.Background(),
				engine.EmbedJobConfig{
					RootPath:       req.RootPath,
					Include:        req.Include,
					Exclude:        req.Exclude,
					Workspace:      req.Workspace,
					Force:          req.Force,
					CollectionName: "ws_" + salter.HashPath(req.Workspace),
				},
				kvStore,
				vecStore,
				provider,
				func(prog engine.EmbedProgress) {
					server.Broadcast("embed.progress", prog)
				},
			)
		}()

		return map[string]bool{"started": true}, nil
	})
}
