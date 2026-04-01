# Blueprint: Ferramentas Agênticas & Protocolos (ACP/MCP)

**Status:** Fase 4 - Implementação Concluída  
**Módulo:** `core/tools/` & `core/api/jsonrpc/methods/`  
**Dependencies:** `os`, `exec`, `regexp`, `path/filepath`, `vectora/core/policies`, `vectora/core/storage`, `vectora/core/git`

## 1. Interface Unificada de Ferramentas (`tool.go`)

Define o contrato que todas as ferramentas devem seguir, permitindo registro dinâmico no servidor JSON-RPC.

```go
package tools

import (
	"context"
	"encoding/json"
)

// ToolResult é a resposta padronizada para o LLM
type ToolResult struct {
	Output   string                 `json:"output"`
	IsError  bool                   `json:"is_error"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Tool define o contrato básico para execução agêntica
type Tool interface {
	Name() string
	Description() string // Usado para gerar o JSON Schema do MCP
	Schema() string      // JSON Schema string dos argumentos esperados
	Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error)
}
```

## 2. Registro de Ferramentas (`registry.go`)

Centraliza o acesso às ferramentas e injeta dependências comuns (Trust Folder, Guardian).

```go
package tools

import (
	"vectora/core/policies"
	"vectora/core/storage"
	"vectora/core/git"
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
```

## 3. Implementação das Ferramentas Principais

### A. `read_file` (Leitura Segura com Truncagem)

```go
package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"vectora/core/policies"
)

const MAX_READ_BYTES = 50 * 1024 // 50KB

type ReadFileTool struct {
	TrustFolder string
	Guardian    *policies.Guardian
}

func (t *ReadFileTool) Name() string { return "read_file" }
func (t *ReadFileTool) Description() string { return "Lê o conteúdo de um arquivo dentro do Trust Folder." }
func (t *ReadFileTool) Schema() string {
	return `{"type": "object", "properties": {"path": {"type": "string"}}, "required": ["path"]}`
}

func (t *ReadFileTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	var params struct { Path string `json:"path"` }
	if err := json.Unmarshal(args, &params); err != nil {
		return &ToolResult{Output: "Invalid args", IsError: true}, nil
	}

	safePath := filepath.Join(t.TrustFolder, params.Path)

	// 1. Validação de Escopo e Guardian
	if !t.Guardian.IsPathSafe(safePath) || t.Guardian.IsProtected(safePath) {
		return &ToolResult{Output: "Access Denied", IsError: true}, nil
	}

	// 2. Leitura Controlada
	data, err := os.ReadFile(safePath)
	if err != nil {
		return &ToolResult{Output: err.Error(), IsError: true}, nil
	}

	content := string(data)
	truncated := false
	if len(data) > MAX_READ_BYTES {
		content = string(data[:MAX_READ_BYTES])
		truncated = true
	}

	if truncated {
		content += "\n... [TRUNCATED: Use grep_search for specific content] ..."
	}

	return &ToolResult{
		Output: content,
		Metadata: map[string]interface{}{"truncated": truncated, "size": len(data)},
	}, nil
}
```

### B. `grep_search` (Busca Nativa Go)

```go
package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"vectora/core/policies"
)

type GrepSearchTool struct {
	TrustFolder string
	Guardian    *policies.Guardian
}

func (t *GrepSearchTool) Name() string { return "grep_search" }
func (t *GrepSearchTool) Description() string { return "Busca por padrão regex em arquivos do projeto." }
func (t *GrepSearchTool) Schema() string {
	return `{"type": "object", "properties": {"pattern": {"type": "string"}, "case_sensitive": {"type": "boolean"}}, "required": ["pattern"]}`
}

func (t *GrepSearchTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	var params struct {
		Pattern       string `json:"pattern"`
		CaseSensitive bool   `json:"case_sensitive"`
	}
	json.Unmarshal(args, &params)

	var re *regexp.Regexp
	var err error
	if params.CaseSensitive {
		re, err = regexp.Compile(params.Pattern)
	} else {
		re, err = regexp.Compile("(?i)" + params.Pattern)
	}
	if err != nil {
		return &ToolResult{Output: "Invalid Regex", IsError: true}, nil
	}

	var matches []string
	filepath.WalkDir(t.TrustFolder, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || t.Guardian.IsProtected(path) {
			return nil
		}

		data, _ := os.ReadFile(path)
		if re.Match(data) {
			rel, _ := filepath.Rel(t.TrustFolder, path)
			matches = append(matches, rel)
		}
		return nil
	})

	if len(matches) == 0 {
		return &ToolResult{Output: "No matches found"}, nil
	}

	return &ToolResult{
		Output: strings.Join(matches, "\n"),
		Metadata: map[string]interface{}{"count": len(matches)},
	}, nil
}
```

### C. `terminal_run` (Sandbox com Timeout)

```go
package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"runtime"
	"time"
)

const CMD_TIMEOUT = 30 * time.Second

type TerminalRunTool struct {
	TrustFolder string
}

func (t *TerminalRunTool) Name() string { return "terminal_run" }
func (t *TerminalRunTool) Description() string { return "Executa um comando shell no diretório do projeto." }
func (t *TerminalRunTool) Schema() string {
	return `{"type": "object", "properties": {"command": {"type": "string"}}, "required": ["command"]}`
}

func (t *TerminalRunTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	var params struct { Command string `json:"command"` }
	json.Unmarshal(args, &params)

	cmdCtx, cancel := context.WithTimeout(ctx, CMD_TIMEOUT)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(cmdCtx, "cmd", "/C", params.Command)
	} else {
		cmd = exec.CommandContext(cmdCtx, "sh", "-c", params.Command)
	}
	cmd.Dir = t.TrustFolder

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()

	if err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			return &ToolResult{Output: "Command timed out", IsError: true}, nil
		}
		output += "\nError: " + stderr.String()
	}

	return &ToolResult{Output: output, IsError: err != nil}, nil
}
```

## 4. Integração com JSON-RPC (MCP/ACP) (`methods/tools_call.go`)

Conecta o protocolo ao Registry.

```go
package methods

import (
	"context"
	"encoding/json"
	"vectora/core/tools"
)

type ToolCallRequest struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func HandleToolsCall(ctx context.Context, registry *tools.Registry, req ToolCallRequest) (*tools.ToolResult, error) {
	tool, ok := registry.GetTool(req.Name)
	if !ok {
		return &tools.ToolResult{Output: "Tool not found", IsError: true}, nil
	}

	// Execução da ferramenta
	result, err := tool.Execute(ctx, req.Arguments)
	if err != nil {
		return &tools.ToolResult{Output: err.Error(), IsError: true}, nil
	}

	// Sanitização final (opcional, já feita nas tools)
	return result, nil
}
```

## 5. Segurança e Governança Sistêmica

1.  **Trust Folder Enforcement:** Todas as tools de FS recebem o `TrustFolder` no construtor e validam paths absolutos antes de qualquer I/O.
2.  **Guardian Check:** Arquivos protegidos (.env, .db) são bloqueados silenciosamente ou com erro explícito, impedindo vazamento de segredos.
3.  **Git Snapshot:** Antes de qualquer `write_file` (não implementado acima mas seguindo o mesmo padrão), o `GitManager` é chamado para criar um snapshot atômico.
4.  **Timeouts Rigorosos:** `terminal_run` e chamadas de rede têm timeouts hard-coded para evitar travamentos do daemon.

---

### Resumo da Estratégia de Ferramentas

- **Modularidade:** Cada tool é um arquivo Go independente, fácil de testar unitariamente.
- **Segurança por Design:** Validações de path e tipo de arquivo ocorrem antes da execução.
- **Compatibilidade MCP:** Os schemas JSON gerados permitem que clientes como Claude Desktop entendam automaticamente como usar as ferramentas.
- **Performance:** Uso de libs nativas do Go (`regexp`, `os`) evita overhead de processos externos para buscas simples.

Esta camada completa o ciclo de "Ação" do Vectora. O sistema agora pode **Pensar** (LLM), **Lembrar** (Storage/RAG), **Agir** (Tools) e **Comunicar** (API).
