import { describe, it, expect } from "vitest"

describe("embeddingModels", () => {
	describe("EMBEDDING_MODEL_PROFILES", () => {
		it("should have gemini provider with gemini-embedding-2-preview model", () => {
			const geminiProfiles = EMBEDDING_MODEL_PROFILES.gemini
			expect(geminiProfiles).toBeDefined()
			expect(geminiProfiles!["gemini-embedding-2-preview"]).toBeDefined()
			expect(geminiProfiles!["gemini-embedding-2-preview"].dimension).toBe(3072)
		})

		it("should have deprecated gemini-embedding-2-preview in gemini profiles for backward compatibility", () => {
			// This is critical for backward compatibility:
			// Users with gemini-embedding-2-preview configured need dimension lookup to work
			// even though the model is migrated to gemini-embedding-2-preview in GeminiEmbedder
			const geminiProfiles = EMBEDDING_MODEL_PROFILES.gemini
			expect(geminiProfiles).toBeDefined()
			expect(geminiProfiles!["gemini-embedding-2-preview"]).toBeDefined()
			expect(geminiProfiles!["gemini-embedding-2-preview"].dimension).toBe(3072)
		})
	})

	describe("getModelDimension", () => {
		it("should return dimension for gemini-embedding-2-preview", () => {
			const dimension = getModelDimension("gemini", "gemini-embedding-2-preview")
			expect(dimension).toBe(3072)
		})

		it("should return dimension for deprecated gemini-embedding-2-preview", () => {
			// This ensures createVectorStore() works for users with gemini-embedding-2-preview configured
			// The dimension should be 3072 (matching gemini-embedding-2-preview) because:
			// 1. GeminiEmbedder migrates gemini-embedding-2-preview to gemini-embedding-2-preview
			// 2. gemini-embedding-2-preview produces 3072-dimensional embeddings
			// 3. Vector store dimension must match the actual embedding dimension
			const dimension = getModelDimension("gemini", "gemini-embedding-2-preview")
			expect(dimension).toBe(3072)
		})

		it("should return undefined for unknown model", () => {
			const dimension = getModelDimension("gemini", "unknown-model")
			expect(dimension).toBeUndefined()
		})

		it("should return undefined for unknown provider", () => {
			const dimension = getModelDimension("unknown-provider" as any, "some-model")
			expect(dimension).toBeUndefined()
		})

		it("should return correct dimensions for openai models", () => {
			expect(getModelDimension("openai", "text-embedding-3-small")).toBe(1536)
			expect(getModelDimension("openai", "text-embedding-3-large")).toBe(3072)
			expect(getModelDimension("openai", "text-embedding-ada-002")).toBe(1536)
		})
	})

	describe("getModelScoreThreshold", () => {
		it("should return score threshold for gemini-embedding-2-preview", () => {
			const threshold = getModelScoreThreshold("gemini", "gemini-embedding-2-preview")
			expect(threshold).toBe(0.4)
		})

		it("should return score threshold for deprecated gemini-embedding-2-preview", () => {
			const threshold = getModelScoreThreshold("gemini", "gemini-embedding-2-preview")
			expect(threshold).toBe(0.4)
		})

		it("should return undefined for unknown model", () => {
			const threshold = getModelScoreThreshold("gemini", "unknown-model")
			expect(threshold).toBeUndefined()
		})
	})

	describe("getDefaultModelId", () => {
		it("should return gemini-embedding-2-preview for gemini provider", () => {
			const defaultModel = getDefaultModelId("gemini")
			expect(defaultModel).toBe("gemini-embedding-2-preview")
		})

		it("should return text-embedding-3-small for openai provider", () => {
			const defaultModel = getDefaultModelId("openai")
			expect(defaultModel).toBe("text-embedding-3-small")
		})

		it("should return codestral-embed-2505 for mistral provider", () => {
			const defaultModel = getDefaultModelId("mistral")
			expect(defaultModel).toBe("codestral-embed-2505")
		})
	})
})
