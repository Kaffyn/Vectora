package server

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/Kaffyn/Vectora/src/core/ai"
)

type Settings struct {
	ActiveTextModel      string `json:"active_text_model"`
	ActiveEmbeddingModel string `json:"active_embedding_model"`
}

type SettingsHandler struct {
	mu             sync.RWMutex
	settings       Settings
	sidecarManager *ai.SidecarManager
}

func NewSettingsHandler(initialText, initialEmbedding string, sm *ai.SidecarManager) *SettingsHandler {
	return &SettingsHandler{
		settings: Settings{
			ActiveTextModel:      initialText,
			ActiveEmbeddingModel: initialEmbedding,
		},
		sidecarManager: sm,
	}
}

func (h *SettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method == http.MethodGet {
		h.mu.RLock()
		defer h.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(h.settings)
		return
	}

	if r.Method == http.MethodPost {
		var req Settings
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
			return
		}

		h.mu.Lock()
		textSwapNeeded := req.ActiveTextModel != "" && req.ActiveTextModel != h.settings.ActiveTextModel
		if req.ActiveTextModel != "" {
			h.settings.ActiveTextModel = req.ActiveTextModel
		}

		// (embedding swapping removed for brevity but could be added the same way)

		state := h.settings
		h.mu.Unlock()

		// Executar o Hot Swap fora do lock se o modelo de texto mudou
		if textSwapNeeded && h.sidecarManager != nil {
			// Na inicialização definimos porta padrão 8081 para texto e false no vulkan. Isso pode ser refatorado se os dados vierem do req
			_ = h.sidecarManager.EnsureRunning(context.Background(), "text", state.ActiveTextModel, 8081, false)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(state)
		return
	}

	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}

// GetSettings permite obter de forma segura as configs a partir de outros services
func (h *SettingsHandler) GetSettings() Settings {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.settings
}
