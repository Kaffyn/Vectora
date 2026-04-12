package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

type GeminiProvider struct {
	apiKey string
	client *genai.Client
}

func NewGeminiProvider(ctx context.Context, apiKey string) (*GeminiProvider, error) {
	if apiKey == "" {
		return nil, errors.New("gemini_api_key_required: Gemini API key was not provided")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("gemini_init_error: %w", err)
	}

	return &GeminiProvider{
		apiKey: apiKey,
		client: client,
	}, nil
}

func (p *GeminiProvider) Name() string {
	return "gemini"
}

func (p *GeminiProvider) IsConfigured() bool {
	return p.apiKey != ""
}

func (p *GeminiProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	config := p.prepareConfig(req)

	var contents []*genai.Content
	for _, m := range req.Messages {
		role := "user"
		if m.Role == RoleAssistant {
			role = "model"
		}
		contents = append(contents, &genai.Content{
			Role: role,
			Parts: []*genai.Part{
				{Text: m.Content},
			},
		})
	}

	// Resolve model alias to canonical Gemini API model ID with "-preview" suffix
	modelID := ResolveGeminiModel(req.Model)

	resp, err := p.client.Models.GenerateContent(ctx, modelID, contents, config)
	if err != nil {
		return CompletionResponse{}, p.wrapError(err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return CompletionResponse{}, errors.New("gemini_no_content_returned")
	}

	content := resp.Candidates[0].Content.Parts[0].Text

	var tCalls []ToolCall
	// Parse tools from Gemini parts in future phases

	return CompletionResponse{
		Content:   content,
		ToolCalls: tCalls,
		Usage: TokenUsage{
			PromptTokens:     int(resp.UsageMetadata.PromptTokenCount),
			CompletionTokens: int(resp.UsageMetadata.CandidatesTokenCount),
			TotalTokens:      int(resp.UsageMetadata.TotalTokenCount),
		},
	}, nil
}

func (p *GeminiProvider) StreamComplete(ctx context.Context, req CompletionRequest) (<-chan CompletionResponse, <-chan error) {
	respChan := make(chan CompletionResponse, 20)
	errChan := make(chan error, 1)

	go func() {
		defer close(respChan)
		defer close(errChan)

		config := p.prepareConfig(req)
		var contents []*genai.Content
		for _, m := range req.Messages {
			role := "user"
			if m.Role == RoleAssistant {
				role = "model"
			}
			contents = append(contents, &genai.Content{
				Role:  role,
				Parts: []*genai.Part{{Text: m.Content}},
			})
		}

		// Resolve model alias to canonical Gemini API model ID with "-preview" suffix
		modelID := ResolveGeminiModel(req.Model)

		iter := p.client.Models.GenerateContentStream(ctx, modelID, contents, config)
		for resp, err := range iter {
			if err != nil {
				errChan <- p.wrapError(err)
				return
			}
			if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
				respChan <- CompletionResponse{
					Content: resp.Candidates[0].Content.Parts[0].Text,
				}
			}
		}
	}()

	return respChan, errChan
}

func (p *GeminiProvider) Embed(ctx context.Context, input string, model string) ([]float32, error) {
	embeddingModel := "gemini-embedding-2-preview" // Correct ID from models-sdk.md

	// Family detection
	if strings.Contains(strings.ToLower(model), "gemini") {
		embeddingModel = "gemini-embedding-2-preview"
	}

	resp, err := p.client.Models.EmbedContent(ctx, embeddingModel, []*genai.Content{{
		Parts: []*genai.Part{{Text: input}},
	}}, nil)
	if err != nil {
		return nil, p.wrapError(err)
	}

	if len(resp.Embeddings) == 0 {
		return nil, errors.New("gemini_no_embeddings_returned")
	}

	return resp.Embeddings[0].Values, nil
}

func (p *GeminiProvider) ListModels(ctx context.Context) ([]string, error) {
	iter, _ := p.client.Models.List(ctx, nil)
	var models []string
	for {
		resp, err := iter.Next(ctx)
		if err != nil {
			break
		}
		models = append(models, resp.Name)
	}
	return models, nil
}

func (p *GeminiProvider) prepareConfig(req CompletionRequest) *genai.GenerateContentConfig {
	temp32 := float32(req.Temperature)
	config := &genai.GenerateContentConfig{
		Temperature:     &temp32,
		MaxOutputTokens: int32(req.MaxTokens),
	}

	if req.SystemPrompt != "" {
		config.SystemInstruction = &genai.Content{
			Parts: []*genai.Part{{Text: req.SystemPrompt}},
		}
	}

	if len(req.Tools) > 0 {
		var tools []*genai.Tool
		var functions []*genai.FunctionDeclaration
		for _, t := range req.Tools {
			var schema genai.Schema
			_ = json.Unmarshal(t.Schema, &schema)
			functions = append(functions, &genai.FunctionDeclaration{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  &schema,
			})
		}
		tools = append(tools, &genai.Tool{
			FunctionDeclarations: functions,
		})
		config.Tools = tools
	}

	return config
}

func (p *GeminiProvider) wrapError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("gemini_sdk_error: %w", err)
}
