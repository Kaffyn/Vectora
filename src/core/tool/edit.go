package tool

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

// EditTool implementa a interface Tool para edição de arquivos via substituição de texto.
type EditTool struct{}

func NewEditTool() *EditTool {
	return &EditTool{}
}

func (t *EditTool) Name() string {
	return "edit"
}

func (t *EditTool) Description() string {
	return `Edita um arquivo existente substituindo um bloco de texto específico por um novo. 
É mais seguro que o write_file para modificações pontuais em arquivos grandes. 
Requer que o texto de alvo (target) seja único no arquivo.`
}

func (t *EditTool) Type() domain.ACPActionType {
	return domain.ACPActionWrite
}

func (t *EditTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Caminho do arquivo.",
			},
			"target": map[string]interface{}{
				"type":        "string",
				"description": "O bloco de texto EXATO que deve ser substituído (incluindo espaços e quebras de linha).",
			},
			"replacement": map[string]interface{}{
				"type":        "string",
				"description": "O novo conteúdo que substituirá o alvo.",
			},
		},
		"required": []string{"path", "target", "replacement"},
	}
}

func (t *EditTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("argumento 'path' é obrigatório")
	}
	target, ok := args["target"].(string)
	if !ok {
		return nil, fmt.Errorf("argumento 'target' é obrigatório")
	}
	replacement, ok := args["replacement"].(string)
	if !ok {
		return nil, fmt.Errorf("argumento 'replacement' é obrigatório")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler arquivo para edição: %w", err)
	}

	// Verificar unicidade do alvo no arquivo
	sContent := string(content)
	count := strings.Count(sContent, target)
	if count == 0 {
		return nil, fmt.Errorf("o texto de alvo não foi encontrado no arquivo")
	}
	if count > 1 {
		return nil, fmt.Errorf("o texto de alvo não é único (encontrado %d vezes). Forneça um contexto maior para garantir unicidade", count)
	}

	// Substituir
	newContent := strings.Replace(sContent, target, replacement, 1)

	// Gravar
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("falha ao gravar arquivo após edição: %w", err)
	}

	return map[string]string{
		"path":   path,
		"status": "success",
	}, nil
}
