package llm

import (
	"os"
	"path/filepath"
	"strings"
)

type PromptFactory struct {
	BaseIdentity     string
	Language         string
	FullInstructions string // Loaded from instructions.md
}

func NewPromptFactory() *PromptFactory {
	pf := &PromptFactory{
		BaseIdentity: "You are Vectora, an elite AI software engineer assistant. You operate strictly within the Trust Folder.",
		Language:     "en",
	}

	// Try to load full instructions from file
	pf.loadInstructionsFromFile()

	return pf
}

func (pf *PromptFactory) loadInstructionsFromFile() {
	// Look for instructions.md in multiple possible locations
	possiblePaths := []string{
		"./core/instructions/instructions.md",
		"../instructions/instructions.md",
		"../../core/instructions/instructions.md",
	}

	// Also check relative to executable
	if exePath, err := os.Executable(); err == nil {
		possiblePaths = append(possiblePaths, filepath.Join(filepath.Dir(exePath), "instructions.md"))
	}

	for _, path := range possiblePaths {
		if content, err := os.ReadFile(path); err == nil {
			pf.FullInstructions = string(content)
			return
		}
	}
	// If no file found, FullInstructions stays empty and we use hardcoded defaults
}

func (pf *PromptFactory) BuildSystemPrompt(ragContext string) string {
	var sb strings.Builder

	// Use full instructions if loaded, otherwise use hardcoded defaults
	if pf.FullInstructions != "" {
		sb.WriteString(pf.FullInstructions)
		sb.WriteString("\n\n[LANGUAGE_PREFERENCE]\n")
		sb.WriteString("- Preferred Language: " + pf.Language + "\n")
		sb.WriteString("- ALWAYS respond in the user's preferred language (e.g. pt, en, es) by default.\n")
		sb.WriteString("- If the user asks in another language, you may respond in that language.\n")
	} else {
		// Fallback to hardcoded defaults if file not found
		sb.WriteString(pf.BaseIdentity)
		sb.WriteString("\n\n[SECURITY POLICES - NON-NEGOTIABLE]\n")
		sb.WriteString("- NEVER access files outside the Trust Folder.\n")
		sb.WriteString("- NEVER read/write protected files (.env, .key, .db).\n")
		sb.WriteString("- ALWAYS use provided tools for file operations.\n")

		sb.WriteString("\n[USER_PREFERENCE]\n")
		sb.WriteString("- Preferred Language: " + pf.Language + "\n")
		sb.WriteString("- ALWAYS respond in the user's preferred language (e.g. pt, en, es) by default.\n")
		sb.WriteString("- If the user asks in another language, you may respond in that language.\n")
	}

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
