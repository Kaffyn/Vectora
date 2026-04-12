package llm

// geminiAliases maps shorthand model names to canonical Gemini API model IDs.
// All Gemini chat models require the "-preview" suffix for API access.
var geminiAliases = map[string]string{
	// Gemini 3.1 series (latest)
	"gemini-3.1-pro":       "gemini-3.1-pro-preview",
	"gemini-3.1-flash":     "gemini-3.1-flash-preview",
	"gemini-3.1-ultra":     "gemini-3.1-ultra-preview",

	// Gemini 3.0 series
	"gemini-3-pro":         "gemini-3-pro-preview",
	"gemini-3-flash":       "gemini-3-flash-preview",

	// Also support exact preview names (passthrough)
	"gemini-3.1-pro-preview":   "gemini-3.1-pro-preview",
	"gemini-3.1-flash-preview": "gemini-3.1-flash-preview",
	"gemini-3.1-ultra-preview": "gemini-3.1-ultra-preview",
	"gemini-3-pro-preview":     "gemini-3-pro-preview",
	"gemini-3-flash-preview":   "gemini-3-flash-preview",

	// Shorthands
	"gemini":        "gemini-3.1-pro-preview", // Default to latest pro
	"flash":         "gemini-3.1-flash-preview",
	"pro":           "gemini-3.1-pro-preview",
}

// ResolveGeminiModel resolves a model alias to the canonical Gemini API model ID.
// If the model is not found in aliases, returns it as-is (assuming it's already correct).
func ResolveGeminiModel(model string) string {
	if resolved, ok := geminiAliases[model]; ok {
		return resolved
	}
	return model
}
