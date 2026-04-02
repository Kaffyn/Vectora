package domain

import "context"

type DocumentRepository interface {
	Save(ctx context.Context, doc *Document) error
	GetByID(ctx context.Context, id string) (*Document, error)
	GetByPath(ctx context.Context, path string) (*Document, error)
}

type ChunkRepository interface {
	Save(ctx context.Context, chunk *Chunk, embedding []float32) error
	Search(ctx context.Context, query string, limit int) ([]*Chunk, error)
	VectorSearch(ctx context.Context, embedding []float32, limit int) ([]*Chunk, error)
	GetByDocumentID(ctx context.Context, docID string) ([]*Chunk, error)
}

type ConversationRepository interface {
	Save(ctx context.Context, conv *Conversation) error
	GetByID(ctx context.Context, id string) (*Conversation, error)
	List(ctx context.Context) ([]*Conversation, error)
	Delete(ctx context.Context, id string) error
}
