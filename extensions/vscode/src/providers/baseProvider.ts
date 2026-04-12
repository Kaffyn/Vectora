/**
 * Base Provider Interface - Phase 9
 * Define a interface abstrata para todos os provedores LLM
 * Suporta Claude, OpenAI, Llama, Gemini e outros provedores
 */

// ============================================================================
// Types & Interfaces
// ============================================================================

export type ProviderType = "claude" | "openai" | "llama" | "gemini" | "custom";

export interface ProviderCapabilities {
  streaming: boolean;
  vision: boolean;
  functionCalling: boolean;
  embedding: boolean;
  caching: boolean;
  maxTokens: number;
  contextWindow: number;
  costPer1kInputTokens: number;
  costPer1kOutputTokens: number;
}

export interface ProviderConfig {
  name: string;
  apiKey?: string;
  baseUrl?: string;
  timeout?: number;
  retryAttempts?: number;
  retryDelay?: number;
  customHeaders?: Record<string, string>;
}

export interface PromptRequest {
  messages: Array<{
    role: "user" | "assistant" | "system";
    content: string;
  }>;
  temperature?: number;
  topP?: number;
  maxTokens?: number;
  stopSequences?: string[];
  systemPrompt?: string;
}

export interface PromptResponse {
  content: string;
  toolCalls?: Array<{
    id: string;
    name: string;
    input: Record<string, unknown>;
  }>;
  usage?: {
    inputTokens: number;
    outputTokens: number;
    totalTokens: number;
  };
  stopReason?: "end_turn" | "tool_use" | "max_tokens" | "stop_sequence";
}

export interface ValidationResult {
  valid: boolean;
  errors: string[];
  warnings: string[];
}

export interface TokenCountResult {
  tokens: number;
  characters: number;
}

export interface RateLimitInfo {
  requestsPerMinute: number;
  tokensPerMinute: number;
  requestsRemaining: number;
  tokensRemaining: number;
  resetAt: Date;
}

// ============================================================================
// Base Provider Abstract Class
// ============================================================================

export abstract class BaseProvider {
  protected name: string;
  protected type: ProviderType;
  protected capabilities: ProviderCapabilities;
  protected config: ProviderConfig;
  protected isInitialized: boolean = false;

  constructor(
    name: string,
    type: ProviderType,
    capabilities: ProviderCapabilities,
    config: ProviderConfig
  ) {
    this.name = name;
    this.type = type;
    this.capabilities = capabilities;
    this.config = config;
  }

  /**
   * Get provider name
   */
  getName(): string {
    return this.name;
  }

  /**
   * Get provider type
   */
  getType(): ProviderType {
    return this.type;
  }

  /**
   * Get provider capabilities
   */
  getCapabilities(): ProviderCapabilities {
    return this.capabilities;
  }

  /**
   * Check if provider is initialized
   */
  isReady(): boolean {
    return this.isInitialized;
  }

  /**
   * Initialize provider with given configuration
   */
  abstract initialize(config: ProviderConfig): Promise<void>;

  /**
   * Validate provider configuration
   */
  abstract validateConfig(): ValidationResult;

  /**
   * Send prompt and get response
   */
  abstract sendPrompt(request: PromptRequest): Promise<PromptResponse>;

  /**
   * Stream prompt response as async generator
   */
  abstract streamPrompt(
    request: PromptRequest
  ): AsyncGenerator<string, void, unknown>;

  /**
   * Count tokens in text
   */
  abstract getTokenCount(text: string): TokenCountResult;

  /**
   * Get current rate limit info
   */
  abstract getRateLimitInfo(): RateLimitInfo | null;

  /**
   * Estimate cost for given token counts
   */
  estimateCost(
    inputTokens: number,
    outputTokens: number
  ): {
    inputCost: number;
    outputCost: number;
    totalCost: number;
  } {
    const inputCost = (inputTokens / 1000) * this.capabilities.costPer1kInputTokens;
    const outputCost = (outputTokens / 1000) * this.capabilities.costPer1kOutputTokens;
    return {
      inputCost,
      outputCost,
      totalCost: inputCost + outputCost,
    };
  }

  /**
   * Check if provider has capability
   */
  hasCapability(capability: keyof ProviderCapabilities): boolean {
    return (this.capabilities[capability] as boolean | number) !== false &&
           (this.capabilities[capability] as boolean | number) !== 0;
  }

  /**
   * Get maximum context window size
   */
  getMaxContextWindow(): number {
    return this.capabilities.contextWindow;
  }

  /**
   * Get maximum tokens that can be requested
   */
  getMaxTokens(): number {
    return this.capabilities.maxTokens;
  }

  /**
   * Safely prepare text for token counting
   */
  protected prepareTextForTokenCounting(text: string): string {
    return text.trim();
  }

  /**
   * Validate API key format
   */
  protected isValidApiKey(): boolean {
    if (!this.config.apiKey) {
      return false;
    }
    return this.config.apiKey.length > 0;
  }

  /**
   * Get provider display name
   */
  getDisplayName(): string {
    return this.name.charAt(0).toUpperCase() + this.name.slice(1);
  }
}

// ============================================================================
// Provider Interface (for type-safe implementations)
// ============================================================================

export interface ILLMProvider {
  getName(): string;
  getType(): ProviderType;
  getCapabilities(): ProviderCapabilities;
  isReady(): boolean;
  initialize(config: ProviderConfig): Promise<void>;
  validateConfig(): ValidationResult;
  sendPrompt(request: PromptRequest): Promise<PromptResponse>;
  streamPrompt(
    request: PromptRequest
  ): AsyncGenerator<string, void, unknown>;
  getTokenCount(text: string): TokenCountResult;
  getRateLimitInfo(): RateLimitInfo | null;
  estimateCost(
    inputTokens: number,
    outputTokens: number
  ): {
    inputCost: number;
    outputCost: number;
    totalCost: number;
  };
  hasCapability(capability: keyof ProviderCapabilities): boolean;
  getMaxContextWindow(): number;
  getMaxTokens(): number;
  getDisplayName(): string;
}
