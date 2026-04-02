package tool

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

// WebFetchTool implementa a interface Tool para leitura de conteúdo de URLs.
type WebFetchTool struct{}

func NewWebFetchTool() *WebFetchTool {
	return &WebFetchTool{}
}

func (t *WebFetchTool) Name() string {
	return "web_fetch"
}

func (t *WebFetchTool) Description() string {
	return "Busca o conteúdo HTML/Markdown de uma URL específica. Use para ler documentações online ou artigos técnicos após uma busca web bem-sucedida."
}

func (t *WebFetchTool) Type() domain.ACPActionType {
	return domain.ACPActionRead
}

func (t *WebFetchTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "A URL completa (http/https) para extração de conteúdo.",
			},
		},
		"required": []string{"url"},
	}
}

func (t *WebFetchTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	url, ok := args["url"].(string)
	if !ok {
		return nil, fmt.Errorf("argumento 'url' é obrigatório")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("falha ao acessar URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("URL retornou status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler corpo da resposta: %w", err)
	}

	// Filtro simplificado de conteúdo (Idealmente converter para Markdown como o Jina faz)
	content := string(body)
	if len(content) > 10000 {
		content = content[:10000] + "..." // Limite para evitar estouro de tokens
	}

	return map[string]string{
		"url":     url,
		"content": content,
	}, nil
}
