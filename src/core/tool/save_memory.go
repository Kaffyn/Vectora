package tool

import (
	"context"
	"fmt"
	"time"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

// SaveMemoryTool implementa a interface Tool para armazenamento de memória semântica.
type SaveMemoryTool struct {
	repo     domain.ChunkRepository
	embedder domain.EmbeddingProvider
}

func NewSaveMemoryTool(repo domain.ChunkRepository, embedder domain.EmbeddingProvider) *SaveMemoryTool {
	return &SaveMemoryTool{repo: repo, embedder: embedder}
}

func (t *SaveMemoryTool) Name() string {
	return "save_memory"
}

func (t *SaveMemoryTool) Description() string {
	return "Salva um fato, decisão ou pedaço de conhecimento importante na memória técnica do Vectora (RAG). Use para que a IA lembre de decisões de projeto futuras ou especificações chave."
}

func (t *SaveMemoryTool) Type() domain.ACPActionType {
	return domain.ACPActionWrite
}

func (t *SaveMemoryTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"knowledge": map[string]interface{}{
				"type":        "string",
				"description": "O conhecimento ou fato técnico a ser memorizado.",
			},
			"tag": map[string]interface{}{
				"type":        "string",
				"description": "Uma etiqueta para categorizar a memória (ex: 'arquitetura', 'db-schema').",
			},
		},
		"required": []string{"knowledge"},
	}
}

func (t *SaveMemoryTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	knowledge, ok := args["knowledge"].(string)
	if !ok {
		return nil, fmt.Errorf("argumento 'knowledge' é obrigatório")
	}

	tag := "general"
	if t, ok := args["tag"].(string); ok {
		tag = t
	}

	// 1. Gerar Embedding
	emb, err := t.embedder.EmbedQuery(ctx, knowledge)
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar embedding: %w", err)
	}

	// 2. Criar Chunk de Memória
	chunk := &domain.Chunk{
		ID:         fmt.Sprintf("mem-%d", time.Now().UnixNano()),
		DocumentID: "internal-memory",
		Content:    knowledge,
		Metadata:   map[string]string{"tag": tag, "type": "memory"},
	}

	// 3. Salvar no Repositório Vetorial (chromem/lancedb)
	if err := t.repo.Save(ctx, chunk, emb); err != nil {
		return nil, fmt.Errorf("falha ao salvar memória: %w", err)
	}

	return map[string]string{
		"id":     chunk.ID,
		"status": "memorized",
	}, nil
}
