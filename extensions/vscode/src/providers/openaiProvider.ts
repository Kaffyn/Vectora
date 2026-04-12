/**
 * OpenAI Provider - Phase 9
 * Implementação do provider OpenAI/GPT
 * Suporta GPT-4, GPT-4 Turbo, GPT-3.5 com function calling e vision
 */

import {
  BaseProvider,
  ProviderCapabilities,
  ProviderConfig,
  PromptRequest,
  PromptResponse,
  ValidationResult,
  TokenCountResult,
  RateLimitInfo,
} from "./baseProvider";

// ============================================================================
// Types
// ============================================================================

interface OpenAIMessage {
  role: "user" | "assistant" | "system";
  content: string | Array<{ type: string; [key: string]: unknown }>;
}

interface OpenAIResponse {
  id: string;
  object: string;
  created: number;
  model: string;
  choices: Array<{
    index: number;
    message: {
      role: string;
      content: string;
    };
    tool_calls?: Array<{
      id: string;
      type: string;
      function: {
        name: string;
        arguments: string;
      };
    }>;
    finish_reason: string;
  }>;
  usage: {
    prompt_tokens: number;
    completion_tokens: number;
    total_tokens: number;
  };
}

// ============================================================================
// OpenAI Provider Implementation
// ============================================================================

export class OpenAIProvider extends BaseProvider {
  private apiKey: string = "";
  private baseUrl: string = "https://api.openai.com/v1";
  private modelVersion: string = "gpt-4o";
  private organization?: string;
  private rateLimits: RateLimitInfo | null = null;

  // Capabilities for GPT-4
  private static CAPABILITIES: ProviderCapabilities = {
    streaming: true,
    vision: true,
    functionCalling: true,
    embedding: true,
    caching: false,
    maxTokens: 4096,
    contextWindow: 128000,
    costPer1kInputTokens: 0.03,
    costPer1kOutputTokens: 0.06,
  };

  // Supported OpenAI models
  private static OPENAI_MODELS = [
    "gpt-4o",
    "gpt-4-turbo",
    "gpt-4",
    "gpt-3.5-turbo",
  ];

  constructor(config?: ProviderConfig) {
    super("OpenAI", "openai", OpenAIProvider.CAPABILITIES, config || {});

    if (config) {
      this.config = config;
      if (config.apiKey) {
        this.apiKey = config.apiKey;
      }
      if (config.baseUrl) {
        this.baseUrl = config.baseUrl;
      }
      if (config.customHeaders?.["organization"]) {
        this.organization = config.customHeaders["organization"];
      }
    }
  }

  /**
   * Initialize OpenAI provider
   */
  async initialize(config: ProviderConfig): Promise<void> {
    this.config = config;

    if (config.apiKey) {
      this.apiKey = config.apiKey;
    }

    if (config.baseUrl) {
      this.baseUrl = config.baseUrl;
    }

    if (config.customHeaders?.["organization"]) {
      this.organization = config.customHeaders["organization"];
    }

    const validation = this.validateConfig();
    if (!validation.valid) {
      throw new Error(
        `OpenAI provider initialization failed: ${validation.errors.join(", ")}`
      );
    }

    this.isInitialized = true;
  }

  /**
   * Validate OpenAI configuration
   */
  validateConfig(): ValidationResult {
    const errors: string[] = [];
    const warnings: string[] = [];

    if (!this.apiKey) {
      errors.push("API key is required for OpenAI provider");
    } else if (!this.apiKey.startsWith("sk-")) {
      warnings.push("API key does not follow expected OpenAI format");
    }

    if (!this.baseUrl) {
      errors.push("Base URL is required for OpenAI provider");
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings,
    };
  }

  /**
   * Send prompt to OpenAI
   */
  async sendPrompt(request: PromptRequest): Promise<PromptResponse> {
    const validation = this.validateConfig();
    if (!validation.valid) {
      throw new Error(`Provider not initialized: ${validation.errors.join(", ")}`);
    }

    const messages: OpenAIMessage[] = request.messages.map((m) => ({
      role: m.role,
      content: m.content,
    }));

    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      Authorization: `Bearer ${this.apiKey}`,
    };

    if (this.organization) {
      headers["OpenAI-Organization"] = this.organization;
    }

    try {
      const response = await fetch(`${this.baseUrl}/chat/completions`, {
        method: "POST",
        headers,
        body: JSON.stringify({
          model: this.modelVersion,
          messages,
          max_tokens: request.maxTokens || this.capabilities.maxTokens,
          temperature: request.temperature || 0.7,
          top_p: request.topP,
          stop: request.stopSequences,
        }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(
          `OpenAI API error: ${error.error?.message || response.statusText}`
        );
      }

      const data = (await response.json()) as OpenAIResponse;

      // Update rate limit info from headers
      this.updateRateLimitInfo(response.headers);

      const choice = data.choices[0];
      const toolCalls = choice.tool_calls?.map((tc) => ({
        id: tc.id,
        name: tc.function.name,
        input: JSON.parse(tc.function.arguments),
      }));

      return {
        content: choice.message.content,
        toolCalls,
        usage: {
          inputTokens: data.usage.prompt_tokens,
          outputTokens: data.usage.completion_tokens,
          totalTokens: data.usage.total_tokens,
        },
        stopReason: this.mapStopReason(choice.finish_reason),
      };
    } catch (error) {
      throw new Error(
        `OpenAI sendPrompt failed: ${error instanceof Error ? error.message : String(error)}`
      );
    }
  }

  /**
   * Stream prompt response
   */
  async *streamPrompt(request: PromptRequest): AsyncGenerator<string, void, unknown> {
    const validation = this.validateConfig();
    if (!validation.valid) {
      throw new Error(`Provider not initialized: ${validation.errors.join(", ")}`);
    }

    const messages: OpenAIMessage[] = request.messages.map((m) => ({
      role: m.role,
      content: m.content,
    }));

    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      Authorization: `Bearer ${this.apiKey}`,
    };

    if (this.organization) {
      headers["OpenAI-Organization"] = this.organization;
    }

    try {
      const response = await fetch(`${this.baseUrl}/chat/completions`, {
        method: "POST",
        headers,
        body: JSON.stringify({
          model: this.modelVersion,
          messages,
          max_tokens: request.maxTokens || this.capabilities.maxTokens,
          temperature: request.temperature || 0.7,
          top_p: request.topP,
          stop: request.stopSequences,
          stream: true,
        }),
      });

      if (!response.ok) {
        throw new Error(`OpenAI API error: ${response.statusText}`);
      }

      if (!response.body) {
        throw new Error("No response body from OpenAI API");
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");

        for (let i = 0; i < lines.length - 1; i++) {
          const line = lines[i].trim();
          if (line.startsWith("data: ") && line !== "data: [DONE]") {
            try {
              const event = JSON.parse(line.slice(6));
              if (
                event.choices &&
                event.choices[0]?.delta?.content
              ) {
                yield event.choices[0].delta.content;
              }
            } catch (e) {
              // Ignore parsing errors
            }
          }
        }

        buffer = lines[lines.length - 1];
      }
    } catch (error) {
      throw new Error(
        `OpenAI streamPrompt failed: ${error instanceof Error ? error.message : String(error)}`
      );
    }
  }

  /**
   * Get token count for text
   */
  getTokenCount(text: string): TokenCountResult {
    // OpenAI approximation: ~1 token per 4 characters
    // More accurate counting would require the js-tiktoken library
    const preparedText = this.prepareTextForTokenCounting(text);
    const tokens = Math.ceil(preparedText.length / 4);
    return {
      tokens,
      characters: preparedText.length,
    };
  }

  /**
   * Get rate limit information
   */
  getRateLimitInfo(): RateLimitInfo | null {
    return this.rateLimits;
  }

  /**
   * Get supported OpenAI models
   */
  getSupportedModels(): string[] {
    return OpenAIProvider.OPENAI_MODELS;
  }

  /**
   * Set OpenAI model version
   */
  setModelVersion(model: string): void {
    if (OpenAIProvider.OPENAI_MODELS.includes(model)) {
      this.modelVersion = model;
    } else {
      throw new Error(`Unsupported OpenAI model: ${model}`);
    }
  }

  /**
   * Get current model version
   */
  getModelVersion(): string {
    return this.modelVersion;
  }

  /**
   * Set organization ID for OpenAI
   */
  setOrganization(orgId: string): void {
    this.organization = orgId;
  }

  // ============================================================================
  // Private Methods
  // ============================================================================

  /**
   * Update rate limit info from response headers
   */
  private updateRateLimitInfo(headers: Headers): void {
    const requestsLimit = parseInt(
      headers.get("x-ratelimit-limit-requests") || "0"
    );
    const requestsRemaining = parseInt(
      headers.get("x-ratelimit-remaining-requests") || "0"
    );
    const tokensLimit = parseInt(
      headers.get("x-ratelimit-limit-tokens") || "0"
    );
    const tokensRemaining = parseInt(
      headers.get("x-ratelimit-remaining-tokens") || "0"
    );
    const resetTokens = parseInt(
      headers.get("x-ratelimit-reset-tokens") || "0"
    );

    if (requestsLimit > 0 || tokensLimit > 0) {
      const resetAt = new Date();
      resetAt.setSeconds(resetAt.getSeconds() + resetTokens);

      this.rateLimits = {
        requestsPerMinute: requestsLimit,
        tokensPerMinute: tokensLimit,
        requestsRemaining,
        tokensRemaining,
        resetAt,
      };
    }
  }

  /**
   * Map OpenAI stop reason to standard format
   */
  private mapStopReason(
    stopReason: string
  ): "end_turn" | "tool_use" | "max_tokens" | "stop_sequence" {
    switch (stopReason) {
      case "stop":
        return "stop_sequence";
      case "length":
        return "max_tokens";
      case "tool_calls":
        return "tool_use";
      case "content_filter":
      default:
        return "end_turn";
    }
  }
}
