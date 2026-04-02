package tool

import (
	"context"
	"fmt"
	"os"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

// ReadFileTool implementa a interface Tool para leitura de arquivos no Vectora.
type ReadFileTool struct{}

func NewReadFileTool() *ReadFileTool {
	return &ReadFileTool{}
}

func (t *ReadFileTool) Name() string {
	return "read_file"
}

func (t *ReadFileTool) Description() string {
	return "Lê o conteúdo integral de um arquivo do sistema de arquivos do usuário. Útil para entender e analisar código, logs ou especificações."
}

func (t *ReadFileTool) Type() domain.ACPActionType {
	return domain.ACPActionRead
}

func (t *ReadFileTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Caminho absoluto ou relativo para o arquivo.",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ReadFileTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("argumento 'path' é obrigatório e deve ser string")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler arquivo: %w", err)
	}

	return map[string]string{
		"path":    path,
		"content": string(content),
	}, nil
}
