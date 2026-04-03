package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Kaffyn/vectora/internal/db"
	"github.com/google/uuid"
)

// MessageService manages the conversation lifecycle, persisting them in a KVStore.
type MessageService struct {
	store db.KVStore
}

func NewMessageService(store db.KVStore) *MessageService {
	return &MessageService{store: store}
}

// CreateConversation generates a new chat session with an optional custom ID and title.
func (s *MessageService) CreateConversation(ctx context.Context, id, title string) (*Conversation, error) {
	if id == "" {
		id = uuid.New().String()
	}
	conv := &Conversation{
		ID:        id,
		Title:     title,
		Messages:  []ChatMessage{},
		UpdatedAt: time.Now(),
	}
	return conv, s.SaveConversation(ctx, conv)
}

// RenameConversation updates ONLY the title metadata.
func (s *MessageService) RenameConversation(ctx context.Context, id, newTitle string) error {
	conv, err := s.GetConversation(ctx, id)
	if err != nil {
		return err
	}
	conv.Title = newTitle
	conv.UpdatedAt = time.Now()
	return s.SaveConversation(ctx, conv)
}

// DeleteConversation removes a session from the persistent store.
func (s *MessageService) DeleteConversation(ctx context.Context, id string) error {
	return s.store.Delete(ctx, "conversations", id)
}

// GetConversation retrieves a chat from the store.
func (s *MessageService) GetConversation(ctx context.Context, id string) (*Conversation, error) {
	data, err := s.store.Get(ctx, "conversations", id)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("chat_not_found: %s", id)
	}

	var conv Conversation
	if err := json.Unmarshal(data, &conv); err != nil {
		return nil, err
	}
	return &conv, nil
}

// AddMessage appends a new interaction to the history and persists it.
func (s *MessageService) AddMessage(ctx context.Context, convID string, role Role, content string) error {
	conv, err := s.GetConversation(ctx, convID)
	if err != nil {
		return err
	}

	msg := ChatMessage{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}

	conv.Messages = append(conv.Messages, msg)
	conv.UpdatedAt = time.Now()

	return s.SaveConversation(ctx, conv)
}

// SaveConversation serializes and saves to the store.
func (s *MessageService) SaveConversation(ctx context.Context, conv *Conversation) error {
	data, err := json.Marshal(conv)
	if err != nil {
		return err
	}
	return s.store.Set(ctx, "conversations", conv.ID, data)
}

// ListConversations returns the IDs of all saved sessions.
func (s *MessageService) ListConversations(ctx context.Context) ([]string, error) {
	return s.store.List(ctx, "conversations", "")
}
