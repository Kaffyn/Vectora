package tool

import (
	"context"
	"fmt"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

// ShellTool implementa a interface Tool para execução de comandos de sistema.
type ShellTool struct {
	osManager domain.OSManager
}

func NewShellTool(osm domain.OSManager) *ShellTool {
	return &ShellTool{osManager: osm}
}

func (t *ShellTool) Name() string {
	return "run_shell_command"
}

func (t *ShellTool) Description() string {
	return `Executa um comando de sistema (Shell) no terminal do usuário. 
Use para compilar projetos, listar processos, formatar código ou qualquer ação de terminal. 
Retorna a saída combinada (stdout/stderr).`
}

func (t *ShellTool) Type() domain.ACPActionType {
	return domain.ACPActionExecute
}

func (t *ShellTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "O comando a ser executado (ex: 'go build', 'dir', 'ls')",
			},
			"arguments": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Lista de argumentos para o comando.",
			},
		},
		"required": []string{"command"},
	}
}

func (t *ShellTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	cmd, ok := args["command"].(string)
	if !ok {
		return nil, fmt.Errorf("argumento 'command' é obrigatório e deve ser string")
	}

	var arguments []string
	if rawArgs, ok := args["arguments"].([]interface{}); ok {
		for _, a := range rawArgs {
			if s, ok := a.(string); ok {
				arguments = append(arguments, s)
			}
		}
	}

	// Delegar para o OSManager que cuida da plataforma (Vulkan/Metal/Nativo)
	output, err := t.osManager.RunCommand(ctx, cmd, arguments)
	if err != nil {
		return map[string]string{
			"status": "error",
			"output": output,
			"error":  err.Error(),
		}, nil
	}

	return map[string]string{
		"status": "success",
		"output": output,
	}, nil
}
