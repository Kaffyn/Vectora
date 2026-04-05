package db

// Chunk represents a data chunk for indexing
type Chunk struct {
	ID        string
	Content   string
	Vector    []float32 // Embedding vector
	Embedding []float32 // Alias for compatibility
	Metadata  map[string]string
	SourceFile string
	ChunkIndex int
}

// ScoredChunk represents a chunk with a similarity score
type ScoredChunk struct {
	Chunk *Chunk
	Score float32
}

// KVStore defines key-value store operations
type KVStore interface {
	Get(ctx interface{}, key string) (interface{}, error)
	Set(ctx interface{}, key string, value interface{}) error
	Delete(ctx interface{}, key string) error
}
