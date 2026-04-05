package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.etcd.io/bbolt"
)

// Store manages data persistence using bbolt
type Store struct {
	db     *bbolt.DB
	logger *slog.Logger
	mu     sync.RWMutex
}

// StoreConfig contains store configuration
type StoreConfig struct {
	DatabasePath string
	Logger       *slog.Logger
}

// NewStore creates a new database store
func NewStore(dbPath string, logger *slog.Logger) (*Store, error) {
	if logger == nil {
		logger = slog.Default()
	}

	// Ensure directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database
	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &Store{
		db:     db,
		logger: logger,
	}

	// Initialize buckets
	if err := store.initBuckets(); err != nil {
		db.Close()
		return nil, err
	}

	logger.Info("Database store initialized", "path", dbPath)
	return store, nil
}

// initBuckets creates required buckets if they don't exist
func (s *Store) initBuckets() error {
	buckets := []string{
		"workspaces",
		"sessions",
		"documents",
		"embeddings",
		"metadata",
		"settings",
		"snapshots",
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		for _, bucketName := range buckets {
			_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
			if err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", bucketName, err)
			}
		}
		return nil
	})
}

// SaveWorkspace saves workspace data
func (s *Store) SaveWorkspace(ctx context.Context, id string, data interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal workspace data: %w", err)
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("workspaces"))
		return b.Put([]byte(id), jsonData)
	})
}

// GetWorkspace retrieves workspace data
func (s *Store) GetWorkspace(ctx context.Context, id string) (json.RawMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var data json.RawMessage
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("workspaces"))
		v := b.Get([]byte(id))
		if v == nil {
			return fmt.Errorf("workspace not found: %s", id)
		}
		data = make([]byte, len(v))
		copy(data, v)
		return nil
	})
	return data, err
}

// ListWorkspaces returns all workspace IDs
func (s *Store) ListWorkspaces(ctx context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var workspaces []string
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("workspaces"))
		return b.ForEach(func(k, v []byte) error {
			workspaces = append(workspaces, string(k))
			return nil
		})
	})
	return workspaces, err
}

// DeleteWorkspace deletes a workspace
func (s *Store) DeleteWorkspace(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("workspaces"))
		return b.Delete([]byte(id))
	})
}

// SaveDocument saves a document
func (s *Store) SaveDocument(ctx context.Context, workspaceID, docID string, content []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("documents"))
		key := fmt.Sprintf("%s:%s", workspaceID, docID)
		return b.Put([]byte(key), content)
	})
}

// GetDocument retrieves a document
func (s *Store) GetDocument(ctx context.Context, workspaceID, docID string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var data []byte
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("documents"))
		key := fmt.Sprintf("%s:%s", workspaceID, docID)
		v := b.Get([]byte(key))
		if v == nil {
			return fmt.Errorf("document not found: %s", docID)
		}
		data = make([]byte, len(v))
		copy(data, v)
		return nil
	})
	return data, err
}

// SaveMetadata saves metadata key-value pairs
func (s *Store) SaveMetadata(ctx context.Context, key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("metadata"))
		return b.Put([]byte(key), jsonData)
	})
}

// GetMetadata retrieves metadata
func (s *Store) GetMetadata(ctx context.Context, key string) (json.RawMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var data json.RawMessage
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("metadata"))
		v := b.Get([]byte(key))
		if v == nil {
			return fmt.Errorf("metadata not found: %s", key)
		}
		data = make([]byte, len(v))
		copy(data, v)
		return nil
	})
	return data, err
}

// Close closes the database
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Stats returns database statistics
func (s *Store) Stats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]interface{})
	s.db.View(func(tx *bbolt.Tx) error {
		txStats := tx.Stats()
		stats["page_count"] = txStats.PageCount
		return nil
	})
	return stats
}
