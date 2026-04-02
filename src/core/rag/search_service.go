package rag

import (
	"context"
	"fmt"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

type SearchService struct {
	metadataRepo domain.ChunkRepository
	vectorRepo   domain.ChunkRepository
}

func NewSearchService(metadata domain.ChunkRepository, vector domain.ChunkRepository) *SearchService {
	return &SearchService{
		metadataRepo: metadata,
		vectorRepo:   vector,
	}
}

func (s *SearchService) VectorSearch(ctx context.Context, embedding []float32, limit int) ([]*domain.Chunk, error) {
	if s.vectorRepo == nil {
		return nil, nil
	}
	return s.vectorRepo.VectorSearch(ctx, embedding, limit)
}

func (s *SearchService) Search(ctx context.Context, query string) ([]*domain.Chunk, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	// Passo 1: Tentar encontrar via Metadados/Full-text (bbolt - In-Process)
	results, err := s.metadataRepo.Search(ctx, query, 5)
	if err != nil {
		return nil, fmt.Errorf("[RAG] Metadata search error (bbolt): %w", err)
	}

	return results, nil
}
