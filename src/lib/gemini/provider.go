package gemini

import (
	"context"
	"fmt"
	"os"

	"github.com/Kaffyn/Vectora/src/core/domain"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiProvider struct {
	modelName string
}

func NewGeminiProvider(modelName string) *GeminiProvider {
	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}
	return &GeminiProvider{modelName: modelName}
}

func (p *GeminiProvider) GenerateResponse(ctx context.Context, prompt string, contextStr string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY não configurada no ambiente")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", err
	}
	defer client.Close()

	model := client.GenerativeModel(p.modelName)

	fullPrompt := fmt.Sprintf("Contexto RAG:\n%s\n\nPergunta do Usuário: %s", contextStr, prompt)

	resp, err := model.GenerateContent(ctx, genai.Text(fullPrompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "Sem resposta do Gemini.", nil
	}

	part := resp.Candidates[0].Content.Parts[0]
	return fmt.Sprintf("%v", part), nil
}

var _ domain.LLMProvider = (*GeminiProvider)(nil)
