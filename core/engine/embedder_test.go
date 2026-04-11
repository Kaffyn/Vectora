package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/llm"
)

type mockKVStore struct {
	data map[string][]byte
}

func (m *mockKVStore) Set(ctx context.Context, bucket, key string, value []byte) error {
	m.data[bucket+":"+key] = value
	return nil
}

func (m *mockKVStore) Get(ctx context.Context, bucket, key string) ([]byte, error) {
	return m.data[bucket+":"+key], nil
}

func (m *mockKVStore) Delete(ctx context.Context, bucket, key string) error {
	delete(m.data, bucket+":"+key)
	return nil
}

func (m *mockKVStore) List(ctx context.Context, bucket, prefix string) ([]string, error) {
	return nil, nil
}

type mockVectorStore struct {
	chunks []db.Chunk
}

func (m *mockVectorStore) UpsertChunk(ctx context.Context, collection string, chunk db.Chunk) error {
	m.chunks = append(m.chunks, chunk)
	return nil
}

func (m *mockVectorStore) Query(ctx context.Context, collection string, vector []float32, topK int) ([]db.ScoredChunk, error) {
	return nil, nil
}

func (m *mockVectorStore) DeleteCollection(ctx context.Context, collection string) error {
	return nil
}

func (m *mockVectorStore) CollectionExists(ctx context.Context, collection string) bool {
	return true
}

type mockEmbedProvider struct{}

func (m *mockEmbedProvider) Name() string       { return "mock" }
func (m *mockEmbedProvider) IsConfigured() bool { return true }
func (m *mockEmbedProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	return []float32{1.0, 2.0, 3.0}, nil
}
func (m *mockEmbedProvider) Complete(ctx context.Context, req llm.CompletionRequest) (llm.CompletionResponse, error) {
	return llm.CompletionResponse{}, nil
}
func (m *mockEmbedProvider) StreamComplete(ctx context.Context, req llm.CompletionRequest) (<-chan llm.CompletionResponse, <-chan error) {
	return nil, nil
}

func TestRunEmbedJob(t *testing.T) {
	// Create a temp directory for testing
	tempDir, err := os.MkdirTemp("", "vectora-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a few test files
	file1 := filepath.Join(tempDir, "test1.go")
	os.WriteFile(file1, []byte("package test\nfunc main() {}"), 0644)

	file2 := filepath.Join(tempDir, "test2.md")
	os.WriteFile(file2, []byte("# Test Title\nThis is a test."), 0644)

	kv := &mockKVStore{data: make(map[string][]byte)}
	storage := &mockVectorStore{}
	provider := &mockEmbedProvider{}

	cfg := EmbedJobConfig{
		RootPath:       tempDir,
		Workspace:      "test_ws",
		CollectionName: "ws_test_ws",
		Force:          true,
	}

	progressCount := 0
	completeCalled := false

	RunEmbedJob(context.Background(), cfg, kv, storage, provider, func(p EmbedProgress) {
		progressCount++
		if p.IsComplete {
			completeCalled = true
		}
	})

	if !completeCalled {
		t.Error("EmbedJob did not call completion callback")
	}

	if len(storage.chunks) == 0 {
		t.Error("No chunks were upserted to vector store")
	}

	// Verify that at least some chunks were created
	foundGo := false
	foundMd := false
	for _, c := range storage.chunks {
		if c.Metadata["language"] == "go" {
			foundGo = true
		}
		if c.Metadata["language"] == "markdown" {
			foundMd = true
		}
	}

	if !foundGo {
		t.Error("Go file was not embedded correctly")
	}
	if !foundMd {
		t.Error("Markdown file was not embedded correctly")
	}
}
