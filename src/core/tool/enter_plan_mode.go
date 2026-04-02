package tool

import (
	"context"
	"fmt"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

// EnterPlanModeTool implementa a interface Tool para orquestração de planejamento agêntico.
type EnterPlanModeTool struct{}

func NewEnterPlanModeTool() *EnterPlanModeTool {
	return &EnterPlanModeTool{}
}

func (t *EnterPlanModeTool) Name() string {
	return "enter_plan_mode"
}

func (t *EnterPlanModeTool) Description() string {
	return `Inicia o modo de planejamento para tarefas complexas que exigem múltiplos passos. 
Use para esboçar a estratégia de solução antes de executar comandos ou editar arquivos. 
Bloqueia ações de escrita até que o plano seja aceito pelo contexto.`
}

func (t *EnterPlanModeTool) Type() domain.ACPActionType {
	return domain.ACPActionRead
}

func (t *EnterPlanModeTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"goal": map[string]interface{}{
				"type":        "string",
				"description": "O objetivo final do plano (ex: 'refatorar o banco de dados para usar Prisma').",
			},
			"steps": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Os passos sequenciais para atingir o objetivo.",
			},
		},
		"required": []string{"goal", "steps"},
	}
}

func (t *EnterPlanModeTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	goal, ok := args["goal"].(string)
	if !ok {
		return nil, fmt.Errorf("argumento 'goal' é obrigatório")
	}

	steps, _ := args["steps"].([]interface{})

	// Em uma implementação real, isso acionaria um estado no agente ou um log persistente
	return map[string]interface{}{
		"status": "planning_active",
		"goal":   goal,
		"steps":  steps,
		"notice": "O plano foi registrado. Prossiga com a execução conforme as diretrizes aprovadas.",
	}, nil
}
