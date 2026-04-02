package domain_test

import (
	"testing"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

func TestNewDocument(t *testing.T) {
	t.Run("should create a valid document", func(t *testing.T) {
		doc, err := domain.NewDocument("res://scripts/player.gd", "extends Node", "gdscript")

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if doc.FilePath != "res://scripts/player.gd" {
			t.Errorf("expected path res://scripts/player.gd, got %s", doc.FilePath)
		}
	})

	t.Run("should fail with empty content", func(t *testing.T) {
		_, err := domain.NewDocument("test.gd", "", "gdscript")
		if err == nil {
			t.Error("expected error for empty content, got nil")
		}
	})
}
