package db

import (
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

// NewBboltDB abre (ou cria) o arquivo bbolt.
func NewBboltDB(dbPath string) (*bolt.DB, error) {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("could not open bbolt db: %w", err)
	}
	return db, nil
}
