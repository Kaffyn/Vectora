package filestore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

type JSONConversationRepo struct {
	baseDir string
	mu      sync.RWMutex
}

func NewJSONConversationRepo(baseDir string) (*JSONConversationRepo, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("could not create base directory: %w", err)
	}
	return &JSONConversationRepo{baseDir: baseDir}, nil
}

func (r *JSONConversationRepo) Save(ctx context.Context, conv *domain.Conversation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if conv.ID == "" {
		return errors.New("conversation ID cannot be empty")
	}

	path := filepath.Join(r.baseDir, conv.ID+".json")
	data, err := json.MarshalIndent(conv, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal conversation: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

func (r *JSONConversationRepo) GetByID(ctx context.Context, id string) (*domain.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	path := filepath.Join(r.baseDir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("conversation not found: %s", id)
		}
		return nil, fmt.Errorf("could not read conversation file: %w", err)
	}

	var conv domain.Conversation
	if err := json.Unmarshal(data, &conv); err != nil {
		return nil, fmt.Errorf("could not unmarshal conversation: %w", err)
	}

	return &conv, nil
}

func (r *JSONConversationRepo) List(ctx context.Context) ([]*domain.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	files, err := os.ReadDir(r.baseDir)
	if err != nil {
		return nil, fmt.Errorf("could not read base directory: %w", err)
	}

	var conversations []*domain.Conversation
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".json" {
			id := f.Name()[:len(f.Name())-5]
			conv, err := r.GetByID(ctx, id)
			if err == nil {
				conversations = append(conversations, conv)
			}
		}
	}

	return conversations, nil
}

func (r *JSONConversationRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	path := filepath.Join(r.baseDir, id+".json")
	return os.Remove(path)
}
