package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// knownOpenAIModels is a static list of known GPT models (from AGENTS.md, April 2026).
// Used as a fallback when the /models API endpoint is unavailable.
var knownOpenAIModels = []string{
	"gpt-5.4-pro", "gpt-5.4-mini", "gpt-5-o1", "gpt-5.4-nano",
	"text-embedding-3-large", "text-embedding-3-small",
}

// knownQwenModels is a static list of known Qwen models (from AGENTS.md, April 2026).
// DashScope API does not always expose a /models endpoint.
var knownQwenModels = []string{
	"qwen3.6-plus", "qwen3.6-turbo", "qwen3.5-omni", "qwen-max",
	"qwen3-embedding-8b", "qwen3-vl-embedding",
}

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

// buildMessages converts Vectora's unified message list to the OpenAI SDK format,
// correctly handling system, user, assistant, and tool-result messages.
func buildMessages(msgs []Message) []openai.ChatCompletionMessageParamUnion {
	result := make([]openai.ChatCompletionMessageParamUnion, 0, len(msgs))
	for _, m := range msgs {
		switch m.Role {
		case RoleSystem:
			result = append(result, openai.SystemMessage(m.Content))
		case RoleTool:
			// Tool results must include the tool_call_id that originated the call.
			result = append(result, openai.ToolMessage(m.ToolCallID, m.Content))
		case RoleAssistant:
			if len(m.ToolCalls) > 0 {
				// Assistant message with tool calls (function call request).
				var oaiCalls []openai.ChatCompletionMessageToolCallParam
				for _, tc := range m.ToolCalls {
					oaiCalls = append(oaiCalls, openai.ChatCompletionMessageToolCallParam{
						ID:   tc.ID,
						Type: "function",
						Function: openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      tc.Name,
							Arguments: tc.Args,
						},
					})
				}
				result = append(result, openai.ChatCompletionMessageParamUnion{
					OfAssistant: &openai.ChatCompletionAssistantMessageParam{
						Content:   openai.ChatCompletionAssistantMessageParamContentUnion{OfString: openai.String(m.Content)},
						ToolCalls: oaiCalls,
					},
				})
			} else {
				result = append(result, openai.AssistantMessage(m.Content))
			}
		default:
			result = append(result, openai.UserMessage(m.Content))
		}
	}
	return result
}

// buildTools converts Vectora's ToolDefinition slice to the OpenAI SDK format.
func buildTools(tools []ToolDefinition) []openai.ChatCompletionToolParam {
	oaiTools := make([]openai.ChatCompletionToolParam, 0, len(tools))
	for _, t := range tools {
		var rawParams map[string]interface{}
		_ = json.Unmarshal(t.Schema, &rawParams)

		oaiTools = append(oaiTools, openai.ChatCompletionToolParam{
			Type: "function",
			Function: openai.FunctionDefinitionParam{
				Name:        t.Name,
				Description: openai.String(t.Description),
				Parameters:  openai.FunctionParameters(rawParams),
			},
		})
	}
	return oaiTools
}

func (p *OpenAIProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	model := ResolveModel(p.name, req.Model)

	params := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(model),
		Messages: buildMessages(req.Messages),
	}

	if req.MaxTokens > 0 {
		params.MaxTokens = openai.Int(int64(req.MaxTokens))
	}
	if req.Temperature > 0 {
		params.Temperature = openai.Float(float64(req.Temperature))
	}
	if len(req.Tools) > 0 {
		params.Tools = buildTools(req.Tools)
	}

	resp, err := p.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return CompletionResponse{}, err
	}

	content := ""
	var tCalls []ToolCall
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Message.Content
		for _, tc := range resp.Choices[0].Message.ToolCalls {
			tCalls = append(tCalls, ToolCall{
				ID:   tc.ID,
				Name: tc.Function.Name,
				Args: tc.Function.Arguments,
			})
		}
	}

	return CompletionResponse{
		Content:   content,
		ToolCalls: tCalls,
		Usage: TokenUsage{
			PromptTokens:     int(resp.Usage.PromptTokens),
			CompletionTokens: int(resp.Usage.CompletionTokens),
			TotalTokens:      int(resp.Usage.TotalTokens),
		},
	}, nil
}

func (p *OpenAIProvider) StreamComplete(ctx context.Context, req CompletionRequest) (<-chan CompletionResponse, <-chan error) {
	respChan := make(chan CompletionResponse, 20)
	errChan := make(chan error, 1)

	model := ResolveModel(p.name, req.Model)

	params := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(model),
		Messages: buildMessages(req.Messages),
	}
	if req.MaxTokens > 0 {
		params.MaxTokens = openai.Int(int64(req.MaxTokens))
	}
	if req.Temperature > 0 {
		params.Temperature = openai.Float(float64(req.Temperature))
	}
	if len(req.Tools) > 0 {
		params.Tools = buildTools(req.Tools)
	}

	go func() {
		defer close(respChan)
		defer close(errChan)

		// Accumulate tool call deltas indexed by tool call index.
		toolCallAccum := make(map[int]*ToolCall)

		stream := p.client.Chat.Completions.NewStreaming(ctx, params)
		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) == 0 {
				continue
			}
			delta := chunk.Choices[0].Delta

			// Stream text content
			if delta.Content != "" {
				respChan <- CompletionResponse{Content: delta.Content}
			}

			// Accumulate tool call deltas
			for _, tc := range delta.ToolCalls {
				idx := int(tc.Index)
				if _, ok := toolCallAccum[idx]; !ok {
					toolCallAccum[idx] = &ToolCall{ID: tc.ID}
				}
				acc := toolCallAccum[idx]
				if tc.ID != "" {
					acc.ID = tc.ID
				}
				acc.Name += tc.Function.Name
				acc.Args += tc.Function.Arguments
			}
		}

		if err := stream.Err(); err != nil {
			errChan <- err
			return
		}

		// Emit accumulated tool calls as a final message if any
		if len(toolCallAccum) > 0 {
			var calls []ToolCall
			for i := 0; i < len(toolCallAccum); i++ {
				if tc, ok := toolCallAccum[i]; ok {
					calls = append(calls, *tc)
				}
			}
			respChan <- CompletionResponse{ToolCalls: calls}
		}
	}()

	return respChan, errChan
}

func (p *OpenAIProvider) Embed(ctx context.Context, input string, model string) ([]float32, error) {
	embeddingModel := ResolveEmbeddingModel(p.name, model)
	if embeddingModel == "" {
		embeddingModel = "text-embedding-3-large"
	}

	params := openai.EmbeddingNewParams{
		Model: openai.EmbeddingModel(embeddingModel),
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

func (p *OpenAIProvider) ListModels(ctx context.Context) ([]string, error) {
	resp, err := p.client.Models.List(ctx)
	if err != nil {
		// Fallback to static model list for providers that don't expose /models
		// (e.g. Qwen DashScope, custom local endpoints).
		if strings.Contains(p.name, "qwen") {
			return knownQwenModels, nil
		}
		return knownOpenAIModels, nil
	}
	var models []string
	for _, m := range resp.Data {
		models = append(models, m.ID)
	}
	return models, nil
}
