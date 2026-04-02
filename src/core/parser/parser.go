package parser

import (
	"encoding/xml"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

type TechnicalParser struct{}

func NewTechnicalParser() *TechnicalParser {
	return &TechnicalParser{}
}

func (p *TechnicalParser) Parse(doc *domain.Document) ([]*domain.Chunk, error) {
	ext := filepath.Ext(doc.FilePath)

	switch strings.ToLower(ext) {
	case ".gd", ".cpp", ".h", ".cs", ".go":
		return p.parseCode(doc)
	case ".md":
		return p.parseMarkdown(doc)
	case ".xml":
		return p.parseEngineDocs(doc)
	default:
		return p.parseText(doc)
	}
}

func (p *TechnicalParser) parseCode(doc *domain.Document) ([]*domain.Chunk, error) {
	ext := filepath.Ext(doc.FilePath)
	if ext == ".gd" {
		return p.parseGDScript(doc)
	}

	// Default simple split by double newline for other code files
	blocks := strings.Split(doc.Content, "\n\n")
	var chunks []*domain.Chunk
	for i, block := range blocks {
		content := strings.TrimSpace(block)
		if content == "" {
			continue
		}
		chunks = append(chunks, &domain.Chunk{
			ID:         fmt.Sprintf("%s_c%d", doc.ID, i),
			DocumentID: doc.ID,
			Content:    content,
			Index:      i,
		})
	}
	return chunks, nil
}

func (p *TechnicalParser) parseGDScript(doc *domain.Document) ([]*domain.Chunk, error) {
	lines := strings.Split(doc.Content, "\n")
	var chunks []*domain.Chunk
	var currentBlock strings.Builder
	idx := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Início de nova função ou sinal/variável significativa
		if strings.HasPrefix(trimmed, "func ") || strings.HasPrefix(trimmed, "signal ") || strings.HasPrefix(trimmed, "class_name ") {
			// Salva o bloco anterior se existir
			if currentBlock.Len() > 0 {
				chunks = append(chunks, &domain.Chunk{
					ID:         fmt.Sprintf("%s_gd%d", doc.ID, idx),
					DocumentID: doc.ID,
					Content:    strings.TrimSpace(currentBlock.String()),
					Index:      idx,
				})
				idx++
				currentBlock.Reset()
			}
		}
		currentBlock.WriteString(line + "\n")
	}

	// Último bloco
	if currentBlock.Len() > 0 {
		chunks = append(chunks, &domain.Chunk{
			ID:         fmt.Sprintf("%s_gd%d", doc.ID, idx),
			DocumentID: doc.ID,
			Content:    strings.TrimSpace(currentBlock.String()),
			Index:      idx,
		})
	}

	return chunks, nil
}

func (p *TechnicalParser) parseMarkdown(doc *domain.Document) ([]*domain.Chunk, error) {
	// Split by headers (#)
	sections := strings.Split(doc.Content, "\n#")
	var chunks []*domain.Chunk
	for i, section := range sections {
		content := strings.TrimSpace(section)
		if content == "" {
			continue
		}
		if i > 0 {
			content = "#" + content // Restore header marker
		}
		chunks = append(chunks, &domain.Chunk{
			ID:         fmt.Sprintf("%s_s%d", doc.ID, i),
			DocumentID: doc.ID,
			Content:    content,
			Index:      i,
		})
	}
	return chunks, nil
}

func (p *TechnicalParser) parseEngineDocs(doc *domain.Document) ([]*domain.Chunk, error) {
	// Estruturas para unmarshal do XML parcial da Godot
	type Method struct {
		Name        string `xml:"name,attr"`
		Description string `xml:"description"`
	}
	type Member struct {
		Name        string `xml:"name,attr"`
		Description string `xml:",chardata"`
	}
	type GodotClass struct {
		Name             string   `xml:"name,attr"`
		Inherits         string   `xml:"inherits,attr"`
		BriefDescription string   `xml:"brief_description"`
		Description      string   `xml:"description"`
		Methods          []Method `xml:"methods>method"`
		Members          []Member `xml:"members>member"`
	}

	var gc GodotClass
	if err := xml.Unmarshal([]byte(doc.Content), &gc); err != nil {
		// Se falhar o XML, tenta texto puro
		return p.parseText(doc)
	}

	var chunks []*domain.Chunk
	idx := 0

	// 1. Chunk da Classe
	if gc.Name != "" {
		content := fmt.Sprintf("Class: %s\nInherits: %s\nSummary: %s\nDetails: %s", gc.Name, gc.Inherits, gc.BriefDescription, gc.Description)
		chunks = append(chunks, &domain.Chunk{
			ID:         fmt.Sprintf("%s_class", doc.ID),
			DocumentID: doc.ID,
			Content:    content,
			Index:      idx,
		})
		idx++
	}

	// 2. Chunks de Métodos
	for _, m := range gc.Methods {
		content := fmt.Sprintf("Method: %s::%s\nDescription: %s", gc.Name, m.Name, strings.TrimSpace(m.Description))
		chunks = append(chunks, &domain.Chunk{
			ID:         fmt.Sprintf("%s_method_%s", doc.ID, m.Name),
			DocumentID: doc.ID,
			Content:    content,
			Index:      idx,
		})
		idx++
	}

	// 3. Chunks de Membros/Propriedades
	for _, mb := range gc.Members {
		content := fmt.Sprintf("Property: %s::%s\nDetails: %s", gc.Name, mb.Name, strings.TrimSpace(mb.Description))
		chunks = append(chunks, &domain.Chunk{
			ID:         fmt.Sprintf("%s_prop_%s", doc.ID, mb.Name),
			DocumentID: doc.ID,
			Content:    content,
			Index:      idx,
		})
		idx++
	}

	return chunks, nil
}

func (p *TechnicalParser) parseText(doc *domain.Document) ([]*domain.Chunk, error) {
	// Simple paragraph split
	paragraphs := strings.Split(doc.Content, "\n\n")
	var chunks []*domain.Chunk
	for i, para := range paragraphs {
		content := strings.TrimSpace(para)
		if content == "" {
			continue
		}
		chunks = append(chunks, &domain.Chunk{
			ID:         fmt.Sprintf("%s_p%d", doc.ID, i),
			DocumentID: doc.ID,
			Content:    content,
			Index:      i,
		})
	}
	return chunks, nil
}
