package db

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestKVStore_FullCycle(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "vectora-kv-*")
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := NewKVStoreAtPath(dbPath)
	if err != nil {
		t.Fatalf("falha ao abrir db teste: %v", err)
	}
	defer store.db.Close()

	ctx := context.Background()
	bucket := "test_bucket"
	key := "hello"
	val := []byte("world")

	// 1. Set
	if err := store.Set(ctx, bucket, key, val); err != nil {
		t.Errorf("Set falhou: %v", err)
	}

	// 2. Get
	got, err := store.Get(ctx, bucket, key)
	if err != nil {
		t.Errorf("Get falhou: %v", err)
	}
	if string(got) != "world" {
		t.Errorf("Esperado 'world', obtido '%s'", string(got))
	}

	// 3. List
	keys, _ := store.List(ctx, bucket, "h")
	if len(keys) != 1 || keys[0] != "hello" {
		t.Errorf("List falhou, chaves: %v", keys)
	}

	// 4. Delete
	store.Delete(ctx, bucket, key)
	got, _ = store.Get(ctx, bucket, key)
	if got != nil {
		t.Error("Key deveria ter sido deletada")
	}
}

func TestVectorStore_Cycle(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "vectora-chroma-*")
	defer os.RemoveAll(tmpDir)

	store, err := NewVectorStoreAtPath(tmpDir)
	if err != nil {
		t.Fatalf("falha chromem test: %v", err)
	}

	ctx := context.Background()
	col := "test_col"
	chunk := Chunk{
		ID:      "pts_1",
		Content: "Vectora is an AI",
		Vector:  []float32{1.0, 0.0, 0.5},
	}

	// 1. Upsert
	if err := store.UpsertChunk(ctx, col, chunk); err != nil {
		t.Errorf("Upsert falhou: %v", err)
	}

	// 2. Query
	res, err := store.Query(ctx, col, []float32{1.0, 0.0, 0.4}, 1)
	if err != nil {
		t.Errorf("Query falhou: %v", err)
	}
	if len(res) == 0 || res[0].ID != "pts_1" {
		t.Error("Nao encontrou o chunk inserido")
	}
}
