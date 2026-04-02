package bbolt

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Kaffyn/Vectora/src/core/domain"
	bolt "go.etcd.io/bbolt"
)

var (
	BucketChunks = []byte("chunks")
)

type BboltChunkRepo struct {
	db *bolt.DB
}

func NewChunkRepo(db *bolt.DB) (*BboltChunkRepo, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(BucketChunks)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize chunks bucket: %w", err)
	}
	return &BboltChunkRepo{db: db}, nil
}

func (r *BboltChunkRepo) Save(ctx context.Context, chunk *domain.Chunk, embedding []float32) error {
	data, err := json.Marshal(chunk)
	if err != nil {
		return fmt.Errorf("failed to marshal chunk: %w", err)
	}

	return r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketChunks)
		return b.Put([]byte(chunk.ID), data)
	})
}

func (r *BboltChunkRepo) Search(ctx context.Context, query string, limit int) ([]*domain.Chunk, error) {
	var chunks []*domain.Chunk
	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketChunks)
		c := b.Cursor()

		for k, v := c.First(); k != nil && len(chunks) < limit; k, v = c.Next() {
			var chunk domain.Chunk
			if err := json.Unmarshal(v, &chunk); err != nil {
				continue
			}
			if strings.Contains(strings.ToLower(chunk.Content), strings.ToLower(query)) {
				chunks = append(chunks, &chunk)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return chunks, nil
}

func (r *BboltChunkRepo) VectorSearch(ctx context.Context, embedding []float32, limit int) ([]*domain.Chunk, error) {
	// Delegado para chromem-go no orquestrador principal
	return nil, nil
}

func (r *BboltChunkRepo) GetByDocumentID(ctx context.Context, docID string) ([]*domain.Chunk, error) {
	var chunks []*domain.Chunk
	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketChunks)
		return b.ForEach(func(k, v []byte) error {
			var chunk domain.Chunk
			if err := json.Unmarshal(v, &chunk); err != nil {
				return err
			}
			if chunk.DocumentID == docID {
				chunks = append(chunks, &chunk)
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return chunks, nil
}
