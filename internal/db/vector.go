package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	vecos "github.com/Kaffyn/vectora/internal/os"
	"github.com/philippgille/chromem-go"
)

type ChromemStore struct {
	db *chromem.DB
}

// Instancia persistência baseada em pastas para Embeddings. 
// Isso atende à Iron Rule 3: Nenhum workspace pode ler conteúdo do outro.
func NewVectorStore() (*ChromemStore, error) {
	osMgr, err := vecos.NewManager()
	if err != nil {
		return nil, err
	}
	baseDir, _ := osMgr.GetAppDataDir()
	path := filepath.Join(baseDir, "data", "chroma")

	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}

	db, err := chromem.NewPersistentDB(path, false)
	if err != nil {
		return nil, fmt.Errorf("vectora_internal_err: falha silenciosa no binding do chromem persistent db: %w", err)
	}

	return &ChromemStore{db: db}, nil
}

func (s *ChromemStore) UpsertChunk(ctx context.Context, collection string, chunk Chunk) error {
	c, err := s.db.GetOrCreateCollection(collection, nil, nil)
	if err != nil {
		return err
	}

	doc := chromem.Document{
		ID:        chunk.ID,
		Content:   chunk.Content,
		Metadata:  chunk.Metadata,
		Embedding: chunk.Vector,
	}

	return c.AddDocument(ctx, doc)
}

func (s *ChromemStore) Query(ctx context.Context, collection string, queryVector []float32, topK int) ([]ScoredChunk, error) {
	c, err := s.db.GetOrCreateCollection(collection, nil, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.QueryEmbedding(ctx, queryVector, topK, nil, nil)
	if err != nil {
		return nil, err
	}

	var results []ScoredChunk
	for _, doc := range res {
		results = append(results, ScoredChunk{
			Chunk: Chunk{
				ID:       doc.ID,
				Content:  doc.Content,
				Metadata: doc.Metadata,
				Vector:   doc.Embedding,
			},
			Score: doc.Similarity,
		})
	}
	
	return results, nil
}

func (s *ChromemStore) DeleteCollection(ctx context.Context, collection string) error {
	return s.db.DeleteCollection(collection)
}

func (s *ChromemStore) CollectionExists(ctx context.Context, collection string) bool {
	c := s.db.GetCollection(collection, nil)
	return c != nil
}
