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
