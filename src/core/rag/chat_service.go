package rag

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

type ChatService struct {
	convRepo  domain.ConversationRepository
	osManager domain.OSManager
	embedder  domain.EmbeddingProvider
	llm       domain.LLMProvider
	searchSvc *SearchService
}

func NewChatService(convRepo domain.ConversationRepository, osManager domain.OSManager, embedder domain.EmbeddingProvider, llm domain.LLMProvider, searchSvc *SearchService) *ChatService {
	return &ChatService{
		convRepo:  convRepo,
		osManager: osManager,
		embedder:  embedder,
		llm:       llm,
		searchSvc: searchSvc,
	}
}

func (s *ChatService) SendMessage(ctx context.Context, convID string, message string) (string, error) {
	// 1. Vetorizar a Query
	emb, err := s.embedder.EmbedQuery(ctx, message)
	if err != nil {
		return "", fmt.Errorf("falha ao vetorizar query: %w", err)
	}

	// 2. Busca RAG
	contextChunks, err := s.searchSvc.VectorSearch(ctx, emb, 5)
	if err != nil {
		fmt.Printf("Aviso: RAG Search falhou: %v\n", err)
	}

	var contextBuilder strings.Builder
	for _, chunk := range contextChunks {
		contextBuilder.WriteString(fmt.Sprintf("\n--- Fonte: %s ---\n%s\n", chunk.ID, chunk.Content))
	}

	// 3. Gerar resposta via LLM
	response, err := s.llm.GenerateResponse(ctx, message, contextBuilder.String())
	if err != nil {
		return "", fmt.Errorf("falha ao gerar resposta: %w", err)
	}

	// 4. Salvar no histórico centralizado
	if convID != "" {
		_ = s.convRepo.Save(ctx, &domain.Conversation{
			ID:        convID,
			UpdatedAt: time.Now(),
		})
	}

	return response, nil
}
