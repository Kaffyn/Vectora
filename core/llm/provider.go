package llm

import (
	"context"
	"io"
)

// Message representa uma unidade de conversa.
type Message struct {
	Role    string // "user", "model", "system"
	Content string
}

// ChatRequest encapsula os parâmetros de uma chamada de chat.
type ChatRequest struct {
	Messages        []Message
	Temperature     float32
	MaxOutputTokens int
	// Tools         []ToolDefinition // Schema das ferramentas disponíveis (a definir)
}

// ChatResponse streama a resposta do modelo.
type ChatResponse struct {
	Text  string
	Done  bool
	Error error
}

// EmbedRequest para geração de vetores.
type EmbedRequest struct {
	Text  string
	Model string // ex: "gemini-embedding-2"
}

// LLMProvider é a interface mestre.
type LLMProvider interface {
	// Chat envia mensagens e retorna um reader para streamar a resposta.
	Chat(ctx context.Context, req ChatRequest) (io.ReadCloser, error)

	// Embed gera o vetor para um texto específico.
	Embed(ctx context.Context, req EmbedRequest) ([]float32, error)

	// Name retorna o identificador do provider (ex: "gemini-pro")
	Name() string
}
