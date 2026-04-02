package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Kaffyn/Vectora/src/core/rag"
)

type ChatHandler struct {
	chatService *rag.ChatService
}

func NewChatHandler(c *rag.ChatService) *ChatHandler {
	return &ChatHandler{chatService: c}
}

type ChatRequest struct {
	Message        string `json:"message"`
	ConversationID string `json:"conversationId"`
}

type ChatResponse struct {
	Reply   string       `json:"reply"`
	Sources []ChatSource `json:"sources"`
}

type ChatSource struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Path    string `json:"path"`
}

func (h *ChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"only POST allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, "invalid body", http.StatusBadRequest)
		return
	}

	// 1. Delegar todo o RAG + LLM para o ChatService
	reply, err := h.chatService.SendMessage(r.Context(), req.ConversationID, req.Message)
	if err != nil {
		fmt.Printf("[Chat] LLM generation failed: %v\n", err)
		http.Error(w, `{"error":"failed to generate response"}`, http.StatusInternalServerError)
		return
	}

	sources := make([]ChatSource, 0)
	// Como o SendMessage internamente faz RAG e incorpora na resposta,
	// podemos retornar fontes se extrairmos no futuro. Por ora, retornamos as fontes geradas mockadas
	// para manter contrato com frontend

	sources = append(sources, ChatSource{
		Title:   "Qwen3 Engine Local",
		Content: "Vetorizado e respondido nativamente.",
		Path:    "local://qwen-engine",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ChatResponse{
		Reply:   reply,
		Sources: sources,
	})
}
