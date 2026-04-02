package tool

import (
	"sync"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

// Registry mantém o catálogo de todas as ferramentas disponíveis no Vectora.
type Registry struct {
	tools map[string]Tool
	mu    sync.RWMutex
}

// NewRegistry cria um novo catálogo pronto para receber ferramentas.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adiciona uma nova ferramenta ao catálogo.
func (r *Registry) Register(t Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[t.Name()] = t
}

// Get retorna uma ferramenta pelo seu nome identificador.
func (r *Registry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

// List retorna todas as ferramentas registradas.
func (r *Registry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		list = append(list, t)
	}
	return list
}

// InitDefaultTools popula o catálogo com as ferramentas padrão de sistema e busca.
func (r *Registry) InitDefaultTools(osm domain.OSManager, repo domain.ChunkRepository, embedder domain.EmbeddingProvider) {
	r.Register(NewReadFileTool())
	r.Register(NewWriteFileTool())
	r.Register(NewReadFolderTool())
	r.Register(NewEditTool())
	r.Register(NewShellTool(osm))
	r.Register(NewFindFilesTool())
	r.Register(NewGrepSearchTool())
	r.Register(NewSaveMemoryTool(repo, embedder))
	r.Register(NewGoogleSearchTool("")) // API Key vazia por enquanto
	r.Register(NewWebFetchTool())
	r.Register(NewEnterPlanModeTool())
}
