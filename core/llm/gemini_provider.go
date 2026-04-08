package llm

import (
	"context"
	"fmt"
	"io"
	"strings"

	"google.golang.org/genai"
)

type GeminiProvider struct {
	client *genai.Client
	model  string
	apiKey string
}

func NewGeminiProvider(apiKey, model string) (*GeminiProvider, error) {
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, err
	}
	return &GeminiProvider{client: client, model: model, apiKey: apiKey}, nil
}

func (g *GeminiProvider) Chat(ctx context.Context, req ChatRequest) (io.ReadCloser, error) {
	// Conversão simplificada para o formato do SDK Gemini
	contents := make([]*genai.Content, len(req.Messages))
	for i, msg := range req.Messages {
		role := "user"
		if msg.Role == "model" {
			role = "model"
		}
		contents[i] = &genai.Content{
			Parts: []*genai.Part{{Text: msg.Content}},
			Role:  role,
		}
	}

	resp, err := g.client.Models.GenerateContent(ctx, g.model, contents, nil)
	if err != nil {
		return nil, err
	}

	text := resp.Candidates[0].Content.Parts[0].Text
	return io.NopCloser(strings.NewReader(text)), nil
}

func (g *GeminiProvider) Embed(ctx context.Context, req EmbedRequest) ([]float32, error) {
	// EmbedContent espera []*genai.Content, não *genai.EmbedContentRequest
	contents := []*genai.Content{
		{
			Parts: []*genai.Part{{Text: req.Text}},
		},
	}

	res, err := g.client.Models.EmbedContent(ctx, g.model, contents, nil)
	if err != nil {
		return nil, err
	}

	// Embeddings é um slice (plural) - pegamos o primeiro
	if len(res.Embeddings) == 0 || res.Embeddings[0] == nil {
		return nil, fmt.Errorf("empty embedding response")
	}
	return res.Embeddings[0].Values, nil
}

func (g *GeminiProvider) Name() string {
	return g.model
}
