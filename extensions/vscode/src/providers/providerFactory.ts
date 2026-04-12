/**
 * Provider Factory - Phase 9
 * Factory para criar e gerenciar instâncias de providers
 */

import { ClaudeProvider } from "./claudeProvider";
import { OpenAIProvider } from "./openaiProvider";
import { LlamaProvider } from "./llamaProvider";
import { GeminiProvider } from "./geminiProvider";
import {
  BaseProvider,
  ProviderType,
  ProviderConfig,
  ILLMProvider,
  ProviderCapabilities,
} from "./baseProvider";

// ============================================================================
// Provider Registry
// ============================================================================

interface ProviderRegistry {
  name: string;
  type: ProviderType;
  displayName: string;
  description: string;
  capabilities: ProviderCapabilities;
  create: (config?: ProviderConfig) => ILLMProvider;
}

// ============================================================================
// Provider Factory
// ============================================================================

export class ProviderFactory {
  private static instance: ProviderFactory;
  private providers: Map<ProviderType, ProviderRegistry> = new Map();
  private activeProviders: Map<ProviderType, ILLMProvider> = new Map();

  private constructor() {
    this.initializeProviders();
  }

  /**
   * Get factory singleton instance
   */
  static getInstance(): ProviderFactory {
    if (!ProviderFactory.instance) {
      ProviderFactory.instance = new ProviderFactory();
    }
    return ProviderFactory.instance;
  }

  /**
   * Initialize default providers
   */
  private initializeProviders(): void {
    this.registerProvider({
      name: "Claude",
      type: "claude",
      displayName: "Claude (Anthropic)",
      description:
        "Claude 3 family - Advanced reasoning and multimodal capabilities",
      capabilities: {
        streaming: true,
        vision: true,
        functionCalling: true,
        embedding: false,
        caching: true,
        maxTokens: 4096,
        contextWindow: 200000,
        costPer1kInputTokens: 0.003,
        costPer1kOutputTokens: 0.015,
      },
      create: (config) => new ClaudeProvider(config),
    });

    this.registerProvider({
      name: "OpenAI",
      type: "openai",
      displayName: "OpenAI (GPT)",
      description: "GPT-4 and GPT-3.5 - Industry-leading language models",
      capabilities: {
        streaming: true,
        vision: true,
        functionCalling: true,
        embedding: true,
        caching: false,
        maxTokens: 4096,
        contextWindow: 128000,
        costPer1kInputTokens: 0.03,
        costPer1kOutputTokens: 0.06,
      },
      create: (config) => new OpenAIProvider(config),
    });

    this.registerProvider({
      name: "Llama",
      type: "llama",
      displayName: "Llama (Meta)",
      description: "Open-source models - Local or cloud via Groq",
      capabilities: {
        streaming: true,
        vision: false,
        functionCalling: false,
        embedding: false,
        caching: false,
        maxTokens: 2048,
        contextWindow: 4096,
        costPer1kInputTokens: 0.0,
        costPer1kOutputTokens: 0.0,
      },
      create: (config) => new LlamaProvider(config),
    });

    this.registerProvider({
      name: "Gemini",
      type: "gemini",
      displayName: "Gemini (Google)",
      description: "Gemini 1.5 - Multimodal and long context",
      capabilities: {
        streaming: true,
        vision: true,
        functionCalling: false,
        embedding: true,
        caching: true,
        maxTokens: 8000,
        contextWindow: 1000000,
        costPer1kInputTokens: 0.0075,
        costPer1kOutputTokens: 0.03,
      },
      create: (config) => new GeminiProvider(config),
    });
  }

  /**
   * Register a new provider type
   */
  registerProvider(registry: ProviderRegistry): void {
    this.providers.set(registry.type, registry);
  }

  /**
   * Create a provider instance
   */
  createProvider(type: ProviderType, config?: ProviderConfig): ILLMProvider {
    const registry = this.providers.get(type);
    if (!registry) {
      throw new Error(`Provider type "${type}" is not registered`);
    }

    return registry.create(config);
  }

  /**
   * Create and initialize a provider
   */
  async createAndInitializeProvider(
    type: ProviderType,
    config: ProviderConfig
  ): Promise<ILLMProvider> {
    const provider = this.createProvider(type, config);
    await provider.initialize(config);
    return provider;
  }

  /**
   * Get or create active provider instance
   */
  getOrCreateProvider(type: ProviderType, config?: ProviderConfig): ILLMProvider {
    let provider = this.activeProviders.get(type);

    if (!provider) {
      provider = this.createProvider(type, config);
      this.activeProviders.set(type, provider);
    }

    return provider;
  }

  /**
   * Get active provider by type
   */
  getActiveProvider(type: ProviderType): ILLMProvider | null {
    return this.activeProviders.get(type) || null;
  }

  /**
   * Get all registered provider types
   */
  getRegisteredProviders(): Array<{
    type: ProviderType;
    name: string;
    displayName: string;
    description: string;
    capabilities: ProviderCapabilities;
  }> {
    const providers: Array<{
      type: ProviderType;
      name: string;
      displayName: string;
      description: string;
      capabilities: ProviderCapabilities;
    }> = [];

    this.providers.forEach((registry) => {
      providers.push({
        type: registry.type,
        name: registry.name,
        displayName: registry.displayName,
        description: registry.description,
        capabilities: registry.capabilities,
      });
    });

    return providers;
  }

  /**
   * Get provider registry by type
   */
  getProviderRegistry(type: ProviderType): ProviderRegistry | null {
    return this.providers.get(type) || null;
  }

  /**
   * Check if provider is supported
   */
  isProviderSupported(type: ProviderType): boolean {
    return this.providers.has(type);
  }

  /**
   * Remove active provider instance
   */
  removeActiveProvider(type: ProviderType): void {
    this.activeProviders.delete(type);
  }

  /**
   * Clear all active providers
   */
  clearActiveProviders(): void {
    this.activeProviders.clear();
  }

  /**
   * Get list of active provider types
   */
  getActiveProviderTypes(): ProviderType[] {
    return Array.from(this.activeProviders.keys());
  }
}

/**
 * Singleton factory instance export
 */
export const providerFactory = ProviderFactory.getInstance();
