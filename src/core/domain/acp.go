package domain

import (
	"context"
)

// ACPPermissionLevel define o grau de autonomia do agente.
type ACPPermissionLevel string

const (
	ACPLevelAskAny    ACPPermissionLevel = "ask-any"    // Confirmar TODA e QUALQUER ação
	ACPLevelGuarded   ACPPermissionLevel = "guarded"   // Confirmar apenas ações de Escrita/Execução
	ACPLevelYOLO      ACPPermissionLevel = "yolo"      // Autonomia total sem confirmação
)

// ACPActionType define a natureza da tarefa autônoma.
type ACPActionType string

const (
	ACPActionRead    ACPActionType = "read"    // Leitura silenciosa (ls, cat, git status)
	ACPActionWrite   ACPActionType = "write"   // Alteração de arquivos (git add, sed, echo)
	ACPActionExecute ACPActionType = "execute" // Execução de binários (go build, bun run)
)

// ACPTask representa uma solicitação de ação autônoma pela IA.
type ACPTask struct {
	ID        string        `json:"id"`
	Type      ACPActionType `json:"type"`
	Command   string        `json:"command"`
	Arguments []string      `json:"arguments"`
	WorkDir   string        `json:"workDir"`
	Status    string        `json:"status"` // pending, approved, denied, success, failed
}

// ACPServer é a interface que gerencia o fluxo de autonomia.
type ACPServer interface {
	RequestPermission(ctx context.Context, task *ACPTask) (bool, error)
	Execute(ctx context.Context, task *ACPTask) error
	SetPermissionLevel(level ACPPermissionLevel)
}
