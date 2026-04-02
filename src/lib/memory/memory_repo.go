package memory

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

type MemoryChunkRepo struct {
	chunks map[string]*domain.Chunk
	mu     sync.RWMutex
}

func NewMemoryChunkRepo() *MemoryChunkRepo {
	return &MemoryChunkRepo{
		chunks: make(map[string]*domain.Chunk),
	}
}

func (r *MemoryChunkRepo) Save(ctx context.Context, chunk *domain.Chunk) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if chunk.Content == "" {
		return errors.New("cannot save empty chunk")
	}

	id := fmt.Sprintf("%s_%d", chunk.DocumentID, chunk.Index)
	r.chunks[id] = chunk
	return nil
}

func (r *MemoryChunkRepo) Search(ctx context.Context, query string, limit int) ([]*domain.Chunk, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if query == "" {
		return nil, errors.New("search query cannot be empty")
	}

	var results []*domain.Chunk
	for _, c := range r.chunks {
		if strings.Contains(strings.ToLower(c.Content), strings.ToLower(query)) {
			results = append(results, c)
		}
		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

func (r *MemoryChunkRepo) GetByDocumentID(ctx context.Context, docID string) ([]*domain.Chunk, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*domain.Chunk
	for _, c := range r.chunks {
		if c.DocumentID == docID {
			results = append(results, c)
		}
	}

	return results, nil
}
