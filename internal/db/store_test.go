package db

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func createTestStore(t *testing.T) (*Store, string) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	logger := slog.Default()

	store, err := NewStore(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	return store, tmpDir
}

func TestNewStore(t *testing.T) {
	store, tmpDir := createTestStore(t)
	defer store.Close()
	defer os.RemoveAll(tmpDir)

	if store == nil {
		t.Fatal("NewStore returned nil")
	}
	if store.db == nil {
		t.Fatal("Database not initialized")
	}
}

func TestSaveAndGetWorkspace(t *testing.T) {
	store, tmpDir := createTestStore(t)
	defer store.Close()
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	testData := map[string]interface{}{
		"name": "Test Workspace",
		"id":   "ws-1",
	}

	// Save workspace
	err := store.SaveWorkspace(ctx, "ws-1", testData)
	if err != nil {
		t.Fatalf("SaveWorkspace failed: %v", err)
	}

	// Get workspace
	data, err := store.GetWorkspace(ctx, "ws-1")
	if err != nil {
		t.Fatalf("GetWorkspace failed: %v", err)
	}

	var retrieved map[string]interface{}
	err = json.Unmarshal(data, &retrieved)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if retrieved["name"] != "Test Workspace" {
		t.Errorf("Data mismatch: expected 'Test Workspace', got %v", retrieved["name"])
	}
}

func TestListWorkspaces(t *testing.T) {
	store, tmpDir := createTestStore(t)
	defer store.Close()
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// Save multiple workspaces
	for i := 1; i <= 3; i++ {
		id := "ws-" + string(rune(48+i))
		data := map[string]interface{}{"id": id}
		if err := store.SaveWorkspace(ctx, id, data); err != nil {
			t.Fatalf("SaveWorkspace failed: %v", err)
		}
	}

	// List workspaces
	workspaces, err := store.ListWorkspaces(ctx)
	if err != nil {
		t.Fatalf("ListWorkspaces failed: %v", err)
	}

	if len(workspaces) != 3 {
		t.Errorf("Expected 3 workspaces, got %d", len(workspaces))
	}
}

func TestDeleteWorkspace(t *testing.T) {
	store, tmpDir := createTestStore(t)
	defer store.Close()
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	testData := map[string]interface{}{"id": "ws-1"}

	// Save workspace
	if err := store.SaveWorkspace(ctx, "ws-1", testData); err != nil {
		t.Fatalf("SaveWorkspace failed: %v", err)
	}

	// Delete workspace
	if err := store.DeleteWorkspace(ctx, "ws-1"); err != nil {
		t.Fatalf("DeleteWorkspace failed: %v", err)
	}

	// Verify it's deleted
	_, err := store.GetWorkspace(ctx, "ws-1")
	if err == nil {
		t.Error("Expected error when getting deleted workspace")
	}
}

func TestSaveAndGetDocument(t *testing.T) {
	store, tmpDir := createTestStore(t)
	defer store.Close()
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	content := []byte("Hello, World!")

	// Save document
	if err := store.SaveDocument(ctx, "ws-1", "doc-1", content); err != nil {
		t.Fatalf("SaveDocument failed: %v", err)
	}

	// Get document
	data, err := store.GetDocument(ctx, "ws-1", "doc-1")
	if err != nil {
		t.Fatalf("GetDocument failed: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("Document content mismatch")
	}
}

func TestSaveAndGetMetadata(t *testing.T) {
	store, tmpDir := createTestStore(t)
	defer store.Close()
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	metadata := map[string]string{"version": "1.0"}

	// Save metadata
	if err := store.SaveMetadata(ctx, "meta-1", metadata); err != nil {
		t.Fatalf("SaveMetadata failed: %v", err)
	}

	// Get metadata
	data, err := store.GetMetadata(ctx, "meta-1")
	if err != nil {
		t.Fatalf("GetMetadata failed: %v", err)
	}

	var retrieved map[string]string
	err = json.Unmarshal(data, &retrieved)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if retrieved["version"] != "1.0" {
		t.Errorf("Metadata mismatch")
	}
}

func TestStats(t *testing.T) {
	store, tmpDir := createTestStore(t)
	defer store.Close()
	defer os.RemoveAll(tmpDir)

	stats := store.Stats()
	if stats == nil {
		t.Fatal("Stats returned nil")
	}
	if _, ok := stats["buckets"]; !ok {
		t.Error("Stats missing 'buckets' field")
	}
}

func BenchmarkSaveWorkspace(b *testing.B) {
	store, tmpDir := createTestStore(&testing.T{})
	defer store.Close()
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	data := map[string]interface{}{"test": "data"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.SaveWorkspace(ctx, "ws-bench", data)
	}
}
