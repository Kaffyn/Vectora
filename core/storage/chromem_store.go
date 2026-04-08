package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/philippgille/chromem-go"
)

type ChromemStore struct {
	db             *chromem.DB
	collections    map[string]*chromem.Collection
	defaultEmbedFn chromem.EmbeddingFunc
}

func NewChromemStore(dataDir string) (*ChromemStore, error) {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, err
	}

	// PersistentDB salva automaticamente em disco a cada alteração
	db, err := chromem.NewPersistentDB(dataDir, false) // false = sem compressão snappy (mais rápido CPU)
	if err != nil {
		return nil, fmt.Errorf("failed to init chromem: %w", err)
	}

	return &ChromemStore{
		db:          db,
		collections: make(map[string]*chromem.Collection),
	}, nil
}

// GetOrCreateCollection garante que a coleção do workspace exista.
// Usamos o ID do workspace como nome da coleção para isolamento total.
func (s *ChromemStore) GetOrCreateCollection(workspaceID string) (*chromem.Collection, error) {
	// Verifica cache local
	if col, ok := s.collections[workspaceID]; ok {
		return col, nil
	}

	// Tenta obter coleção existente
	col := s.db.GetCollection(workspaceID, nil)
	if col != nil {
		s.collections[workspaceID] = col
		return col, nil
	}

	// Cria nova coleção (nil = embedding function padrão)
	col, err := s.db.CreateCollection(workspaceID, map[string]string{
		"type":    "codebase",
		"version": "1.0",
	}, nil)

	if err != nil {
		return nil, err
	}
	s.collections[workspaceID] = col
	return col, nil
}

// AddVectors insere novos chunks no índice vetorial.
func (s *ChromemStore) AddVectors(ctx context.Context, workspaceID string, docs []chromem.Document) error {
	col, err := s.GetOrCreateCollection(workspaceID)
	if err != nil {
		return err
	}

	// AddDocuments é atômico para o batch (0 = sem paralelismo)
	return col.AddDocuments(ctx, docs, 0)
}

// DeleteVectors remove vetores específicos por ID.
func (s *ChromemStore) DeleteVectors(ctx context.Context, workspaceID string, ids []string) error {
	_, err := s.GetOrCreateCollection(workspaceID)
	if err != nil {
		return err
	}
	// chromem-go não expõe DeleteDocuments diretamente, usamos Where filter
	// Para deletar, precisamos recriar sem os IDs
	// Workaround: marcamos como deletados no metadata
	for _, id := range ids {
		// Nota: chromem-go v0.7.0 não tem método direto de deleção
		// Implementação futura: reconstruir coleção sem os IDs
		_ = id // suppress unused variable
	}
	return fmt.Errorf("DeleteVectors não implementado no chromem-go v0.7.0")
}

// QueryVectors busca os vizinhos mais próximos de um embedding.
func (s *ChromemStore) QueryVectors(ctx context.Context, workspaceID string, embedding []float32, limit int) ([]chromem.Result, error) {
	col, err := s.GetOrCreateCollection(workspaceID)
	if err != nil {
		return nil, err
	}

	// QueryEmbedding espera o vetor já calculado
	results, err := col.QueryEmbedding(ctx, embedding, limit, nil, nil)
	if err != nil {
		return nil, err
	}
	return results, nil
}
