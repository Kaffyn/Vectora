package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

// ConversationHandler handles all /api/conversations/* routes.
type ConversationHandler struct {
	repo domain.ConversationRepository
}

func NewConversationHandler(repo domain.ConversationRepository) *ConversationHandler {
	return &ConversationHandler{repo: repo}
}

// ServeHTTP is the single entry-point; dispatches by method and path suffix.
func (h *ConversationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Enable CORS for the Next.js dev server.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	// /api/conversations          → list / create
	// /api/conversations/{id}     → get / update / delete
	// /api/conversations/{id}/messages → append message
	path := strings.TrimPrefix(r.URL.Path, "/api/conversations")
	path = strings.TrimPrefix(path, "/")
	parts := strings.SplitN(path, "/", 3)
	id := ""
	suffix := ""
	if len(parts) > 0 {
		id = parts[0]
	}
	if len(parts) > 1 {
		suffix = parts[1]
	}

	switch {
	case id == "" && r.Method == http.MethodGet:
		h.listConversations(w, r)
	case id == "" && r.Method == http.MethodPost:
		h.createConversation(w, r)
	case id != "" && suffix == "" && r.Method == http.MethodGet:
		h.getConversation(w, r, id)
	case id != "" && suffix == "" && (r.Method == http.MethodPut || r.Method == http.MethodPatch):
		h.updateConversation(w, r, id)
	case id != "" && suffix == "" && r.Method == http.MethodDelete:
		h.deleteConversation(w, r, id)
	case id != "" && suffix == "messages" && r.Method == http.MethodPost:
		h.appendMessage(w, r, id)
	default:
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
	}
}

func (h *ConversationHandler) listConversations(w http.ResponseWriter, r *http.Request) {
	convs, err := h.repo.List(r.Context())
	if err != nil {
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if convs == nil {
		convs = []*domain.Conversation{}
	}
	json.NewEncoder(w).Encode(convs)
}

func (h *ConversationHandler) createConversation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, "invalid body", http.StatusBadRequest)
		return
	}
	now := time.Now()
	conv := &domain.Conversation{
		ID:        req.ID,
		Title:     req.Title,
		Messages:  []domain.Message{},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := h.repo.Save(r.Context(), conv); err != nil {
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(conv)
}

func (h *ConversationHandler) getConversation(w http.ResponseWriter, r *http.Request, id string) {
	conv, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		jsonErr(w, "not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(conv)
}

func (h *ConversationHandler) updateConversation(w http.ResponseWriter, r *http.Request, id string) {
	conv, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		jsonErr(w, "not found", http.StatusNotFound)
		return
	}
	var req struct {
		Title    *string          `json:"title"`
		Messages []domain.Message `json:"messages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Title != nil {
		conv.Title = *req.Title
	}
	if req.Messages != nil {
		conv.Messages = req.Messages
	}
	conv.UpdatedAt = time.Now()
	if err := h.repo.Save(r.Context(), conv); err != nil {
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(conv)
}

func (h *ConversationHandler) deleteConversation(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.repo.Delete(r.Context(), id); err != nil {
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ConversationHandler) appendMessage(w http.ResponseWriter, r *http.Request, id string) {
	conv, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		// Auto-create if does not exist yet
		now := time.Now()
		conv = &domain.Conversation{ID: id, Title: "New Conversation", Messages: []domain.Message{}, CreatedAt: now, UpdatedAt: now}
	}
	var msg domain.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		jsonErr(w, "invalid body", http.StatusBadRequest)
		return
	}
	msg.Timestamp = time.Now()
	conv.Messages = append(conv.Messages, msg)
	conv.UpdatedAt = time.Now()
	if err := h.repo.Save(r.Context(), conv); err != nil {
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(conv)
}

func jsonErr(w http.ResponseWriter, msg string, code int) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
