package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.etcd.io/bbolt"
)

const (
	BucketConfig      = "config"
	BucketFileIndex   = "file_index"
	BucketChatHistory = "chat_history"
	KeyWorkspaceState = "workspace_state"
)

type BBoltStore struct {
	db *bbolt.DB
	mu sync.RWMutex // Proteção extra para leituras concorrentes em memória se necessário
}

func NewBBoltStore(dbPath string) (*BBoltStore, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		return nil, fmt.Errorf("failed to create db dir: %w", err)
	}

	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{
		Timeout:      1 * time.Second, // Fail fast se estiver locked
		NoGrowSync:   false,
		FreelistType: bbolt.FreelistArrayType, // Melhor performance para writes sequenciais
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open bbolt: %w", err)
	}

	// Inicialização Idempotente dos Buckets
	err = db.Update(func(tx *bbolt.Tx) error {
		buckets := []string{BucketConfig, BucketFileIndex, BucketChatHistory}
		for _, b := range buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(b)); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		db.Close()
		return nil, err
	}

	return &BBoltStore{db: db}, nil
}

// SaveWorkspaceMeta persiste o estado global do workspace.
func (s *BBoltStore) SaveWorkspaceMeta(meta WorkspaceMeta) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketConfig))
		return b.Put([]byte(KeyWorkspaceState), data)
	})
}

// GetWorkspaceMeta recupera o estado. Retorna erro se não existir.
func (s *BBoltStore) GetWorkspaceMeta() (*WorkspaceMeta, error) {
	var meta WorkspaceMeta
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketConfig))
		data := b.Get([]byte(KeyWorkspaceState))
		if data == nil {
			return fmt.Errorf("workspace state not found")
		}
		return json.Unmarshal(data, &meta)
	})
	return &meta, err
}

// UpsertFileIndex atualiza ou cria o índice de um arquivo.
func (s *BBoltStore) UpsertFileIndex(relPath string, entry FileIndexEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketFileIndex))
		return b.Put([]byte(relPath), data)
	})
}

// GetFileIndex recupera o índice de um arquivo específico.
func (s *BBoltStore) GetFileIndex(relPath string) (*FileIndexEntry, error) {
	var entry FileIndexEntry
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketFileIndex))
		data := b.Get([]byte(relPath))
		if data == nil {
			return fmt.Errorf("index not found for %s", relPath)
		}
		return json.Unmarshal(data, &entry)
	})
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// DeleteFileIndex remove o índice de um arquivo (usado na re-indexação).
func (s *BBoltStore) DeleteFileIndex(relPath string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketFileIndex))
		return b.Delete([]byte(relPath))
	})
}

// Close fecha o banco de dados de forma segura.
func (s *BBoltStore) Close() error {
	return s.db.Close()
}
