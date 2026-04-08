package ingestion

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Kaffyn/Vectora/core/policies"
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
