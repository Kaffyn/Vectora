package tool

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

// WriteFileTool implementa a interface Tool para criação de arquivos no Vectora.
type WriteFileTool struct{}

func NewWriteFileTool() *WriteFileTool {
	return &WriteFileTool{}
}

func (t *WriteFileTool) Name() string {
	return "write_file"
}

func (t *WriteFileTool) Description() string {
	return "Cria um novo arquivo ou sobrescreve um existente com o conteúdo fornecido. Automaticamente cria diretórios se necessário."
}

func (t *WriteFileTool) Type() domain.ACPActionType {
	return domain.ACPActionWrite
}

func (t *WriteFileTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Caminho do arquivo (ex: 'src/main.go').",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Conteúdo textual a ser gravado.",
			},
		},
		"required": []string{"path", "content"},
	}
}

func (t *WriteFileTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("argumento 'path' é obrigatório e deve ser string")
	}

	content, ok := args["content"].(string)
	if !ok {
		return nil, fmt.Errorf("argumento 'content' é obrigatório e deve ser string")
	}

	// Criar diretórios pai se faltarem
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("falha ao criar diretórios: %w", err)
	}

	// Escrever arquivo
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("falha ao gravar arquivo: %w", err)
	}

	return map[string]string{
		"path":   path,
		"status": "success",
	}, nil
}
