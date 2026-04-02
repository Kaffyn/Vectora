package llm

import (
	"context"
	"errors"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

type GeminiProvider struct {
	client *googleai.GoogleAI
}

func NewGeminiProvider(ctx context.Context, apiKey string) (*GeminiProvider, error) {
	if apiKey == "" {
		return nil, errors.New("gemini_api_key_required: a chave Gemini API não foi informada")
	}

	// Langchaingo's GoogleAI costuma extrair de fallback vars, injetamos explicitamente em caso de falha silenciosa.
	os.Setenv("GEMINI_API_KEY", apiKey)

	client, err := googleai.New(ctx, googleai.WithAPIKey(apiKey), googleai.WithDefaultModel("gemini-1.5-flash"))
	if err != nil {
		return nil, err
	}
	
	return &GeminiProvider{client: client}, nil
}

func (p *GeminiProvider) Name() string {
	return "gemini"
}

func (p *GeminiProvider) IsConfigured() bool {
	return p.client != nil
}

func (p *GeminiProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	var content []llms.MessageContent
	
	if req.SystemPrompt != "" {
		content = append(content, llms.TextParts(llms.ChatMessageTypeSystem, req.SystemPrompt))
	}

	for _, msg := range req.Messages {
		var role llms.ChatMessageType
		switch msg.Role {
		case RoleUser:
			role = llms.ChatMessageTypeHuman
		case RoleAssistant:
			role = llms.ChatMessageTypeAI
		case RoleSystem:
			role = llms.ChatMessageTypeSystem
		case RoleTool:
			role = llms.ChatMessageTypeTool
		// Fallback amigável
		default:
			role = llms.ChatMessageTypeHuman 
		}
		content = append(content, llms.TextParts(role, msg.Content))
	}

	opts := []llms.CallOption{
		llms.WithMaxTokens(req.MaxTokens),
		llms.WithTemperature(float64(req.Temperature)),
	}

	// TODO: Fazer parse de Tools (req.Tools) em llms.Tool conforme a interface langchaingo
	
	resp, err := p.client.GenerateContent(ctx, content, opts...)
	if err != nil {
		return CompletionResponse{}, err
	}

	// Mapeia Resposta
	if len(resp.Choices) == 0 {
		return CompletionResponse{}, errors.New("llm_empty_response: não houveram choices enviadas")
	}

	choice := resp.Choices[0]
	
	// TODO: Mapear ToolCalls geradas (choice.ToolCalls) na response
	
	return CompletionResponse{
		Content: choice.Content,
		Usage: TokenUsage{
			// TODO: googleai metrics se fornecidas
		},
	}, nil
}

func (p *GeminiProvider) Embed(ctx context.Context, input string) ([]float32, error) {
	// A camada de Embedding será baseada no `text-embedding-004` (Gemini embeddings default).
	embClient, err := p.client.CreateEmbedding(ctx, []string{input})
	if err != nil {
		return nil, err
	}
	if len(embClient) > 0 {
		return embClient[0], nil
	}
	return nil, errors.New("gemini_embedding_failed: sem vetores retornados")
}
