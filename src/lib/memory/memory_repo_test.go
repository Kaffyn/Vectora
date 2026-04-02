package memory_test

import (
	"context"
	"testing"

	"github.com/Kaffyn/Vectora/src/core/domain"
	"github.com/Kaffyn/Vectora/src/lib/memory"
)

func TestMemoryChunkRepo_300(t *testing.T) {
	repo := memory.NewMemoryChunkRepo()
	ctx := context.Background()

	// 1. HAPPY PATH: Save and search
	t.Run("HappyPath: Save and Search Chunk", func(t *testing.T) {
		chunk := &domain.Chunk{
			ID:         "chunk_001",
			DocumentID: "doc_001",
			Content:    "The Vectora engine is built in Go.",
			Index:      0,
		}

		err := repo.Save(ctx, chunk)
		if err != nil {
			t.Errorf("failed to save chunk: %v", err)
		}

		results, err := repo.Search(ctx, "Vectora", 5)
		if err != nil {
			t.Errorf("failed to search: %v", err)
		}

		if len(results) == 0 {
			t.Error("expected search results, got none")
		}

		if results[0].Content != chunk.Content {
			t.Errorf("expected content %s, got %s", chunk.Content, results[0].Content)
		}
	})

	// 2. NEGATIVE: Empty search or invalid characters
	t.Run("Negative: Search for Empty Query", func(t *testing.T) {
		_, err := repo.Search(ctx, "", 5)
		if err == nil {
			t.Error("expected error for empty search, got nil")
		}
	})

	t.Run("Negative: Save Empty Content Chunk", func(t *testing.T) {
		err := repo.Save(ctx, &domain.Chunk{Content: ""})
		if err == nil {
			t.Error("expected error saving empty content, got nil")
		}
	})

	// 3. EDGE CASE: Large amount of data search
	t.Run("EdgeCase: Multiple Chunks (1000)", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			err := repo.Save(ctx, &domain.Chunk{
				DocumentID: "doc_edge",
				Content:    "Edge Case Content " + string(rune(i)),
				Index:      i,
			})
			if err != nil {
				t.Fatalf("failed to save chunk %d: %v", i, err)
			}
		}

		results, err := repo.Search(ctx, "Edge Case Content", 5)
		if err != nil {
			t.Fatalf("failed to search in edge case: %v", err)
		}

		if len(results) != 5 {
			t.Errorf("expected 5 results due to limit, got %d", len(results))
		}
	})
}
