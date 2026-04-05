package core

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Chunk struct {
	ID         string            `json:"id"`
	Content    string            `json:"content"`
	SourceFile string            `json:"source_file"`
	ChunkIndex int               `json:"chunk_index"`
	Metadata   map[string]string `json:"metadata"`
}

type QueryResult struct {
	ID            string    `json:"id"`
	Answer        string    `json:"answer"`
	Sources       []*Chunk  `json:"sources"`
	Thinking      string    `json:"thinking,omitempty"`
	ExecutionTime int64     `json:"execution_time_ms"`
	Model         string    `json:"model"`
}

type RAGEngine struct {
	wsManager *WorkspaceManager
	chunks    map[string][]*Chunk // wsID -> chunks
}

func NewRAGEngine(wsManager *WorkspaceManager) *RAGEngine {
	return &RAGEngine{
		wsManager: wsManager,
		chunks:    make(map[string][]*Chunk),
	}
}

func (r *RAGEngine) IndexChunk(ctx context.Context, wsID string, chunk *Chunk) error {
	if _, err := r.wsManager.Get(ctx, wsID); err != nil {
		return err
	}

	if r.chunks[wsID] == nil {
		r.chunks[wsID] = make([]*Chunk, 0)
	}

	r.chunks[wsID] = append(r.chunks[wsID], chunk)
	return nil
}

func (r *RAGEngine) Query(ctx context.Context, wsID, query string) (*QueryResult, error) {
	ws, err := r.wsManager.Get(ctx, wsID)
	if err != nil {
		return nil, err
	}

	// Get chunks for this workspace
	chunks := r.chunks[wsID]
	if len(chunks) == 0 {
		return &QueryResult{
			ID:     fmt.Sprintf("q_%d", time.Now().Unix()),
			Answer: "Workspace não possui chunks indexados",
			Model:  "local",
		}, nil
	}

	// Simple relevance search (TOP-K = 5)
	relevant := r.searchRelevant(query, chunks, 5)

	// Build context
	contextStr := r.buildContext(relevant)

	// Simula resposta LLM
	answer := fmt.Sprintf("Baseado nos seus documentos:\n\n%s\n\nResposta gerada.", contextStr[:100])

	return &QueryResult{
		ID:            fmt.Sprintf("q_%d", time.Now().Unix()),
		Answer:        answer,
		Sources:       relevant,
		ExecutionTime: 150,
		Model:         ws.Name,
	}, nil
}

func (r *RAGEngine) searchRelevant(query string, chunks []*Chunk, topK int) []*Chunk {
	// Simple text-based relevance (sem embedding por enquanto)
	queryWords := strings.Fields(strings.ToLower(query))

	scored := make([]struct {
		chunk *Chunk
		score int
	}, 0)

	for _, chunk := range chunks {
		score := 0
		contentLower := strings.ToLower(chunk.Content)
		for _, word := range queryWords {
			if strings.Contains(contentLower, word) {
				score++
			}
		}
		if score > 0 {
			scored = append(scored, struct {
				chunk *Chunk
				score int
			}{chunk, score})
		}
	}

	// Sort by score (simples)
	result := make([]*Chunk, 0, topK)
	for _, s := range scored {
		if len(result) < topK {
			result = append(result, s.chunk)
		}
	}

	return result
}

func (r *RAGEngine) buildContext(chunks []*Chunk) string {
	var sb strings.Builder
	for i, chunk := range chunks {
		sb.WriteString(fmt.Sprintf("Documento %d:\n%s\n\n", i+1, chunk.Content))
	}
	return sb.String()
}
