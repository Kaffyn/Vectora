package ingestion

import (
	"regexp"
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
