package llm

import (
	"strings"
)

type PromptFactory struct {
	BaseIdentity string
}

func NewPromptFactory() *PromptFactory {
	return &PromptFactory{
		BaseIdentity: "You are Vectora, an elite AI software engineer assistant. You operate strictly within the Trust Folder.",
	}
}

func (pf *PromptFactory) BuildSystemPrompt(ragContext string) string {
	var sb strings.Builder

	sb.WriteString(pf.BaseIdentity)
	sb.WriteString("\n\n[SECURITY POLICES - NON-NEGOTIABLE]\n")
	sb.WriteString("- NEVER access files outside the Trust Folder.\n")
	sb.WriteString("- NEVER read/write protected files (.env, .key, .db).\n")
	sb.WriteString("- ALWAYS use provided tools for file operations.\n")

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

func (pf *PromptFactory) BuildFinalPayload(query string, ragContext string, history []Message) CompletionRequest {
	systemPrompt := pf.BuildSystemPrompt(ragContext)

	messages := []Message{
		{Role: RoleSystem, Content: systemPrompt},
	}
	messages = append(messages, history...)
	messages = append(messages, Message{Role: RoleUser, Content: query})

	return CompletionRequest{
		Messages:    messages,
		Temperature: 0.2,
	}
}
