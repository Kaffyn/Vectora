package llm

import "strings"

// GatewayModelCatalog dynamically lists all known models available via gateways.
// This is now derived from the central ModelRegistry in models.go.
var GatewayModelCatalog map[string][]string

func init() {
	GatewayModelCatalog = make(map[string][]string)
	for _, info := range ModelRegistry {
		GatewayModelCatalog[info.Family] = append(GatewayModelCatalog[info.Family], info.GatewayID)
	}
}


// AllGatewayModels returns a flat slice of every known gateway model.
// Suitable for use as a static fallback in ListModels().
func AllGatewayModels() []string {
	var all []string
	for _, models := range GatewayModelCatalog {
		all = append(all, models...)
	}
	return all
}

// GatewayModelsForProvider returns models whose prefix matches the given provider name.
// For example, GatewayModelsForProvider("openrouter") returns all models.
// GatewayModelsForProvider("anthropic") returns only Anthropic models.
func GatewayModelsForProvider(providerName string) []string {
	lower := strings.ToLower(providerName)

	// OpenRouter and Anannas expose ALL providers.
	if lower == "openrouter" || lower == "anannas" || lower == "openai" {
		return AllGatewayModels()
	}

	// Otherwise match catalog key prefix.
	if models, ok := GatewayModelCatalog[lower]; ok {
		return models
	}

	return AllGatewayModels()
}

// EmbeddingModelForGateway determines the embedding model to use based on family.
// Returns the native embedding ID if known, or empty string to trigger
// the global router fallback to Voyage AI (voyage-3-large).
func EmbeddingModelForGateway(model string) string {
	if model == "" {
		return FamilyEmbeddingRegistry["openai"] // text-embedding-3-large
	}

	family := FamilyFromModel(model)
	if emb, ok := FamilyEmbeddingRegistry[family]; ok {
		return emb
	}

	// No native embedding known for this family (e.g. Anthropic, Meta).
	// Return empty to allow higher-level fallback to Voyage AI in router.go.
	return ""
}
