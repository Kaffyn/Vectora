package tool

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

// GrepSearchTool implementa a interface Tool para busca de texto em arquivos.
type GrepSearchTool struct{}

func NewGrepSearchTool() *GrepSearchTool {
	return &GrepSearchTool{}
}

func (t *GrepSearchTool) Name() string {
	return "grep_search"
}

func (t *GrepSearchTool) Description() string {
	return "Busca por um padrão de texto ou substring dentro de todos os arquivos de um diretório (recursivo). Útil para encontrar usos de funções ou variáveis no projeto."
}

func (t *GrepSearchTool) Type() domain.ACPActionType {
	return domain.ACPActionRead
}

func (t *GrepSearchTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "O padrão de texto a ser buscado.",
			},
			"root": map[string]interface{}{
				"type":        "string",
				"description": "O diretório inicial para a busca (default: '.').",
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *GrepSearchTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	pattern, ok := args["pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("argumento 'pattern' é obrigatório")
	}

	root := "."
	if r, ok := args["root"].(string); ok {
		root = r
	}

	results := make([]map[string]interface{}, 0)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Ignorar erros de acesso
		}
		if info.IsDir() || strings.Contains(path, ".git") {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		lineNum := 1
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, pattern) {
				results = append(results, map[string]interface{}{
					"path":       path,
					"lineNumber": lineNum,
					"content":    strings.TrimSpace(line),
				})
			}
			lineNum++
			if len(results) > 100 { // Limite de 100 resultados
				return fmt.Errorf("too many results")
			}
		}
		return nil
	})

	if err != nil && err.Error() != "too many results" {
		return nil, err
	}

	return map[string]interface{}{
		"pattern": pattern,
		"results": results,
		"count":   len(results),
	}, nil
}
