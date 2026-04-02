package bbolt_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Kaffyn/Vectora/src/core/domain"
	"github.com/Kaffyn/Vectora/src/lib/bbolt"
	bolt "go.etcd.io/bbolt"
)

func setupTestDB(t *testing.T) (*bolt.DB, func()) {
	tmpDir := os.TempDir()
	dbPath := filepath.Join(tmpDir, "vectora_test_bbolt.db")
	db, err := bolt.Open(dbPath, 0666, nil)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return db, cleanup
}

func TestConversationRepo_300Percent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := bbolt.NewConversationRepo(db)
	if err != nil {
		t.Fatalf("Failed to create repo: %v", err)
	}

	ctx := context.Background()

	// 1. HAPPY PATH: Save and Retrieve
	t.Run("HappyPath_SaveAndGet", func(t *testing.T) {
		conv := &domain.Conversation{
			ID:        "conv-1",
			Title:     "Test Chat",
			CreatedAt: time.Now(),
		}

		if err := repo.Save(ctx, conv); err != nil {
			t.Errorf("Failed to save: %v", err)
		}

		got, err := repo.GetByID(ctx, "conv-1")
		if err != nil {
			t.Errorf("Failed to get: %v", err)
		}

		if got.Title != conv.Title {
			t.Errorf("Expected title %s, got %s", conv.Title, got.Title)
		}
	})

	// 2. NEGATIVE: Get non-existent conversation
	t.Run("Negative_GetNonExistent", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "missing")
		if err == nil {
			t.Error("Expected error for non-existent conversation, got nil")
		}
	})

	// 3. EDGE CASE: Save large content / Bulk simulation
	t.Run("EdgeCase_BulkSave", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			id := "bulk-" + string(rune(i))
			repo.Save(ctx, &domain.Conversation{ID: id})
		}
		list, err := repo.List(ctx)
		if err != nil {
			t.Errorf("Failed to list bulk: %v", err)
		}
		if len(list) < 100 {
			t.Errorf("Expected at least 100 convs, got %d", len(list))
		}
	})
}
