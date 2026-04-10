package db

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	vecos "github.com/Kaffyn/Vectora/core/os"
	"go.etcd.io/bbolt"
)

type BBoltStore struct {
	db *bbolt.DB
}

// NewKVStore initializes the root bbolt KV store connection using OS-specific AppData.
func NewKVStore() (*BBoltStore, error) {
	osMgr, err := vecos.NewManager()
	if err != nil {
		return nil, err
	}
	baseDir, _ := osMgr.GetAppDataDir()
	dbPath := filepath.Join(baseDir, "data", "vectora.db")
	return NewKVStoreAtPath(dbPath)
}

func NewKVStoreAtPath(dbPath string) (*BBoltStore, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, err
	}
	options := &bbolt.Options{Timeout: 1 * time.Second}
	db, err := bbolt.Open(dbPath, 0600, options)
	if err != nil {
		return nil, err
	}
	return &BBoltStore{db: db}, nil
}

func (s *BBoltStore) Set(ctx context.Context, bucket string, key string, value []byte) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		return b.Put([]byte(key), value)
	})
}

func (s *BBoltStore) Get(ctx context.Context, bucket string, key string) ([]byte, error) {
	var val []byte
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bbolt_bucket_empty: bucket '%s' not found", bucket)
		}
		v := b.Get([]byte(key))
		if v != nil {
			val = make([]byte, len(v))
			copy(val, v)
		}
		return nil
	})
	return val, err
}

func (s *BBoltStore) Delete(ctx context.Context, bucket string, key string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		return b.Delete([]byte(key))
	})
}

func (s *BBoltStore) List(ctx context.Context, bucket string, prefix string) ([]string, error) {
	var keys []string
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		prefixBytes := []byte(prefix)
		for k, _ := c.Seek(prefixBytes); k != nil && bytes.HasPrefix(k, prefixBytes); k, _ = c.Next() {
			keys = append(keys, string(k))
		}
		return nil
	})
	return keys, err
}

func (s *BBoltStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
