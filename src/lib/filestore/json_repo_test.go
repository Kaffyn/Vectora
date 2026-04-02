package filestore_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Kaffyn/Vectora/src/core/domain"
	"github.com/Kaffyn/Vectora/src/lib/filestore"
)

func TestJSONConversationRepo_300(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "conversation_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	repo, err := filestore.NewJSONConversationRepo(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// 1. HAPPY PATH: Create and recover a conversation
	t.Run("HappyPath: Save and Get Conversation", func(t *testing.T) {
		conv := domain.NewConversation("conf_001")
		conv.AddMessage("user", "Hello Vectora")
		conv.AddMessage("assistant", "Hello! How can I help you today?")

		err := repo.Save(ctx, conv)
		if err != nil {
			t.Errorf("expected no error saving, got %v", err)
		}

		got, err := repo.GetByID(ctx, "conf_001")
		if err != nil {
			t.Errorf("expected no error getting, got %v", err)
		}

		if len(got.Messages) != 2 {
			t.Errorf("expected 2 messages, got %d", len(got.Messages))
		}
	})

	// 2. NEGATIVE: Try to read a non-existent session or corrupt data
	t.Run("Negative: Get Non-existent Session", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "invalid_id")
		if err == nil {
			t.Error("expected error for non-existent session, got nil")
		}
	})

	t.Run("Negative: Corrupt Data", func(t *testing.T) {
		err := os.WriteFile(filepath.Join(tempDir, "corrupt.json"), []byte("{invalid json}"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		_, err = repo.GetByID(ctx, "corrupt")
		if err == nil {
			t.Error("expected error for corrupt JSON, got nil")
		}
	})

	// 3. EDGE CASE: Large conversation
	t.Run("EdgeCase: Large Conversation (1000 messages)", func(t *testing.T) {
		conv := domain.NewConversation("large_session")
		for i := 0; i < 1000; i++ {
			conv.AddMessage("user", "Message number "+string(rune(i)))
		}

		err := repo.Save(ctx, conv)
		if err != nil {
			t.Fatalf("failed to save large conversation: %v", err)
		}

		got, err := repo.GetByID(ctx, "large_session")
		if err != nil {
			t.Fatalf("failed to get large conversation: %v", err)
		}

		if len(got.Messages) != 1000 {
			t.Errorf("expected 1000 messages, got %d", len(got.Messages))
		}
	})
}
