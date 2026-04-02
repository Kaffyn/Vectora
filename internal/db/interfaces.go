package db

import (
	"context"
)

// Chunk representa um fragmento de texto enriquecido e indexável com seus metadados.
type Chunk struct {
	ID       string
	Content  string
	Metadata map[string]string
	Vector   []float32
}

// ScoredChunk estende um Chunk anexando a Similaridade baseada em busca Cosmos (KNN).
type ScoredChunk struct {
	Chunk
	Score float32
}

// VectorStore provê abstração do Banco de Vetores (Chromem-go).
type VectorStore interface {
	UpsertChunk(ctx context.Context, collection string, chunk Chunk) error
	Query(ctx context.Context, collection string, queryVector []float32, topK int) ([]ScoredChunk, error)
	DeleteCollection(ctx context.Context, collection string) error
	CollectionExists(ctx context.Context, collection string) bool
}

// KVStore provê abstração do Banco Analítico (BBolt).
type KVStore interface {
	Set(ctx context.Context, bucket string, key string, value []byte) error
	Get(ctx context.Context, bucket string, key string) ([]byte, error)
	Delete(ctx context.Context, bucket string, key string) error
	List(ctx context.Context, bucket string, prefix string) ([]string, error)
}
