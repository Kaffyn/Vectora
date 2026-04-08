package llm

import (
	"fmt"

	"github.com/pkoukk/tiktoken-go"
)

const (
	SafetyMargin         = 500  // Margem de segurança em tokens para evitar off-by-one errors
	SystemPromptOverhead = 1000 // Reserva para system prompt e policies
)

type ContextManager struct {
	encoding         *tiktoken.Tiktoken
	MaxContextTokens int
}

func NewContextManager(modelName string, maxTokens int) (*ContextManager, error) {
	enc, err := tiktoken.EncodingForModel(modelName)
	if err != nil {
		// Fallback para cl100k_base (usado por GPT-4 e Gemini)
		enc, err = tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			return nil, err
		}
	}
	return &ContextManager{encoding: enc, MaxContextTokens: maxTokens}, nil
}

// CountTokens calcula o número exato de tokens.
func (cm *ContextManager) CountTokens(text string) int {
	return len(cm.encoding.Encode(text, nil, nil))
}

// TrimMessages aplica a estratégia de truncagem agressiva.
// Prioridade: System Prompt > RAG Context > Recent History > Old History
func (cm *ContextManager) TrimMessages(systemPrompt string, ragContext string, history []Message) ([]Message, error) {
	availableTokens := cm.MaxContextTokens - SystemPromptOverhead - cm.CountTokens(systemPrompt) - cm.CountTokens(ragContext)

	if availableTokens <= 0 {
		return nil, fmt.Errorf("system prompt and RAG context exceed model limits")
	}

	var trimmedHistory []Message
	currentTokens := 0

	// Itera do mais recente para o mais antigo
	for i := len(history) - 1; i >= 0; i-- {
		msg := history[i]
		tokens := cm.CountTokens(msg.Content)

		if currentTokens+tokens > availableTokens {
			break // Corta o histórico antigo
		}

		trimmedHistory = append([]Message{msg}, trimmedHistory...)
		currentTokens += tokens
	}

	return trimmedHistory, nil
}
