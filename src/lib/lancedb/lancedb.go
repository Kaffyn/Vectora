package lancedb

import (
	"context"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

type LanceDBChunkRepository struct {
	path string
}

func NewLanceDBChunkRepository(ctx context.Context, path string, collection string) (*LanceDBChunkRepository, error) {
	return &LanceDBChunkRepository{path: path}, nil
}

func (r *LanceDBChunkRepository) Save(ctx context.Context, chunk *domain.Chunk, embedding []float32) error {
	return nil
}

func (r *LanceDBChunkRepository) Search(ctx context.Context, query string, limit int) ([]*domain.Chunk, error) {
	return nil, nil
}

func (r *LanceDBChunkRepository) VectorSearch(ctx context.Context, embedding []float32, limit int) ([]*domain.Chunk, error) {
	return nil, nil
}

func (r *LanceDBChunkRepository) GetByDocumentID(ctx context.Context, id string) ([]*domain.Chunk, error) {
	return nil, nil
}
