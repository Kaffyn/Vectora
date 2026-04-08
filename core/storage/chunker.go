package storage

import (
	"strings"
)

// SimpleChunker divide o texto em blocos sobrepostos baseados em caracteres/tokens aproximados.
// Nota: Para MVP, usamos contagem de runes como proxy de tokens.
// Em produção, integraríamos com tiktoken para precisão absoluta.
func SimpleChunker(content string, config ChunkConfig) []string {
	if len(content) == 0 {
		return []string{}
	}

	var chunks []string
	runes := []rune(content)
	totalRunes := len(runes)

	// Ajuste simples: 1 token ~= 4 caracteres em média para código inglês/latino
	maxChars := config.MaxTokens * 4
	overlapChars := config.Overlap * 4

	if maxChars >= totalRunes {
		return []string{content}
	}

	start := 0
	for start < totalRunes {
		end := start + maxChars
		if end > totalRunes {
			end = totalRunes
		}

		// Tenta quebrar em linha nova para não cortar código no meio
		chunkEnd := end
		if end < totalRunes {
			// Procura a próxima newline após o limite para quebrar limpo
			nextNewline := strings.IndexRune(string(runes[end:]), '\n')
			if nextNewline != -1 && nextNewline < (maxChars/2) { // Não procure muito longe
				chunkEnd = end + nextNewline + 1
			}
		}

		chunk := string(runes[start:chunkEnd])
		chunks = append(chunks, chunk)

		// Move o start para trás do overlap para a "costura"
		start = chunkEnd - overlapChars
		if start < 0 {
			start = 0
		}

		// Segurança contra loop infinito se o overlap for maior que o chunk
		if start >= chunkEnd {
			break
		}
	}

	return chunks
}
