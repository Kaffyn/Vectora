package rag

import (
	"context"
	"fmt"

	"github.com/Kaffyn/Vectora/src/core/ai"
	"github.com/Kaffyn/Vectora/src/core/domain"
)

type IngestService struct {
	docRepo   domain.DocumentRepository
	chunkRepo domain.ChunkRepository
	embedder  domain.EmbeddingProvider
	sm        *ai.SidecarManager
}

func NewIngestService(docRepo domain.DocumentRepository, chunkRepo domain.ChunkRepository, embedder domain.EmbeddingProvider, sm *ai.SidecarManager) *IngestService {
	return &IngestService{
		docRepo:   docRepo,
		chunkRepo: chunkRepo,
		embedder:  embedder,
		sm:        sm,
	}
}

func (s *IngestService) Ingest(ctx context.Context, doc *domain.Document) error {
	// Step 1: Save the Document
	if err := s.docRepo.Save(ctx, doc); err != nil {
		return fmt.Errorf("failed to save document: %w", err)
	}

	// Step 2: Create chunks (Simple chunking for now)
	if doc.Content != "" {
		// Mock chunking logic: split by double newline
		chunk := &domain.Chunk{
			ID:         doc.ID + "_c0",
			DocumentID: doc.ID,
			Content:    doc.Content,
			Index:      0,
		}

		// Garante que o sidecar de embedding está rodando (JIT)
		activeEmbeddingModel := s.sm.GetActiveModel("embedding")
		if activeEmbeddingModel == "" {
			activeEmbeddingModel = "qwen3-embedding-0.6b" // fallback seguro
		}
		if err := s.sm.EnsureRunning(ctx, "embedding", activeEmbeddingModel, 8082, false); err != nil {
			return fmt.Errorf("failed to wake up embedding sidecar: %w", err)
		}

		// Generate Embedding
		embedding, err := s.embedder.EmbedQuery(ctx, chunk.Content)
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}

		if err := s.chunkRepo.Save(ctx, chunk, embedding); err != nil {
			return fmt.Errorf("failed to save chunk: %w", err)
		}
	}

	return nil
}
