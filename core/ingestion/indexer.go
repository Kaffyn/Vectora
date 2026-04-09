package ingestion

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/Kaffyn/Vectora/core/policies"
)

type Indexer struct {
	LLM      llm.Provider
	Guardian *policies.Guardian
	Parser   *ParserSelector
	Graph    *DependencyGraph
}

func NewIndexer(provider llm.Provider, guardian *policies.Guardian) *Indexer {
	return &Indexer{
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

		fileCount++
		return nil
	})

	if err != nil {
		return err
	}

	fmt.Printf("Parsing complete: %d files in %v\n", fileCount, time.Since(startTime))
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
