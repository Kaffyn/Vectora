package llama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// LlamaInference cliente para o servidor llama.cpp (Chat/Completion)
type LlamaInference struct {
	serverURL string
	client    *http.Client
}

func NewLlamaInference(serverURL string) *LlamaInference {
	return &LlamaInference{
		serverURL: serverURL,
		client: &http.Client{
			Timeout: 60 * time.Second, // Timeout maior para inferência
		},
	}
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Messages    []ChatMessage `json:"messages"`
	Temperature float32       `json:"temperature"`
	Stream      bool          `json:"stream"`
}

type ChatCompletionResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

// Chat realiza uma chamada ao endpoint /v1/chat/completions (estilo OpenAI)
func (i *LlamaInference) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	reqBody, _ := json.Marshal(ChatCompletionRequest{
		Messages:    messages,
		Temperature: 0.7,
		Stream:      false, // SSE será implementado posteriormente se solicitado
	})

	url := fmt.Sprintf("%s/v1/chat/completions", i.serverURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := i.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("llama inference server access error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llama inference server error status: %s", resp.Status)
	}

	var chatResp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode chat response: %w", err)
	}

	if len(chatResp.Choices) > 0 {
		return chatResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no choices returned from llama server")
}

// GenerateResponse implementa domain.LLMProvider usando o template Qwen3
func (i *LlamaInference) GenerateResponse(ctx context.Context, prompt string, contextText string) (string, error) {
	systemMsg := "Você é o Vectora, um assistente técnico especialista em Desenvolvimento de Jogos (Godot, C++, Shaders). " +
		"Use o contexto fornecido para responder de forma precisa. Se não souber, diga que não encontrou nos arquivos locais."

	userContent := fmt.Sprintf("Contexto Local:\n%s\n\nPergunta: %s", contextText, prompt)

	messages := []ChatMessage{
		{Role: "system", Content: systemMsg},
		{Role: "user", Content: userContent},
	}

	return i.Chat(ctx, messages)
}
