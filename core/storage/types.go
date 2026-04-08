package storage

import (
	"time"
)

// WorkspaceMeta armazena o estado de alto nível do projeto indexado.
// Serializado no bucket 'config' sob a chave 'workspace_state'.
type WorkspaceMeta struct {
	Path          string    `json:"path"`
	LastIndexedAt time.Time `json:"last_indexed_at"`
	ModelProvider string    `json:"model_provider"` // ex: "gemini-embedding-2"
	TotalFiles    int       `json:"total_files"`
	TotalChunks   int       `json:"total_chunks"`
	Status        string    `json:"status"` // "idle", "indexing", "error"
}

// FileIndexEntry mapeia um arquivo físico para seus vetores no Chromem.
// Chave no bucket 'file_index': path relativo do arquivo.
type FileIndexEntry struct {
	AbsolutePath string    `json:"absolute_path"`
	ContentHash  string    `json:"content_hash"` // SHA-256
	ChunkIDs     []string  `json:"chunk_ids"`    // IDs dos documentos no Chromem
	UpdatedAt    time.Time `json:"updated_at"`
	SizeBytes    int64     `json:"size_bytes"`
}

// ChatMessage representa uma interação atômica.
type ChatMessage struct {
	Role      string                 `json:"role"`    // "user", "assistant", "system"
	Content   string                 `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"` // Ex: tool_calls
}

// ChunkConfig define as regras de quebra de texto.
type ChunkConfig struct {
	MaxTokens int `json:"max_tokens"`
	Overlap   int `json:"overlap"`
}

var DefaultChunkConfig = ChunkConfig{
	MaxTokens: 800, // Conservador para embeddings precisos
	Overlap:   100, // "Costura quente" para contexto
}
