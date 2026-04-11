package db

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
)

// Chunk represents an enriched and indexable text fragment with its metadata.
type Chunk struct {
	ID       string
	Content  string
	Metadata map[string]string
	Vector   []float32
}

// ScoredChunk extends a Chunk by attaching Similarity based on Cosine search (KNN).
type ScoredChunk struct {
	Chunk
	Score float32
}

// FileIndexEntry mapeia um arquivo físico para seus vetores no Chromem.
type FileIndexEntry struct {
	AbsolutePath string   `json:"absolute_path"`
	ContentHash  string   `json:"content_hash"`
	ChunkIDs     []string `json:"chunk_ids"`
	SizeBytes    int64    `json:"size_bytes"`
}

// CalculateHash gera SHA-256 de um conteúdo.
func CalculateHash(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))
}

// VectorStore provides an abstraction for the Vector Database (Chromem-go).
type VectorStore interface {
	UpsertChunk(ctx context.Context, collection string, chunk Chunk) error
	Query(ctx context.Context, collection string, queryVector []float32, topK int) ([]ScoredChunk, error)
	DeleteCollection(ctx context.Context, collection string) error
	CollectionExists(ctx context.Context, collection string) bool
	Close() error
}

// KVStore provides an abstraction for the Analytical Database (BBolt).
type KVStore interface {
	Set(ctx context.Context, bucket string, key string, value []byte) error
	Get(ctx context.Context, bucket string, key string) ([]byte, error)
	Delete(ctx context.Context, bucket string, key string) error
	List(ctx context.Context, bucket string, prefix string) ([]string, error)
	Close() error
}
