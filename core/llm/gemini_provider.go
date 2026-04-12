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

			// NEW: If we have raw Gemini content in metadata, use it as is
			if rawContent, ok := m.Metadata["gemini_content"]; ok {
				if contentBytes, err := json.Marshal(rawContent); err == nil {
					var gContent genai.Content
					if err := json.Unmarshal(contentBytes, &gContent); err == nil {
						contents = append(contents, &gContent)
						continue
					}
				}
			}

			var parts []*genai.Part
			if m.Content != "" {
				parts = append(parts, &genai.Part{Text: m.Content})
			}
			for _, tc := range m.ToolCalls {
				// Reconstruct function call for history
				var args map[string]any
				_ = json.Unmarshal([]byte(tc.Args), &args)
				parts = append(parts, &genai.Part{
					FunctionCall: &genai.FunctionCall{
						Name: tc.Name,
						Args: args,
					},
				})
			}
			contents = append(contents, &genai.Content{
				Role:  role,
				Parts: parts,
			})
			continue
		}
		if m.Role == RoleTool {
			// Tool results (FunctionResponse) must be in a "function" role or similar context depending on SDK
			// In genai SDK, we're likely using contents with FunctionResponse parts
			contents = append(contents, &genai.Content{
				Role: "function",
				Parts: []*genai.Part{
					{
						FunctionResponse: &genai.FunctionResponse{
							Name:     m.ToolCallID, // In Vectora, we sometimes use name as ID for Gemini
							Response: map[string]any{"result": m.Content},
						},
					},
				},
			})
			continue
		}

		contents = append(contents, &genai.Content{
			Role: role,
			Parts: []*genai.Part{
				{Text: m.Content},
			},
		})
	}

	// Resolve model alias to canonical Gemini API model ID with "-preview" suffix
	modelID := ResolveModel("gemini", req.Model)

	resp, err := p.client.Models.GenerateContent(ctx, modelID, contents, config)
	if err != nil {
		// Attempting once a fallback if permitted
		fallbackModel := req.FallbackModel
		if fallbackModel == "" {
			fallbackModel = "gemini-3-flash-preview"
		}

		if fallbackModel != req.Model {
			fallbackReq := req
			fallbackReq.Model = fallbackModel
			fallbackReq.FallbackModel = ""
			return p.Complete(ctx, fallbackReq)
		}
		return CompletionResponse{}, p.wrapError(err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return CompletionResponse{}, errors.New("gemini_no_content_returned")
	}

	var content string
	var tCalls []ToolCall

	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			content += part.Text
		}
		if part.FunctionCall != nil {
			fc := part.FunctionCall
			argsJSON, _ := json.Marshal(fc.Args)
			tCalls = append(tCalls, ToolCall{
				ID:   fc.Name, // Gemini uses name as the link for function responses
				Name: fc.Name,
				Args: string(argsJSON),
			})
		}
	}

	return CompletionResponse{
		Content:   content,
		ToolCalls: tCalls,
		Metadata: map[string]any{
			"gemini_content": resp.Candidates[0].Content,
		},
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
				var parts []*genai.Part
				if m.Content != "" {
					parts = append(parts, &genai.Part{Text: m.Content})
				}
				for _, tc := range m.ToolCalls {
					var args map[string]any
					_ = json.Unmarshal([]byte(tc.Args), &args)
					parts = append(parts, &genai.Part{
						FunctionCall: &genai.FunctionCall{
							Name: tc.Name,
							Args: args,
						},
					})
				}
				contents = append(contents, &genai.Content{
					Role:  role,
					Parts: parts,
				})
				continue
			}
			if m.Role == RoleTool {
				contents = append(contents, &genai.Content{
					Role: "function",
					Parts: []*genai.Part{
						{
							FunctionResponse: &genai.FunctionResponse{
								Name:     m.ToolCallID,
								Response: map[string]any{"result": m.Content},
							},
						},
					},
				})
				continue
			}
			contents = append(contents, &genai.Content{
				Role:  role,
				Parts: []*genai.Part{{Text: m.Content}},
			})
		}

		// Resolve model alias to canonical Gemini API model ID with "-preview" suffix
		modelID := ResolveModel("gemini", req.Model)

		iter := p.client.Models.GenerateContentStream(ctx, modelID, contents, config)
		for resp, err := range iter {
			if err != nil {
				errChan <- p.wrapError(err)
				return
			}
			if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
				var content string
				var tCalls []ToolCall
				for _, part := range resp.Candidates[0].Content.Parts {
					if part.Text != "" {
						content += part.Text
					}
					if part.FunctionCall != nil {
						fc := part.FunctionCall
						argsJSON, _ := json.Marshal(fc.Args)
						tCalls = append(tCalls, ToolCall{
							ID:   fc.Name,
							Name: fc.Name,
							Args: string(argsJSON),
						})
					}
				}
				respChan <- CompletionResponse{
					Content:   content,
					ToolCalls: tCalls,
					Metadata: map[string]any{
						"gemini_content": resp.Candidates[0].Content,
					},
				}
			}
		}
	}()

	return respChan, errChan
}

func (p *GeminiProvider) Embed(ctx context.Context, input string, model string) ([]float32, error) {
	// Specialized resolution for embeddings to avoid 404/NOT_FOUND with chat models
	embeddingModel := ResolveEmbeddingModel("google", model)
	if embeddingModel == "" {
		embeddingModel = "gemini-embedding-2-preview"
	}

	// Add Task Instructions for Gemini 2.0 (improves retrieval quality)
	// Reference: https://ai.google.dev/gemini-api/docs/models/gemini/embedding-2
	prefix := ""
	if len(input) > 200 {
		prefix = "task:code retrieval\n"
	} else if len(input) > 0 {
		prefix = "task:search query\n"
	}

	resp, err := p.client.Models.EmbedContent(ctx, embeddingModel, []*genai.Content{{
		Parts: []*genai.Part{{Text: prefix + input}},
	}}, nil)
	if err != nil {
		// If we hit a rate limit (429), try once more after a short wait or just wrap it
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "RESOURCE_EXHAUSTED") {
			return nil, fmt.Errorf("gemini_embedding_quota_exceeded: %w", err)
		}
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
