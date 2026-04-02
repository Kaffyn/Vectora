package chromem_test

import (
	"context"
	"testing"

	"github.com/Kaffyn/Vectora/src/core/domain"
	"github.com/Kaffyn/Vectora/src/lib/chromem"
)

func TestVectorRepo_TDD(t *testing.T) {
	ctx := context.Background()
	repo, err := chromem.NewVectorRepo("", "test_collection")
	if err != nil {
		t.Fatalf("Failed to create VectorRepo: %v", err)
	}

	// 1. HAPPY PATH: Save and VectorSearch
	t.Run("HappyPath_SaveAndSearch", func(t *testing.T) {
		chunk := &domain.Chunk{
			ID:       "chunk-1",
			Content:  "Hello World",
			Metadata: map[string]string{"key": "val"},
		}
		embedding := []float32{0.1, 0.2, 0.3}

		if err := repo.Save(ctx, chunk, embedding); err != nil {
			t.Errorf("Failed to save: %v", err)
		}

		results, err := repo.VectorSearch(ctx, embedding, 1)
		if err != nil {
			t.Errorf("Search error: %v", err)
		}

		if len(results) == 0 {
			t.Fatal("Expected 1 result, got 0")
		}
		if results[0].ID != "chunk-1" {
			t.Errorf("Expected chunk-1, got %s", results[0].ID)
		}
	})

	// ... Negative and Edge subtests can follow
}
