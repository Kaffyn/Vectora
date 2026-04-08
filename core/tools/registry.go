package tools

import (
	"vectora/core/git"
	"vectora/core/policies"
	"vectora/core/storage"
)

type Registry struct {
	Tools       map[string]Tool
	Guardian    *policies.Guardian
	Storage     *storage.Engine
	GitManager  *git.Manager
	TrustFolder string
}

func NewRegistry(trustFolder string, guardian *policies.Guardian, storage *storage.Engine, gitMgr *git.Manager) *Registry {
	r := &Registry{
		Tools:       make(map[string]Tool),
		Guardian:    guardian,
		Storage:     storage,
		GitManager:  gitMgr,
		TrustFolder: trustFolder,
	}

	// Registro das Ferramentas MVP
	r.Register(&ReadFileTool{TrustFolder: trustFolder, Guardian: guardian})
	r.Register(&GrepSearchTool{TrustFolder: trustFolder, Guardian: guardian})
	r.Register(&TerminalRunTool{TrustFolder: trustFolder})
	// Adicionar WriteFile, Edit, etc.

	return r
}

func (r *Registry) Register(t Tool) {
	r.Tools[t.Name()] = t
}

func (r *Registry) GetTool(name string) (Tool, bool) {
	t, ok := r.Tools[name]
	return t, ok
}
