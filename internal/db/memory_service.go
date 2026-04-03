package db

import (
	"context"
	"fmt"
	"runtime"
	"github.com/philippgille/chromem-go"
)

// MemoryService manages the indexing of volatile and permanent knowledge.
type MemoryService struct {
	db         *chromem.DB
	collection *chromem.Collection
}

func NewMemoryService(ctx context.Context, dbPath string) (*MemoryService, error) {
	db, err := chromem.NewPersistentDB(dbPath, false)
	if err != nil {
		return nil, fmt.Errorf("memory_db_init_failed: %w", err)
	}
	col, err := db.GetOrCreateCollection("knowledge_base", nil, nil)
	if err != nil {
		return nil, err
	}

	return &MemoryService{
		db:         db,
		collection: col,
	}, nil
}

// StoreInsight guarda um fato ou busca no banco vetorial.
func (s *MemoryService) StoreInsight(ctx context.Context, id, content string, metadata map[string]string) error {
	doc := chromem.Document{
		ID:       id,
		Content:  content,
		Metadata: metadata,
	}
	return s.collection.AddDocuments(ctx, []chromem.Document{doc}, runtime.NumCPU())
}

// SearchInsight performs local RAG over accumulated knowledge.
func (s *MemoryService) SearchInsight(ctx context.Context, query string, topK int) ([]string, error) {
	results, err := s.collection.Query(ctx, query, topK, nil, nil)
	if err != nil {
		return nil, err
	}

	var insights []string
	for _, res := range results {
		insights = append(insights, res.Content)
	}
	return insights, nil
}
