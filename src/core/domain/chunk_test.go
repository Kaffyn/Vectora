package domain_test

import (
	"testing"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

func TestNewChunk(t *testing.T) {
	doc, _ := domain.NewDocument("res://scripts/player.gd", "extends Node", "gdscript")

	t.Run("should create a valid chunk", func(t *testing.T) {
		chunk, err := domain.NewChunk(doc, "extends Node", 0)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if chunk.Content != "extends Node" {
			t.Errorf("expected content 'extends Node', got %s", chunk.Content)
		}

		if chunk.DocumentID != doc.ID {
			t.Errorf("expected document ID mismatch")
		}
	})

	t.Run("should fail with empty content", func(t *testing.T) {
		_, err := domain.NewChunk(doc, "", 0)
		if err == nil {
			t.Error("expected error for empty content")
		}
	})
}
