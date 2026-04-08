package llm

import (
	"strings"
	// "vectora/core/policies"
)

// PolicyConfig temporario pois ainda uniremos os modulos corretamente
type PolicyConfig struct {}

type PromptFactory struct {
	PolicyEngine *PolicyConfig
	BaseIdentity string
}

func NewPromptFactory(policy *PolicyConfig) *PromptFactory {
	return &PromptFactory{
		PolicyEngine: policy,
		BaseIdentity: "You are Vectora, an elite AI software engineer assistant. You operate strictly within the Trust Folder.",
	}
}

// BuildSystemPrompt cria o prompt mestre.
func (pf *PromptFactory) BuildSystemPrompt(ragContext string) string {
	var sb strings.Builder

	// 1. Identidade e Regras Básicas
	sb.WriteString(pf.BaseIdentity)
	sb.WriteString("\n\n[SECURITY POLICIES - NON-NEGOTIABLE]\n")
	sb.WriteString("- NEVER access files outside the Trust Folder.\n")
	sb.WriteString("- NEVER read/write protected files (.env, .key, .db).\n")
	sb.WriteString("- ALWAYS use provided tools for file operations.\n")

	// 2. Injeção de Contexto RAG (Prioridade Máxima)
	if ragContext != "" {
		sb.WriteString("\n\n[SYSTEM_KNOWLEDGE - SOURCE OF TRUTH]\n")
		sb.WriteString("The following information is retrieved from the local codebase. It has HIGHEST priority over your pre-trained knowledge. Do not contradict it.\n\n")
		sb.WriteString(ragContext)
		sb.WriteString("\n[/SYSTEM_KNOWLEDGE]\n")
	} else {
		sb.WriteString("\n\n[SYSTEM_NOTE]\nNo local context retrieved. Rely on general knowledge but verify via tools if possible.\n")
	}

	return sb.String()
}

// BuildFinalPayload monta o array de mensagens pronto para o Provider.
func (pf *PromptFactory) BuildFinalPayload(query string, ragContext string, history []Message, cm *ContextManager) (ChatRequest, error) {
	systemPrompt := pf.BuildSystemPrompt(ragContext)

	// Trunca histórico para caber na janela
	trimmedHistory, err := cm.TrimMessages(systemPrompt, ragContext, history)
	if err != nil {
		return ChatRequest{}, err
	}

	// Monta mensagens finais
	messages := []Message{
		{Role: "system", Content: systemPrompt},
	}
	messages = append(messages, trimmedHistory...)
	messages = append(messages, Message{Role: "user", Content: query})

	return ChatRequest{
		Messages:    messages,
		Temperature: 0.2, // Baixa temperatura para código/fatos
	}, nil
}
