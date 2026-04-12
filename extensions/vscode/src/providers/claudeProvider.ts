/**
 * Claude Provider - Phase 9
 * Implementação do provider Claude (primário)
 * Suporta Claude 3 Opus, Sonnet, Haiku com streaming e vision
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

interface ClaudeMessage {
  role: "user" | "assistant";
  content: string | Array<{ type: string; [key: string]: unknown }>;
}

interface ClaudeResponse {
  id: string;
  type: string;
  role: string;
  content: Array<{
    type: string;
    text?: string;
    id?: string;
    name?: string;
    input?: Record<string, unknown>;
  }>;
  model: string;
  stop_reason: string;
  stop_sequence?: string;
  usage: {
    input_tokens: number;
    output_tokens: number;
    cache_creation_input_tokens?: number;
    cache_read_input_tokens?: number;
  };
}

// ============================================================================
// Claude Provider Implementation
// ============================================================================

export class ClaudeProvider extends BaseProvider {
  private apiKey: string = "";
  private baseUrl: string = "https://api.anthropic.com/v1";
  private modelVersion: string = "claude-3-5-sonnet-20241022";
  private rateLimits: RateLimitInfo | null = null;
  private requestCount: number = 0;
  private tokenCount: number = 0;
  private lastResetTime: Date = new Date();

  // Capabilities for Claude 3.5 Sonnet
  private static CAPABILITIES: ProviderCapabilities = {
    streaming: true,
    vision: true,
    functionCalling: true,
    embedding: false,
    caching: true,
    maxTokens: 4096,
    contextWindow: 200000,
    costPer1kInputTokens: 0.003,
    costPer1kOutputTokens: 0.015,
  };

  // Supported Claude models
  private static CLAUDE_MODELS = [
    "claude-3-5-sonnet-20241022",
    "claude-3-5-haiku-20241022",
    "claude-3-opus-20250219",
  ];

  constructor(config?: ProviderConfig) {
    super("Claude", "claude", ClaudeProvider.CAPABILITIES, config || {});

    if (config) {
      this.config = config;
      if (config.apiKey) {
        this.apiKey = config.apiKey;
      }
      if (config.baseUrl) {
        this.baseUrl = config.baseUrl;
      }
    }
  }

  /**
   * Initialize Claude provider
   */
  async initialize(config: ProviderConfig): Promise<void> {
    this.config = config;

    if (config.apiKey) {
      this.apiKey = config.apiKey;
    }

    if (config.baseUrl) {
      this.baseUrl = config.baseUrl;
    }

    const validation = this.validateConfig();
    if (!validation.valid) {
      throw new Error(`Claude provider initialization failed: ${validation.errors.join(", ")}`);
    }

    this.isInitialized = true;
  }

  /**
   * Validate Claude configuration
   */
  validateConfig(): ValidationResult {
    const errors: string[] = [];
    const warnings: string[] = [];

    if (!this.apiKey) {
      errors.push("API key is required for Claude provider");
    } else if (!this.apiKey.startsWith("sk-ant-")) {
      warnings.push("API key does not follow expected Claude format");
    }

    if (!this.baseUrl) {
      errors.push("Base URL is required for Claude provider");
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings,
    };
  }

  /**
   * Send prompt to Claude
   */
  async sendPrompt(request: PromptRequest): Promise<PromptResponse> {
    const validation = this.validateConfig();
    if (!validation.valid) {
      throw new Error(`Provider not initialized: ${validation.errors.join(", ")}`);
    }

    const messages = request.messages
      .filter((m) => m.role !== "system")
      .map((m) => ({
        role: m.role,
        content: m.content,
      })) as ClaudeMessage[];

    const systemPrompt = request.messages.find((m) => m.role === "system")?.content || request.systemPrompt;

    try {
      const response = await fetch(`${this.baseUrl}/messages`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "x-api-key": this.apiKey,
          "anthropic-version": "2023-06-01",
        },
        body: JSON.stringify({
          model: this.modelVersion,
          max_tokens: request.maxTokens || this.capabilities.maxTokens,
          temperature: request.temperature || 0.7,
          top_p: request.topP,
          stop_sequences: request.stopSequences,
          system: systemPrompt,
          messages,
        }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(`Claude API error: ${error.error?.message || response.statusText}`);
      }

      const data = (await response.json()) as ClaudeResponse;

      // Update rate limit info from headers
      this.updateRateLimitInfo(response.headers);

      // Extract tool calls and content
      const toolCalls = data.content
        .filter((c) => c.type === "tool_use")
        .map((c) => ({
          id: c.id || "",
          name: c.name || "",
          input: c.input || {},
        }));

      const content = data.content
        .filter((c) => c.type === "text")
        .map((c) => c.text || "")
        .join("\n");

      return {
        content,
        toolCalls: toolCalls.length > 0 ? toolCalls : undefined,
        usage: {
          inputTokens: data.usage.input_tokens,
          outputTokens: data.usage.output_tokens,
          totalTokens: data.usage.input_tokens + data.usage.output_tokens,
        },
        stopReason: this.mapStopReason(data.stop_reason),
      };
    } catch (error) {
      throw new Error(`Claude sendPrompt failed: ${error instanceof Error ? error.message : String(error)}`);
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

    const messages = request.messages
      .filter((m) => m.role !== "system")
      .map((m) => ({
        role: m.role,
        content: m.content,
      })) as ClaudeMessage[];

    const systemPrompt = request.messages.find((m) => m.role === "system")?.content || request.systemPrompt;

    try {
      const response = await fetch(`${this.baseUrl}/messages`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "x-api-key": this.apiKey,
          "anthropic-version": "2023-06-01",
        },
        body: JSON.stringify({
          model: this.modelVersion,
          max_tokens: request.maxTokens || this.capabilities.maxTokens,
          temperature: request.temperature || 0.7,
          top_p: request.topP,
          stop_sequences: request.stopSequences,
          system: systemPrompt,
          messages,
          stream: true,
        }),
      });

      if (!response.ok) {
        throw new Error(`Claude API error: ${response.statusText}`);
      }

      if (!response.body) {
        throw new Error("No response body from Claude API");
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
          if (line.startsWith("data: ")) {
            try {
              const event = JSON.parse(line.slice(6));
              if (event.type === "content_block_delta" && event.delta?.type === "text_delta") {
                yield event.delta.text;
              }
            } catch (e) {
              // Ignore parsing errors
            }
          }
        }

        buffer = lines[lines.length - 1];
      }
    } catch (error) {
      throw new Error(`Claude streamPrompt failed: ${error instanceof Error ? error.message : String(error)}`);
    }
  }

  /**
   * Get token count for text
   */
  getTokenCount(text: string): TokenCountResult {
    // Claude uses a simple approximation: ~1 token per 3-4 characters
    // More accurate counting would require the tokenizer library
    const preparedText = this.prepareTextForTokenCounting(text);
    const tokens = Math.ceil(preparedText.length / 3.5);
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
   * Get supported Claude models
   */
  getSupportedModels(): string[] {
    return ClaudeProvider.CLAUDE_MODELS;
  }

  /**
   * Set Claude model version
   */
  setModelVersion(model: string): void {
    if (ClaudeProvider.CLAUDE_MODELS.includes(model)) {
      this.modelVersion = model;
    } else {
      throw new Error(`Unsupported Claude model: ${model}`);
    }
  }

  /**
   * Get current model version
   */
  getModelVersion(): string {
    return this.modelVersion;
  }

  // ============================================================================
  // Private Methods
  // ============================================================================

  /**
   * Update rate limit info from response headers
   */
  private updateRateLimitInfo(headers: Headers): void {
    const requestsLimit = parseInt(headers.get("anthropic-ratelimit-requests-limit") || "0");
    const requestsRemaining = parseInt(headers.get("anthropic-ratelimit-requests-remaining") || "0");
    const tokensLimit = parseInt(headers.get("anthropic-ratelimit-tokens-limit") || "0");
    const tokensRemaining = parseInt(headers.get("anthropic-ratelimit-tokens-remaining") || "0");
    const resetDate = headers.get("anthropic-ratelimit-reset-tokens");

    if (requestsLimit > 0 || tokensLimit > 0) {
      this.rateLimits = {
        requestsPerMinute: requestsLimit,
        tokensPerMinute: tokensLimit,
        requestsRemaining,
        tokensRemaining,
        resetAt: resetDate ? new Date(resetDate) : new Date(),
      };
    }
  }

  /**
   * Map Claude stop reason to standard format
   */
  private mapStopReason(
    stopReason: string
  ): "end_turn" | "tool_use" | "max_tokens" | "stop_sequence" {
    switch (stopReason) {
      case "end_turn":
        return "end_turn";
      case "tool_use":
        return "tool_use";
      case "max_tokens":
        return "max_tokens";
      case "stop_sequence":
        return "stop_sequence";
      default:
        return "end_turn";
    }
  }
}
