package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type ClaudeProvider struct {
	apiKey string
	client anthropic.Client
}

func NewClaudeProvider(ctx context.Context, apiKey string) (*ClaudeProvider, error) {
	if apiKey == "" {
		return nil, errors.New("claude_api_key_required: Claude API key was not provided")
	}

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	// NewClient returns Client (value), not pointer
	return &ClaudeProvider{
		apiKey: apiKey,
		client: client,
	}, nil
}

func (p *ClaudeProvider) Name() string {
	return "claude"
}

func (p *ClaudeProvider) IsConfigured() bool {
	return p.apiKey != ""
}

func (p *ClaudeProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	params := p.prepareParams(req)

	msg, err := p.client.Messages.New(ctx, params)
	if err != nil {
		if req.FallbackModel != "" && req.FallbackModel != req.Model {
			// Try fallback model
			fallbackReq := req
			fallbackReq.Model = req.FallbackModel
			fallbackReq.FallbackModel = "" // Avoid infinite recursion
			return p.Complete(ctx, fallbackReq)
		}
		return CompletionResponse{}, p.wrapError(err)
	}

	var content string
	var tCalls []ToolCall
	for _, union := range msg.Content {
		switch block := union.AsAny().(type) {
		case anthropic.TextBlock:
			content += block.Text
		case anthropic.ToolUseBlock:
			tCalls = append(tCalls, ToolCall{
				ID:   block.ID,
				Name: block.Name,
				Args: string(block.Input),
			})
		}
	}

	return CompletionResponse{
		Content:   content,
		ToolCalls: tCalls,
		Usage: TokenUsage{
			PromptTokens:     int(msg.Usage.InputTokens),
			CompletionTokens: int(msg.Usage.OutputTokens),
			TotalTokens:      int(msg.Usage.InputTokens + msg.Usage.OutputTokens),
		},
	}, nil
}

func (p *ClaudeProvider) StreamComplete(ctx context.Context, req CompletionRequest) (<-chan CompletionResponse, <-chan error) {
	respChan := make(chan CompletionResponse, 20)
	errChan := make(chan error, 1)

	go func() {
		defer close(respChan)
		defer close(errChan)

		params := p.prepareParams(req)
		stream := p.client.Messages.NewStreaming(ctx, params)

		for stream.Next() {
			event := stream.Current()
			switch ev := event.AsAny().(type) {
			case anthropic.ContentBlockDeltaEvent:
				// Stream text content
				if txt, ok := ev.Delta.AsAny().(anthropic.TextDelta); ok {
					respChan <- CompletionResponse{
						Content: txt.Text,
					}
				}
				// Note: Tool call partial JSON is available during streaming via ToolUseInputJsonDelta,
				// but we accumulate complete tool calls until stream ends (see TODO below)
			}
		}

		if err := stream.Err(); err != nil {
			errChan <- p.wrapError(err)
			return
		}

		// TODO: Extract complete tool calls from final message
		// The anthropic SDK's streaming interface provides tool call data through:
		// - ContentBlockStartEvent (marks start of tool use block)
		// - ToolUseInputJsonDelta events (partial JSON arguments)
		// - ContentBlockStopEvent (marks end of tool use block)
		//
		// We should accumulate these into complete ToolCall objects and emit them
		// as a final batch, similar to how Complete() does it.
		// Currently this is partially implemented; full implementation pending
		// availability of SDK type definitions for these delta events.
	}()

	return respChan, errChan
}

func (p *ClaudeProvider) Embed(ctx context.Context, input string, model string) ([]float32, error) {
	return nil, errors.New("claude_no_native_embedding")
}

func (p *ClaudeProvider) ListModels(ctx context.Context) ([]string, error) {
	var models []string
	for k := range claudeAliases {
		if !strings.Contains(k, "-") || k == "sonnet" || k == "opus" || k == "haiku" {
			models = append(models, k)
		}
	}
	// Return a balanced list
	return models, nil
}

// claudeAliases maps shorthand model names to canonical Anthropic model IDs.
// Uses SDK constants where available so typos are caught at compile time.
var claudeAliases = map[string]anthropic.Model{
	// Claude 4.6 series (latest) — hyphen format (internal style)
	"claude-sonnet-4-6": anthropic.ModelClaudeSonnet4_6,
	"claude-opus-4-6":   anthropic.ModelClaudeOpus4_6,
	// Claude 4.6 series — dot format (AGENTS.md / OpenRouter style)
	"claude-4.6-sonnet": anthropic.ModelClaudeSonnet4_6,
	"claude-4.6-opus":   anthropic.ModelClaudeOpus4_6,
	// Claude 4.5 series — hyphen format
	"claude-sonnet-4-5": anthropic.ModelClaudeSonnet4_5,
	"claude-haiku-4-5":  anthropic.ModelClaudeHaiku4_5,
	"claude-opus-4-5":   anthropic.ModelClaudeOpus4_5,
	// Claude 4.5 series — dot format
	"claude-4.5-sonnet": anthropic.ModelClaudeSonnet4_5,
	"claude-4.5-haiku":  anthropic.ModelClaudeHaiku4_5,
	"claude-4.5-opus":   anthropic.ModelClaudeOpus4_5,
	// Convenience shorthands
	"sonnet":     anthropic.ModelClaudeSonnet4_6,
	"opus":       anthropic.ModelClaudeOpus4_6,
	"haiku":      anthropic.ModelClaudeHaiku4_5,
	"sonnet-4-6": anthropic.ModelClaudeSonnet4_6,
	"opus-4-6":   anthropic.ModelClaudeOpus4_6,
}

func (p *ClaudeProvider) prepareParams(req CompletionRequest) anthropic.MessageNewParams {
	modelID := anthropic.Model(req.Model)
	if modelID == "" {
		modelID = anthropic.ModelClaudeHaiku4_5
	} else if resolved, ok := claudeAliases[strings.ToLower(req.Model)]; ok {
		modelID = resolved
	}

	var messages []anthropic.MessageParam
	for _, msg := range req.Messages {
		if msg.Role == RoleSystem {
			continue
		}

		var m anthropic.MessageParam
		switch msg.Role {
		case RoleAssistant:
			// When assistant has tool calls, include them in the message
			if len(msg.ToolCalls) > 0 {
				var blocks []anthropic.ContentBlockParamUnion
				// Add text content if present
				if msg.Content != "" {
					blocks = append(blocks, anthropic.NewTextBlock(msg.Content))
				}
				// Add tool use blocks
				for _, tc := range msg.ToolCalls {
					// Parse tool arguments (they're in JSON string format)
					var input any
					_ = json.Unmarshal([]byte(tc.Args), &input)
					blocks = append(blocks, anthropic.NewToolUseBlock(tc.ID, input, tc.Name))
				}
				m = anthropic.NewAssistantMessage(blocks...)
			} else {
				m = anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content))
			}
		case RoleTool:
			// Tool result message
			m = anthropic.NewUserMessage(anthropic.NewToolResultBlock(msg.ToolCallID, msg.Content, false))
		default:
			m = anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content))
		}
		messages = append(messages, m)
	}

	params := anthropic.MessageNewParams{
		Model:     modelID,
		MaxTokens: int64(req.MaxTokens),
		Messages:  messages,
	}

	if req.SystemPrompt != "" {
		params.System = []anthropic.TextBlockParam{{
			Text: req.SystemPrompt,
		}}
	}

	if len(req.Tools) > 0 {
		var tools []anthropic.ToolUnionParam
		for _, t := range req.Tools {
			var rawSchema map[string]interface{}
			_ = json.Unmarshal(t.Schema, &rawSchema)

			schema := anthropic.ToolInputSchemaParam{}
			if props, ok := rawSchema["properties"]; ok {
				schema.Properties = props
			}
			if reqs, ok := rawSchema["required"].([]interface{}); ok {
				var rStrings []string
				for _, r := range reqs {
					if s, ok := r.(string); ok {
						rStrings = append(rStrings, s)
					}
				}
				schema.Required = rStrings
			}

			tools = append(tools, anthropic.ToolUnionParam{
				OfTool: &anthropic.ToolParam{
					Name:        t.Name,
					Description: anthropic.String(t.Description),
					InputSchema: schema,
				},
			})
		}
		params.Tools = tools
	}

	if req.Temperature > 0 {
		params.Temperature = anthropic.Float(float64(req.Temperature))
	}

	return params
}

func (p *ClaudeProvider) wrapError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("claude_sdk_error: %w", err)
}
