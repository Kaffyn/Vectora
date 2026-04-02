package rag_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Kaffyn/Vectora/src/core/domain"
	"github.com/Kaffyn/Vectora/src/core/rag"
)

// Manual Mock for ChunkRepository
type MockChunkRepo struct {
	Chunks             []*domain.Chunk
	ShouldFail         bool
	LastSavedChunk     *domain.Chunk
	LastSavedEmbedding []float32
}

func (m *MockChunkRepo) Save(ctx context.Context, c *domain.Chunk, e []float32) error {
	m.LastSavedChunk = c
	m.LastSavedEmbedding = e
	return nil
}
func (m *MockChunkRepo) Search(ctx context.Context, q string, l int) ([]*domain.Chunk, error) {
	if m.ShouldFail {
		return nil, errors.New("database connection failed")
	}
	return m.Chunks, nil
}
func (m *MockChunkRepo) VectorSearch(ctx context.Context, e []float32, l int) ([]*domain.Chunk, error) {
	if m.ShouldFail {
		return nil, errors.New("vector database connection failed")
	}
	return m.Chunks, nil
}
func (m *MockChunkRepo) GetByDocumentID(ctx context.Context, id string) ([]*domain.Chunk, error) {
	return nil, nil
}

func TestSearchService_300(t *testing.T) {
	t.Run("HappyPath: Simple Search", func(t *testing.T) {
		repo := &MockChunkRepo{
			Chunks: []*domain.Chunk{
				{Content: "ASAbility explanation", Index: 0},
			},
		}
		service := rag.NewSearchService(repo, nil)

		results, err := service.Search(context.Background(), "ASAbility")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(results) != 1 {
			t.Errorf("expected 1 result, got %d", len(results))
		}
	})

	t.Run("Negative: Empty Query", func(t *testing.T) {
		repo := &MockChunkRepo{}
		service := rag.NewSearchService(repo, nil)

		_, err := service.Search(context.Background(), "")
		if err == nil {
			t.Error("expected error for empty query, got nil")
		}
	})

	t.Run("Negative: Database Down", func(t *testing.T) {
		repo := &MockChunkRepo{ShouldFail: true}
		service := rag.NewSearchService(repo, nil)

		_, err := service.Search(context.Background(), "Something")
		if err == nil {
			t.Error("expected database error, got nil")
		}
	})

	t.Run("EdgeCase: Extremely Long Query", func(t *testing.T) {
		repo := &MockChunkRepo{Chunks: []*domain.Chunk{}}
		service := rag.NewSearchService(repo, nil)

		longQuery := "this is a very long query..."
		_, err := service.Search(context.Background(), longQuery)
		if err != nil {
			t.Errorf("expected no crash for long query, got error %v", err)
		}
	})

	t.Run("EdgeCase: Query with No Matches", func(t *testing.T) {
		repo := &MockChunkRepo{Chunks: []*domain.Chunk{}}
		service := rag.NewSearchService(repo)

		results, err := service.Search(context.Background(), "non-existent-content")
		if err != nil {
			t.Errorf("expected no error for no matches, got %v", err)
		}
		if len(results) != 0 {
			t.Errorf("expected 0 results, got %d", len(results))
		}
	})
}
