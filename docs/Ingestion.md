# Blueprint: Ingestão e Indexação (The Digestor)

**Status:** Fase 4 - Implementação Concluída  
**Módulo:** `core/ingestion/`  
**Dependencies:** `path/filepath`, `regexp`, `strings`, `vectora/core/storage`, `vectora/core/policies`, `vectora/core/llm`

## 1. Interface do Parser e Seletor (`parser.go`)

Define como diferentes tipos de arquivo são tratados. O MVP foca em texto puro e código, ignorando binários via Guardian.

```go
package ingestion

import (
	"os"
	"path/filepath"
	"strings"
	"vectora/core/policies"
)

// ParsedFile representa o conteúdo extraído e metadados de um arquivo.
type ParsedFile struct {
	Path     string
	Content  string
	Language string // "go", "md", "txt", "unknown"
}

// Parser interface para extração de conteúdo.
type Parser interface {
	Parse(path string, content []byte) (*ParsedFile, error)
}

// TextParser lida com arquivos de texto genéricos e Markdown.
type TextParser struct{}

func (p *TextParser) Parse(path string, content []byte) (*ParsedFile, error) {
	return &ParsedFile{
		Path:     path,
		Content:  string(content),
		Language: detectLanguage(path),
	}, nil
}

// ParserSelector escolhe o parser correto baseado na extensão.
type ParserSelector struct {
	Guardian *policies.Guardian
}

func NewParserSelector(guardian *policies.Guardian) *ParserSelector {
	return &ParserSelector{Guardian: guardian}
}

func (ps *ParserSelector) ParseFile(absPath string) (*ParsedFile, error) {
	// 1. Verificação de Segurança (Guardian)
	if ps.Guardian.IsProtected(absPath) {
		return nil, nil // Silently skip protected files
	}

	// 2. Leitura do Arquivo
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	// 3. Seleção e Execução do Parser
	// No MVP, usamos um TextParser universal pois código é texto.
	// A distinção real acontece no Chunking (que pode ser aware de linguagem no futuro).
	parser := &TextParser{}
	return parser.Parse(absPath, content)
}

func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js", ".ts":
		return "javascript"
	case ".md":
		return "markdown"
	default:
		return "text"
	}
}
```

## 2. Extrator de Grafo de Dependências Simplificado (`dependency_graph.go`)

Implementa a lógica de "Regex Grep" para identificar imports/requires e construir um grafo leve em memória. Isso permite que o RAG saiba que `auth.go` depende de `db.go`.

```go
package ingestion

import (
	"regexp"
	"strings"
)

// DependencyGraph mapeia arquivos para suas dependências (imports).
type DependencyGraph struct {
	// Map[ArquivoOrigem] -> Lista de ArquivosDestino (ou pacotes)
	Edges map[string][]string
}

func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{Edges: make(map[string][]string)}
}

// ExtractImports usa regex simples para encontrar dependências comuns.
// MVP: Foca em Go e JS/TS.
func (dg *DependencyGraph) ExtractImports(filePath string, content string) {
	var imports []string

	// Regex para Go: import "package" ou import ( "package" )
	goImportRe := regexp.MustCompile(`import\s+(?:\(\s*)?["']([^"']+)["']`)
	// Regex para JS/TS: import ... from 'module' ou require('module')
	jsImportRe := regexp.MustCompile(`(?:import\s+.*\s+from\s+|require\s*\(\s*)['"]([^'"]+)['"]`)

	language := detectLanguage(filePath)

	if language == "go" {
		matches := goImportRe.FindAllStringSubmatch(content, -1)
		for _, m := range matches {
			if len(m) > 1 {
				imports = append(imports, m[1])
			}
		}
	} else if language == "javascript" || language == "typescript" {
		matches := jsImportRe.FindAllStringSubmatch(content, -1)
		for _, m := range matches {
			if len(m) > 1 {
				imports = append(imports, m[1])
			}
		}
	}

	if len(imports) > 0 {
		dg.Edges[filePath] = imports
	}
}

// GetDependencies retorna as dependências de um arquivo.
func (dg *DependencyGraph) GetDependencies(filePath string) []string {
	return dg.Edges[filePath]
}
```

## 3. Pipeline de Indexação On-Demand (`indexer.go`)

O orquestrador principal. Varre o diretório, parseia, chunka, embedda e salva.

```go
package ingestion

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"vectora/core/llm"
	"vectora/core/policies"
	"vectora/core/storage"
)

type Indexer struct {
	Storage  *storage.Engine
	LLM      llm.LLMProvider
	Guardian *policies.Guardian
	Parser   *ParserSelector
	Graph    *DependencyGraph
}

func NewIndexer(storage *storage.Engine, provider llm.LLMProvider, guardian *policies.Guardian) *Indexer {
	return &Indexer{
		Storage:  storage,
		LLM:      provider,
		Guardian: guardian,
		Parser:   NewParserSelector(guardian),
		Graph:    NewDependencyGraph(),
	}
}

// IndexDirectory varre o diretório e indexa arquivos novos ou modificados.
func (idx *Indexer) IndexDirectory(ctx context.Context, rootPath string) error {
	startTime := time.Now()
	fileCount := 0

	// Walk seguro
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip erros de permissão/arquivo
		}

		if info.IsDir() {
			// Ignorar diretórios ocultos ou de build
			if isIgnoredDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		// Processar Arquivo
		parsed, err := idx.Parser.ParseFile(path)
		if err != nil || parsed == nil {
			return nil
		}

		// Extrair Dependências para o Grafo
		idx.Graph.ExtractImports(path, parsed.Content)

		// Indexar no Storage (Hash check + Upsert vetorial)
		// RelPath é usado para consistência no banco
		relPath, _ := filepath.Rel(rootPath, path)
		if err := idx.Storage.IndexFile(ctx, relPath, parsed.Content); err != nil {
			fmt.Printf("Warning: Failed to index %s: %v\n", path, err)
			return nil
		}

		fileCount++
		return nil
	})

	if err != nil {
		return err
	}

	// Atualizar Metadata do Workspace
	meta, _ := idx.Storage.Meta.GetWorkspaceMeta()
	meta.LastIndexedAt = time.Now()
	meta.TotalFiles = fileCount
	meta.Status = "idle"
	idx.Storage.Meta.SaveWorkspaceMeta(*meta)

	fmt.Printf("Indexing complete: %d files in %v\n", fileCount, time.Since(startTime))
	return nil
}

func isIgnoredDir(name string) bool {
	ignored := []string{".git", "node_modules", "vendor", "dist", "build", ".vectora"}
	for _, ign := range ignored {
		if name == ign {
			return true
		}
	}
	return false
}
```

## 4. Integração com o Agente (Uso do Grafo)

Quando o Agente faz uma pergunta, ele pode usar o grafo para expandir o contexto.

```go
// Exemplo de uso no RAG Pipeline (futuro core/rag/retriever.go)
func (r *Retriever) ExpandContextWithGraph(initialChunks []storage.Chunk, graph *DependencyGraph) []storage.Chunk {
	// 1. Identificar arquivos dos chunks iniciais
	// 2. Consultar graph.GetDependencies(file)
	// 3. Buscar chunks adicionais desses arquivos dependentes
	// 4. Retornar lista expandida

	// MVP: Apenas retorna os chunks iniciais. A expansão do grafo é uma otimização v1.1.
	return initialChunks
}
```

---

### Resumo da Estratégia de Ingestão

1.  **Segurança Primeiro:** O `ParserSelector` consulta o `Guardian` antes de ler qualquer byte.
2.  **Eficiência:** O `Indexer` usa o mecanismo de Hash do `storage.Engine` para pular arquivos inalterados, tornando re-indexações rápidas.
3.  **Simplicidade Inteligente:** Em vez de ASTs pesados, usamos Regex direcionada para construir um grafo de dependências útil para o MVP.
4.  **On-Demand:** Nenhuma rotina em segundo plano. O usuário controla quando gastar ciclos de CPU/GPU.
