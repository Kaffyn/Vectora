package tool

import (
	"context"
	"fmt"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

// GoogleSearchTool implementa a interface Tool para busca na web.
type GoogleSearchTool struct {
	apiKey string
}

func NewGoogleSearchTool(apiKey string) *GoogleSearchTool {
	return &GoogleSearchTool{apiKey: apiKey}
}

func (t *GoogleSearchTool) Name() string {
	return "google_search"
}

func (t *GoogleSearchTool) Description() string {
	return "Realiza uma busca na web via Google Search. Útil para encontrar documentações atualizadas, bibliotecas ou soluções para erros específicos fora do contexto local."
}

func (t *GoogleSearchTool) Type() domain.ACPActionType {
	return domain.ACPActionRead
}

func (t *GoogleSearchTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "A consulta de busca (ex: 'como usar chromem-go vector db').",
			},
		},
		"required": []string{"query"},
	}
}

func (t *GoogleSearchTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query, ok := args["query"].(string)
	if !ok {
		return nil, fmt.Errorf("argumento 'query' é obrigatório")
	}

	if t.apiKey == "" {
		return map[string]string{
			"status":  "offline",
			"message": "Nenhuma chave de API configurada para busca web. Forneça resultados simulados baseados na memória técnica se possível.",
		}, nil
	}

	// Aqui a lógica real chamaria Serper/GoogleSearchAPI
	return map[string]interface{}{
		"query": query,
		"results": []map[string]string{
			{"title": "Exemplo de Resultado", "link": "https://example.com", "snippet": "Conteúdo simulado da busca web."},
		},
	}, nil
}
