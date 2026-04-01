package llm

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

type GeminiProvider struct {
	client       *googleai.GoogleAI
	systemPrompt string
	toolsSpec    string
}

func NewGeminiProvider(ctx context.Context, apiKey string) (*GeminiProvider, error) {
	if apiKey == "" {
		return nil, errors.New("gemini_api_key_required: Gemini API key was not provided")
	}

	os.Setenv("GEMINI_API_KEY", apiKey)

	client, err := googleai.New(ctx, 
		googleai.WithAPIKey(apiKey), 
		googleai.WithDefaultModel("gemini-1.5-flash"),
		googleai.WithDefaultEmbeddingModel("gemini-embedding-2-preview"),
	)
	if err != nil {
		return nil, err
	}
	
	p, t := LoadMasterInstructions()
	
	return &GeminiProvider{
		client:       client,
		systemPrompt: p,
		toolsSpec:    t,
	}, nil
}

func (p *GeminiProvider) Name() string {
	return "gemini"
}

func (p *GeminiProvider) IsConfigured() bool {
	return p.client != nil
}

func (p *GeminiProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	var content []llms.MessageContent
	
	if req.SystemPrompt == "" {
		req.SystemPrompt = p.systemPrompt + "\n\nAVAILABLE TOOLS SPECIFICATION (JSON SCHEMA):\n" + p.toolsSpec
	}

	if req.SystemPrompt != "" {
		content = append(content, llms.TextParts(llms.ChatMessageTypeSystem, req.SystemPrompt))
	}

	for _, msg := range req.Messages {
		var chatRole llms.ChatMessageType
		switch msg.Role {
		case RoleUser:
			chatRole = llms.ChatMessageTypeHuman
		case RoleAssistant:
			chatRole = llms.ChatMessageTypeAI
		case RoleSystem:
			chatRole = llms.ChatMessageTypeSystem
		case RoleTool:
			chatRole = llms.ChatMessageTypeTool
		default:
			chatRole = llms.ChatMessageTypeHuman 
		}
		content = append(content, llms.TextParts(chatRole, msg.Content))
	}

	var langchanTools []llms.Tool
	for _, t := range req.Tools {
		var params interface{}
		if err := json.Unmarshal(t.Schema, &params); err != nil {
			params = nil // Fallback if schema is invalid
		}

		langchanTools = append(langchanTools, llms.Tool{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  params,
			},
		})
	}

	modelName := req.Model
	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}
	modelName = strings.TrimPrefix(modelName, "google/")

	opts := []llms.CallOption{
		llms.WithModel(modelName),
		llms.WithMaxTokens(req.MaxTokens),
		llms.WithTemperature(float64(req.Temperature)),
	}
	
	if len(langchanTools) > 0 {
		opts = append(opts, llms.WithTools(langchanTools))
	}
	
	resp, err := p.client.GenerateContent(ctx, content, opts...)
	if err != nil {
		return CompletionResponse{}, err
	}

	if len(resp.Choices) == 0 {
		return CompletionResponse{}, errors.New("llm_empty_response: no choices were returned by the provider")
	}

	choice := resp.Choices[0]
	
	var tCalls []ToolCall
	for _, lcTc := range choice.ToolCalls {
		tCalls = append(tCalls, ToolCall{
			ID:   lcTc.ID,
			Name: lcTc.FunctionCall.Name,
			Args: lcTc.FunctionCall.Arguments,
		})
	}
	
	return CompletionResponse{
		Content:   choice.Content,
		ToolCalls: tCalls,
		Usage: TokenUsage{},
	}, nil
}

func (p *GeminiProvider) Embed(ctx context.Context, input string) ([]float32, error) {
	embClient, err := p.client.CreateEmbedding(ctx, []string{input})
	if err != nil {
		return nil, err
	}
	if len(embClient) > 0 {
		return embClient[0], nil
	}
	return nil, errors.New("gemini_embedding_failed: no vectors returned")
}
