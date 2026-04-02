package rag_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Kaffyn/Vectora/src/core/domain"
	"github.com/Kaffyn/Vectora/src/core/rag"
)

type MockDocumentRepo struct {
	ShouldFail bool
}

func (m *MockDocumentRepo) Save(ctx context.Context, doc *domain.Document) error {
	if m.ShouldFail {
		return errors.New("save failed")
	}
	return nil
}

func (m *MockDocumentRepo) GetByID(ctx context.Context, id string) (*domain.Document, error) {
	return nil, nil
}

func (m *MockDocumentRepo) GetByPath(ctx context.Context, path string) (*domain.Document, error) {
	return nil, nil
}

type MockEmbedder struct {
	ShouldFail bool
}

func (m *MockEmbedder) Generate(ctx context.Context, text string) ([]float32, error) {
	if m.ShouldFail {
		return nil, errors.New("embedding failed")
	}
	return []float32{0.1, 0.2, 0.3}, nil
}

func TestIngestService_300(t *testing.T) {
	docRepo := &MockDocumentRepo{}
	chunkRepo := &MockChunkRepo{}
	embedder := &MockEmbedder{}
	service := rag.NewIngestService(docRepo, chunkRepo, embedder)

	t.Run("HappyPath: Ingest Valid File", func(t *testing.T) {
		doc := &domain.Document{ID: "doc_1", FilePath: "test.gd", Content: "func _ready(): pass"}
		err := service.Ingest(context.Background(), doc)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if chunkRepo.LastSavedChunk == nil {
			t.Fatal("expected chunk to be saved")
		}
		if len(chunkRepo.LastSavedEmbedding) == 0 {
			t.Fatal("expected embedding to be saved")
		}
	})

	t.Run("Negative: Save Document Fails", func(t *testing.T) {
		badDocRepo := &MockDocumentRepo{ShouldFail: true}
		service := rag.NewIngestService(badDocRepo, chunkRepo, embedder)
		doc := &domain.Document{ID: "doc_err", FilePath: "err.gd", Content: "error"}

		err := service.Ingest(context.Background(), doc)
		if err == nil {
			t.Error("expected error when saving fails, got nil")
		}
	})

	t.Run("Negative: Embedding Generation Fails", func(t *testing.T) {
		badEmbedder := &MockEmbedder{ShouldFail: true}
		service := rag.NewIngestService(docRepo, chunkRepo, badEmbedder)
		doc := &domain.Document{ID: "doc_err_emb", FilePath: "err_emb.gd", Content: "error"}

		err := service.Ingest(context.Background(), doc)
		if err == nil {
			t.Error("expected error when embedding fails, got nil")
		}
	})

	t.Run("EdgeCase: Ingest Empty Document", func(t *testing.T) {
		doc := &domain.Document{ID: "doc_empty", FilePath: "empty.gd", Content: ""}
		err := service.Ingest(context.Background(), doc)
		if err == nil {
			t.Log("Accepted empty document (expected 0 chunks)")
		}
	})

	t.Run("EdgeCase: Ingest Large Document", func(t *testing.T) {
		largeContent := ""
		for i := 0; i < 1000; i++ {
			largeContent += "func line(): pass\n"
		}
		doc := &domain.Document{ID: "doc_large", FilePath: "large.gd", Content: largeContent}
		err := service.Ingest(context.Background(), doc)
		if err != nil {
			t.Errorf("expected no error for large doc, got %v", err)
		}
	})
}
