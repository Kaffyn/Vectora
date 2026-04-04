package acp

import (
	"github.com/Kaffyn/Vectora/internal/db"
	"github.com/Kaffyn/Vectora/internal/tools"
)

// AgentContext acts as a conductor over the tools that the AI can perceive.
// In the future, ACP calls may run sandboxed.
type AgentContext struct {
	Registry *tools.Registry
}

func NewAgent(kvStore db.KVStore) *AgentContext {
	reg := tools.NewRegistry()

	// Load Base Permissions Kit
	reg.Register(&tools.ReadFileTool{})
	reg.Register(&tools.WriteFileTool{})
	reg.Register(&tools.ReadFolderTool{})
	reg.Register(&tools.EditTool{})

	reg.Register(&tools.FindFilesTool{})
	reg.Register(&tools.GrepSearchTool{})
	reg.Register(&tools.ShellTool{})

	reg.Register(&tools.SaveMemoryTool{KV: kvStore}) // Fixado: Dependencia Injetada Oficialmente!
	reg.Register(&tools.PlanModeTool{})

	reg.Register(&tools.GoogleSearchTool{})
	reg.Register(&tools.WebFetchTool{})

	return &AgentContext{
		Registry: reg,
	}
}
