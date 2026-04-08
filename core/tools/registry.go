package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/policies"
)

type Registry struct {
	Tools       map[string]Tool
	Guardian    *policies.Guardian
	TrustFolder string
	KV          db.KVStore
}

func NewRegistry(trustFolder string, guardian *policies.Guardian, kv db.KVStore) *Registry {
	r := &Registry{
		Tools:       make(map[string]Tool),
		Guardian:    guardian,
		TrustFolder: trustFolder,
		KV:          kv,
	}

	r.Register(&ReadFileTool{TrustFolder: trustFolder, Guardian: guardian})
	r.Register(&WriteFileTool{TrustFolder: trustFolder, Guardian: guardian})
	r.Register(&ReadFolderTool{TrustFolder: trustFolder, Guardian: guardian})
	r.Register(&EditTool{TrustFolder: trustFolder, Guardian: guardian})
	r.Register(&FindFilesTool{TrustFolder: trustFolder, Guardian: guardian})
	r.Register(&GrepSearchTool{TrustFolder: trustFolder, Guardian: guardian})
	r.Register(&ShellTool{TrustFolder: trustFolder})
	r.Register(&SaveMemoryTool{TrustFolder: trustFolder, Guardian: guardian, KV: kv})
	r.Register(&GoogleSearchTool{})
	r.Register(&WebFetchTool{})

	return r
}

func (r *Registry) Register(t Tool) {
	r.Tools[t.Name()] = t
}

func (r *Registry) GetTool(name string) (Tool, bool) {
	t, ok := r.Tools[name]
	return t, ok
}

func (r *Registry) GetAll() []Tool {
	var all []Tool
	for _, t := range r.Tools {
		all = append(all, t)
	}
	return all
}

func (r *Registry) ExecuteStringArgs(ctx context.Context, name string, argsJSON string) (*ToolResult, error) {
	t, exists := r.Tools[name]
	if !exists {
		return &ToolResult{IsError: true, Output: fmt.Sprintf("tool '%s' not found", name)}, nil
	}
	return t.Execute(ctx, json.RawMessage(argsJSON))
}
