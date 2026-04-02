package bbolt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Kaffyn/Vectora/src/core/domain"
	bolt "go.etcd.io/bbolt"
)

var (
	BucketConversations = []byte("conversations")
)

type BboltConversationRepo struct {
	db *bolt.DB
}

func NewConversationRepo(db *bolt.DB) (*BboltConversationRepo, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(BucketConversations)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize buckets: %w", err)
	}
	return &BboltConversationRepo{db: db}, nil
}

func (r *BboltConversationRepo) Save(ctx context.Context, conv *domain.Conversation) error {
	conv.UpdatedAt = time.Now()
	data, err := json.Marshal(conv)
	if err != nil {
		return fmt.Errorf("failed to marshal conversation: %w", err)
	}

	return r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketConversations)
		return b.Put([]byte(conv.ID), data)
	})
}

func (r *BboltConversationRepo) GetByID(ctx context.Context, id string) (*domain.Conversation, error) {
	var conv domain.Conversation
	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketConversations)
		data := b.Get([]byte(id))
		if data == nil {
			return fmt.Errorf("conversation %s not found", id)
		}
		return json.Unmarshal(data, &conv)
	})
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *BboltConversationRepo) List(ctx context.Context) ([]*domain.Conversation, error) {
	var conversations []*domain.Conversation
	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketConversations)
		return b.ForEach(func(k, v []byte) error {
			var conv domain.Conversation
			if err := json.Unmarshal(v, &conv); err != nil {
				return err
			}
			conversations = append(conversations, &conv)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return conversations, nil
}

func (r *BboltConversationRepo) Delete(ctx context.Context, id string) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketConversations)
		return b.Delete([]byte(id))
	})
}
