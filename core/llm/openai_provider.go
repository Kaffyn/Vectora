package llm

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OpenAIProvider struct {
	client openai.Client
	name   string
}

func NewOpenAIProvider(apiKey string, baseURL string, name string) *OpenAIProvider {
	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}

	return &OpenAIProvider{
		client: openai.NewClient(opts...),
		name:   name,
	}
}

func (p *OpenAIProvider) Name() string {
	return p.name
}

func (p *OpenAIProvider) IsConfigured() bool {
	return true
}

func (p *OpenAIProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	messages := make([]openai.ChatCompletionMessageParamUnion, len(req.Messages))
	for i, m := range req.Messages {
		switch m.Role {
		case RoleSystem:
			messages[i] = openai.SystemMessage(m.Content)
		case RoleAssistant:
			messages[i] = openai.AssistantMessage(m.Content)
		default:
			messages[i] = openai.UserMessage(m.Content)
		}
	}

	params := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(req.Model),
		Messages: messages,
	}

	if req.MaxTokens > 0 {
		params.MaxTokens = openai.Int(int64(req.MaxTokens))
	}
	if req.Temperature > 0 {
		params.Temperature = openai.Float(float64(req.Temperature))
	}

	resp, err := p.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return CompletionResponse{}, err
	}

	content := ""
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Message.Content
	}

	return CompletionResponse{
		Content: content,
		Usage: TokenUsage{
			PromptTokens:     int(resp.Usage.PromptTokens),
			CompletionTokens: int(resp.Usage.CompletionTokens),
			TotalTokens:      int(resp.Usage.TotalTokens),
		},
	}, nil
}

func (p *OpenAIProvider) StreamComplete(ctx context.Context, req CompletionRequest) (<-chan CompletionResponse, <-chan error) {
	respChan := make(chan CompletionResponse, 1)
	errChan := make(chan error, 1)

	messages := make([]openai.ChatCompletionMessageParamUnion, len(req.Messages))
	for i, m := range req.Messages {
		switch m.Role {
		case RoleSystem:
			messages[i] = openai.SystemMessage(m.Content)
		case RoleAssistant:
			messages[i] = openai.AssistantMessage(m.Content)
		default:
			messages[i] = openai.UserMessage(m.Content)
		}
	}

	params := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(req.Model),
		Messages: messages,
	}
	if req.MaxTokens > 0 {
		params.MaxTokens = openai.Int(int64(req.MaxTokens))
	}
	if req.Temperature > 0 {
		params.Temperature = openai.Float(float64(req.Temperature))
	}

	go func() {
		defer close(respChan)
		defer close(errChan)

		stream := p.client.Chat.Completions.NewStreaming(ctx, params)
		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) > 0 {
				respChan <- CompletionResponse{
					Content: chunk.Choices[0].Delta.Content,
				}
			}
		}

		if err := stream.Err(); err != nil {
			errChan <- err
		}
	}()

	return respChan, errChan
}

func (p *OpenAIProvider) Embed(ctx context.Context, input string) ([]float32, error) {
	params := openai.EmbeddingNewParams{
		Model: openai.EmbeddingModelTextEmbedding3Small,
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(input),
		},
	}

	resp, err := p.client.Embeddings.New(ctx, params)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) > 0 {
		vec := make([]float32, len(resp.Data[0].Embedding))
		for i, v := range resp.Data[0].Embedding {
			vec[i] = float32(v)
		}
		return vec, nil
	}

	return nil, fmt.Errorf("no embedding returned")
}
