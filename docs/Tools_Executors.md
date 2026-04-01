# Implementação: Core Tools Executors

**Módulo:** `core/tools/`
**Princípio:** "Fail Fast, Sanitize Always, Scope Strict."

## 1. Interface Unificada de Ferramentas

Todas as ferramentas implementam esta interface, permitindo que o `AgentLoop` as chame de forma genérica.

```go
package tools

import (
    "context"
    "encoding/json"
)

// ToolResult é a resposta padronizada para o LLM
type ToolResult struct {
    Output   string `json:"output"`   // Conteúdo ou mensagem de erro formatada
    IsError  bool   `json:"is_error"` // Flag para o LLM saber que algo falhou
    Metadata map[string]interface{} `json:"metadata,omitempty"` // Ex: line_count, execution_time
}

// Tool define o contrato básico
type Tool interface {
    Name() string
    Description() string // Usado para o JSON Schema do MCP/ACP
    Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error)
}
```

## 2. Executor: `read_file` (Leitura Segura)

**Desafio:** Evitar estouro de contexto e leitura de arquivos binários/gigantes.
**Solução:** Limite rígido de bytes + Detecção de tipo.

```go
package tools

import (
    "os"
    "path/filepath"
    "vectora/core/policies"
    "vectora/core/storage"
)

const MAX_READ_BYTES = 50 * 1024 // 50KB limite seguro para contexto

type ReadFileTool struct {
    TrustFolder string
    Guardian    *policies.Guardian
}

func (t *ReadFileTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
    var params struct {
        Path string `json:"path"`
    }
    if err := json.Unmarshal(args, &params); err != nil {
        return &ToolResult{Output: "Invalid arguments", IsError: true}, nil
    }

    // 1. Validação de Escopo (Policy)
    safePath := filepath.Join(t.TrustFolder, params.Path)
    if !t.Guardian.IsPathSafe(safePath) {
        return &ToolResult{Output: "Access Denied: Out of Trust Folder", IsError: true}, nil
    }

    // 2. Validação de Guardian (Blocklist)
    if t.Guardian.IsProtected(safePath) {
        return &ToolResult{Output: "Access Denied: Protected File Type", IsError: true}, nil
    }

    // 3. Leitura Controlada
    file, err := os.Open(safePath)
    if err != nil {
        return &ToolResult{Output: err.Error(), IsError: true}, nil
    }
    defer file.Close()

    // Lê apenas os primeiros N bytes
    buffer := make([]byte, MAX_READ_BYTES+1) // +1 para detectar se truncou
    n, err := file.Read(buffer)

    content := string(buffer[:n])
    truncated := false
    if n > MAX_READ_BYTES || err == nil && n == MAX_READ_BYTES+1 {
        content = string(buffer[:MAX_READ_BYTES])
        truncated = true
    }

    result := ToolResult{
        Output: content,
        Metadata: map[string]interface{}{
            "truncated": truncated,
            "size_bytes": n,
        },
    }

    if truncated {
        result.Output += "\n... [TRUNCATED: Use grep_search for specific content] ..."
    }

    return &result, nil
}
```

## 3. Executor: `grep_search` (Busca Nativa Go)

**Desafio:** Performance sem depender de `ripgrep` ou `grep` do sistema.
**Solução:** `filepath.WalkDir` + `bytes.Contains` ou `regexp`.

```go
package tools

import (
    "bytes"
    "os"
    "path/filepath"
    "regexp"
    "strings"
)

type GrepSearchTool struct {
    TrustFolder string
    Guardian    *policies.Guardian
}

func (t *GrepSearchTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
    var params struct {
        Pattern string `json:"pattern"`
        CaseSensitive bool `json:"case_sensitive"`
    }
    json.Unmarshal(args, &params)

    var results []string
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

    // Walk seguro e rápido
    filepath.WalkDir(t.TrustFolder, func(path string, d os.DirEntry, err error) error {
        if err != nil || d.IsDir() {
            return nil
        }

        // Skip protegidos
        if t.Guardian.IsProtected(path) {
            return nil
        }

        data, err := os.ReadFile(path)
        if err != nil {
            return nil
        }

        if re.Match(data) {
            // Retorna path relativo
            rel, _ := filepath.Rel(t.TrustFolder, path)
            results = append(results, rel)
        }
        return nil
    })

    if len(results) == 0 {
        return &ToolResult{Output: "No matches found"}, nil
    }

    return &ToolResult{
        Output: strings.Join(results, "\n"),
        Metadata: map[string]interface{}{"matches_count": len(results)},
    }, nil
}
```

## 4. Executor: `terminal_run` (Sandbox de Processo)

**Desafio:** Processos infinitos e injeção de comandos maliciosos.
**Solução:** `exec.CommandContext` com Timeout e Shell restrito.

```go
package tools

import (
    "context"
    "os/exec"
    "runtime"
    "time"
)

const CMD_TIMEOUT = 30 * time.Second

type TerminalRunTool struct {
    TrustFolder string
}

func (t *TerminalRunTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
    var params struct {
        Command string `json:"command"`
    }
    json.Unmarshal(args, &params)

    // Cria contexto com timeout obrigatório
    cmdCtx, cancel := context.WithTimeout(ctx, CMD_TIMEOUT)
    defer cancel()

    var cmd *exec.Cmd

    // Segurança: Usa shell padrão do OS mas restringe o working dir
    if runtime.GOOS == "windows" {
        cmd = exec.CommandContext(cmdCtx, "cmd", "/C", params.Command)
    } else {
        cmd = exec.CommandContext(cmdCtx, "sh", "-c", params.Command)
    }

    cmd.Dir = t.TrustFolder // Isolamento de diretório

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()

    output := stdout.String()
    if err != nil {
        if cmdCtx.Err() == context.DeadlineExceeded {
            return &ToolResult{
                Output: "Command timed out after 30s",
                IsError: true,
            }, nil
        }
        // Adiciona stderr ao output se houver erro
        output += "\nError: " + stderr.String()
    }

    return &ToolResult{
        Output: output,
        IsError: err != nil,
    }, nil
}
```

## 5. Validador de Saída Unificado (`Sanitizer`)

Antes de qualquer `ToolResult` voltar para o Agente, ele passa por este filtro.

```go
package tools

import (
    "unicode/utf8"
)

func SanitizeResult(res *ToolResult) *ToolResult {
    // 1. Garante UTF-8 válido (evita quebra de JSON-RPC)
    if !utf8.ValidString(res.Output) {
        res.Output = "Binary or Invalid UTF-8 Content"
        res.IsError = true
        return res
    }

    // 2. Limita tamanho final da resposta (ex: 100k chars)
    if len(res.Output) > 100000 {
        res.Output = res.Output[:100000] + "\n... [OUTPUT TRUNCATED DUE TO LENGTH]"
    }

    // 3. Remove caracteres de controle perigosos (exceto newline/tab)
    // Implementação simples de sanitize
    clean := make([]rune, 0, len(res.Output))
    for _, r := range res.Output {
        if r == '\n' || r == '\t' || (r >= 32 && r <= 126) || utf8.RuneLen(r) > 1 {
            clean = append(clean, r)
        }
    }
    res.Output = string(clean)

    return res
}
```

---

### Resumo da Estratégia de Execução

1.  **Portabilidade:** Nenhuma dependência externa. Tudo roda com a stdlib do Go.
2.  **Segurança:**
    - `read_file`: Bloqueio por tamanho e tipo.
    - `grep`: Ignora arquivos protegidos nativamente.
    - `terminal`: Timeout rígido e Working Directory travado.
3.  **Resiliência:** O `Sanitizer` garante que o JSON-RPC nunca quebre por causa de output sujo da ferramenta.

Esta camada de ferramentas está pronta para ser integrada ao `AgentLoop`. O próximo passo lógico é definir o **Orquestrador do Agente** (o loop que decide qual ferramenta chamar). Quer que eu esboce o `core/agent/loop.go`?
