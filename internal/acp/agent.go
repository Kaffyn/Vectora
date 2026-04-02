package acp

import (
	"github.com/Kaffyn/vectora/internal/tools"
)

// AgentContext atua como maestro sobre as ferramentas que a IA poderá visualizar.
// Em um futuro as chamadas de ACP poderão rodar sandboxed.
type AgentContext struct {
	Registry *tools.Registry
}

func NewAgent() *AgentContext {
	reg := tools.NewRegistry()
	
	// Carrega Kit Base de Permissões
	reg.Register(&tools.ReadFileTool{})
	reg.Register(&tools.WriteFileTool{})
	reg.Register(&tools.ReadFolderTool{})
	reg.Register(&tools.EditTool{})
	
	reg.Register(&tools.FindFilesTool{})
	reg.Register(&tools.GrepSearchTool{})
	reg.Register(&tools.ShellTool{})
	
	reg.Register(&tools.SaveMemoryTool{KV: nil}) // FIXME: Injete dinamicamente depois
	reg.Register(&tools.PlanModeTool{})
	
	reg.Register(&tools.GoogleSearchTool{})
	reg.Register(&tools.WebFetchTool{})

	return &AgentContext{
		Registry: reg,
	}
}
