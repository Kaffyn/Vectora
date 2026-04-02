package chromem

import (
	"context"
	"fmt"
	"sync"

	"github.com/Kaffyn/Vectora/src/core/domain"
	"github.com/philippgille/chromem-go"
)

type VectorRepo struct {
	db          *chromem.DB
	collections map[string]*chromem.Collection
	mu          sync.RWMutex
}

func NewVectorRepo() *VectorRepo {
	return &VectorRepo{
		db:          chromem.NewDB(),
		collections: make(map[string]*chromem.Collection),
	}
}

// LoadIndex (Simplified for chromem-go native behavior)
func (r *VectorRepo) LoadIndex(ctx context.Context, id string, path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// chromem-go gerencia coleções em memória.
	// Para persistência real, usaríamos db.Save(folder) / chromem.NewDBFromFolder
	// Por enquanto, apenas registramos a coleção.
	coll := r.db.GetCollection(id, nil)
	if coll == nil {
		var err error
		coll, err = r.db.CreateCollection(id, nil, nil)
		if err != nil {
			return err
		}
	}

	r.collections[id] = coll
	return nil
}

// PersistIndex (Not natively available as method, placeholder or simplified)
func (r *VectorRepo) PersistIndex(ctx context.Context, id string, path string) error {
	return nil // chromem-go persistência é feita via DB.Save no nível global
}

func (r *VectorRepo) SaveToIndex(ctx context.Context, id string, chunk *domain.Chunk, embedding []float32) error {
	r.mu.Lock()
	coll, ok := r.collections[id]
	if !ok {
		var err error
		coll, err = r.db.CreateCollection(id, nil, nil)
		if err != nil {
			r.mu.Unlock()
			return err
		}
		r.collections[id] = coll
	}
	r.mu.Unlock()

	doc := chromem.Document{
		ID:        chunk.ID,
		Content:   chunk.Content,
		Embedding: embedding,
		Metadata: map[string]string{
			"document_id": chunk.DocumentID,
			"index":       fmt.Sprintf("%d", chunk.Index),
		},
	}

	return coll.AddDocument(ctx, doc)
}

func (r *VectorRepo) VectorSearch(ctx context.Context, embedding []float32, limit int) ([]*domain.Chunk, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var allChunks []*domain.Chunk
	for _, coll := range r.collections {
		res, err := coll.QueryEmbedding(ctx, embedding, limit, nil, nil)
		if err != nil {
			continue
		}

		for _, node := range res {
			allChunks = append(allChunks, &domain.Chunk{
				ID:         node.ID,
				Content:    node.Content,
				DocumentID: node.Metadata["document_id"],
			})
		}
	}

	if len(allChunks) > limit {
		allChunks = allChunks[:limit]
	}

	return allChunks, nil
}

func (r *VectorRepo) Search(ctx context.Context, query string, limit int) ([]*domain.Chunk, error) {
	// Nota: Em RAG real, primeiro fazemos embedding da query string
	return r.VectorSearch(ctx, nil, limit)
}

func (r *VectorRepo) GetByDocumentID(ctx context.Context, docID string) ([]*domain.Chunk, error) {
	return nil, nil // TODO: Implementar filtro por metadata
}

// Save implementa domain.ChunkRepository (Sempre usa o index padrão "default")
func (r *VectorRepo) Save(ctx context.Context, chunk *domain.Chunk, embedding []float32) error {
	return r.SaveToIndex(ctx, "default", chunk, embedding)
}

var _ domain.ChunkRepository = (*VectorRepo)(nil)
