package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"time"

	"github.com/philippgille/chromem-go"
)

type Engine struct {
	Meta        *BBoltStore
	Vec         *ChromemStore
	RootPath    string
	WorkspaceID string
}

// NewEngine inicializa o storage para um workspace específico.
// workspaceHash deve ser único para o path do projeto (ex: SHA256 do path absoluto).
func NewEngine(baseRoot string, workspaceHash string) (*Engine, error) {
	workspaceDir := filepath.Join(baseRoot, workspaceHash)

	metaStore, err := NewBBoltStore(filepath.Join(workspaceDir, "metadata.db"))
	if err != nil {
		return nil, fmt.Errorf("init bbolt: %w", err)
	}

	vecStore, err := NewChromemStore(filepath.Join(workspaceDir, "vector_db"))
	if err != nil {
		metaStore.Close()
		return nil, fmt.Errorf("init chromem: %w", err)
	}

	return &Engine{
		Meta:        metaStore,
		Vec:         vecStore,
		RootPath:    workspaceDir,
		WorkspaceID: workspaceHash,
	}, nil
}

// CalculateHash gera SHA-256 de um conteúdo.
func CalculateHash(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))
}

// IndexFile é a operação atômica de alta nível:
// 1. Calcula hash novo.
// 2. Compara com BBolt.
// 3. Se diferente, deleta vetores antigos no Chromem.
// 4. Chunka novo conteúdo.
// 5. Insere vetores novos no Chromem.
// 6. Atualiza BBolt.
func (e *Engine) IndexFile(ctx context.Context, relPath string, content string) error {
	newHash := CalculateHash(content)

	// 1. Verifica se precisa re-indexar
	existing, err := e.Meta.GetFileIndex(relPath)
	if err == nil && existing != nil && existing.ContentHash == newHash {
		return nil // Nada mudou, skip rápido
	}

	// 2. Limpeza de vetores antigos se existirem
	if existing != nil && len(existing.ChunkIDs) > 0 {
		if err := e.Vec.DeleteVectors(ctx, e.WorkspaceID, existing.ChunkIDs); err != nil {
			return fmt.Errorf("delete old vectors: %w", err)
		}
	}

	// 3. Chunking
	chunks := SimpleChunker(content, DefaultChunkConfig)
	if len(chunks) == 0 {
		return nil
	}

	// 4. Preparar documentos para Chromem
	docs := make([]chromem.Document, len(chunks))
	newChunkIDs := make([]string, len(chunks))

	for i, chunk := range chunks {
		// ID único: hash_do_arquivo:indice
		chunkID := fmt.Sprintf("%s:%d", newHash, i)
		newChunkIDs[i] = chunkID

		docs[i] = chromem.Document{
			ID:      chunkID,
			Content: chunk,
			Metadata: map[string]string{
				"source": relPath,
				"hash":   newHash,
				"index":  fmt.Sprintf("%d", i),
			},
		}
	}

	// 5. Inserir no Chromem
	if err := e.Vec.AddVectors(ctx, e.WorkspaceID, docs); err != nil {
		return fmt.Errorf("add new vectors: %w", err)
	}

	// 6. Atualizar Metadata no BBolt
	entry := FileIndexEntry{
		AbsolutePath: filepath.Join(e.RootPath, relPath), // Ou path original do projeto
		ContentHash:  newHash,
		ChunkIDs:     newChunkIDs,
		SizeBytes:    int64(len(content)),
		UpdatedAt:    time.Now(),
	}

	if err := e.Meta.UpsertFileIndex(relPath, entry); err != nil {
		// Rollback parcial: deveríamos deletar os vetores inseridos se falhar aqui
		e.Vec.DeleteVectors(ctx, e.WorkspaceID, newChunkIDs)
		return fmt.Errorf("update meta: %w", err)
	}

	return nil
}

func (e *Engine) Close() error {
	return e.Meta.Close()
}
