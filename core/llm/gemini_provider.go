package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type GeminiProvider struct {
	apiKey string
	client *http.Client
}

func NewGeminiProvider(ctx context.Context, apiKey string) (*GeminiProvider, error) {
	if apiKey == "" {
		return nil, errors.New("gemini_api_key_required: Gemini API key was not provided")
	}

	return &GeminiProvider{
		apiKey: apiKey,
		client: &http.Client{},
	}, nil
}

// Supported Gemini models.
const (
	// GeminiFlash is the default model: fast, low-cost, ideal for RAG queries.
	GeminiFlash = "gemini-3-flash-preview"
	// GeminiPro is the reasoning model: used for complex agentic tasks.
	GeminiPro = "gemini-3.1-pro-preview"
	// GeminiEmbedding is the current embedding model (3072 dims).
	GeminiEmbedding = "gemini-embedding-2-preview"
)

func (p *GeminiProvider) Name() string {
	return "gemini"
}

func (p *GeminiProvider) IsConfigured() bool {
	return p.apiKey != ""
}

type geminiPart struct {
	Text       string      `json:"text,omitempty"`
	InlineData *inlineData `json:"inlineData,omitempty"`
}

type inlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	Contents          []geminiContent          `json:"contents"`
	SystemInstruction *geminiContent           `json:"system_instruction,omitempty"`
	GenerationConfig  map[string]interface{}   `json:"generationConfig,omitempty"`
	Tools             []map[string]interface{} `json:"tools,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content      geminiContent `json:"content"`
		FinishReason string        `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

func (p *GeminiProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	model := req.Model
	if model == "" {
		model = GeminiFlash
	}
	if !strings.HasPrefix(model, "models/") {
		model = "models/" + strings.TrimPrefix(model, "google/")
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/%s:generateContent?key=%s", model, p.apiKey)

	// Gemini separates system messages from the contents array
	var systemText string
	var contents []geminiContent
	for _, msg := range req.Messages {
		if msg.Role == RoleSystem {
			systemText += msg.Content + "\n"
			continue
		}
		role := "user"
		if msg.Role == RoleAssistant {
			role = "model"
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: msg.Content}},
		})
	}

	geminiReq := geminiRequest{
		Contents: contents,
		GenerationConfig: map[string]interface{}{
			"temperature":     req.Temperature,
			"maxOutputTokens": req.MaxTokens,
		},
	}

	if systemText != "" {
		geminiReq.SystemInstruction = &geminiContent{
			Parts: []geminiPart{{Text: strings.TrimSpace(systemText)}},
		}
	}

	// Tool mapping
	if len(req.Tools) > 0 {
		var functions []map[string]interface{}
		for _, t := range req.Tools {
			var params map[string]interface{}
			json.Unmarshal(t.Schema, &params)
			functions = append(functions, map[string]interface{}{
				"name":        t.Name,
				"description": t.Description,
				"parameters":  params,
			})
		}
		geminiReq.Tools = []map[string]interface{}{
			{"function_declarations": functions},
		}
	}

	body, _ := json.Marshal(geminiReq)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return CompletionResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return CompletionResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return CompletionResponse{}, fmt.Errorf("gemini API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var gResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&gResp); err != nil {
		return CompletionResponse{}, err
	}

	if len(gResp.Candidates) == 0 {
		return CompletionResponse{}, errors.New("llm_empty_response")
	}

	candidate := gResp.Candidates[0]
	var content string

	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			content += part.Text
		}
		// TODO: Implement native Tool Call parsing for Gemini here if needed in future
	}

	return CompletionResponse{
		Content: content,
		Usage: TokenUsage{
			PromptTokens:     gResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: gResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      gResp.UsageMetadata.TotalTokenCount,
		},
	}, nil
}

func (p *GeminiProvider) Embed(ctx context.Context, input string) ([]float32, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent?key=%s", GeminiEmbedding, p.apiKey)

	reqBody := map[string]interface{}{
		"model": "models/" + GeminiEmbedding,
		"content": map[string]interface{}{
			"parts": []map[string]interface{}{{"text": input}},
		},
	}

	body, _ := json.Marshal(reqBody)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Embedding struct {
			Values []float32 `json:"values"`
		} `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Embedding.Values, nil
}
