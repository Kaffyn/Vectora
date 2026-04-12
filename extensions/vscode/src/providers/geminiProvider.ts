/**
 * Gemini Provider - Phase 9
 * Implementação do provider Google Gemini
 * Suporta Gemini 1.5, Gemini Pro com vision e embeddings
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

interface GeminiMessage {
  role: "user" | "model";
  parts: Array<{
    text?: string;
    inline_data?: {
      mime_type: string;
      data: string;
    };
  }>;
}

interface GeminiResponse {
  candidates: Array<{
    content: {
      parts: Array<{
        text: string;
      }>;
    };
    finish_reason: string;
    index: number;
  }>;
  usage_metadata: {
    prompt_token_count: number;
    candidates_token_count: number;
    total_token_count: number;
  };
}

interface GeminiStreamResponse {
  candidates: Array<{
    content: {
      parts: Array<{
        text: string;
      }>;
    };
    finish_reason?: string;
    index: number;
  }>;
  usage_metadata?: {
    prompt_token_count: number;
    candidates_token_count: number;
    total_token_count: number;
  };
}

// ============================================================================
// Gemini Provider Implementation
// ============================================================================

export class GeminiProvider extends BaseProvider {
  private apiKey: string = "";
  private baseUrl: string = "https://generativelanguage.googleapis.com/v1beta";
  private modelVersion: string = "gemini-1.5-pro-latest";
  private rateLimits: RateLimitInfo | null = null;

  // Capabilities for Gemini 1.5 Pro
  private static CAPABILITIES: ProviderCapabilities = {
    streaming: true,
    vision: true,
    functionCalling: false,
    embedding: true,
    caching: true,
    maxTokens: 8000,
    contextWindow: 1000000,
    costPer1kInputTokens: 0.0075,
    costPer1kOutputTokens: 0.03,
  };

  // Supported Gemini models
  private static GEMINI_MODELS = [
    "gemini-1.5-pro-latest",
    "gemini-1.5-pro",
    "gemini-1.5-flash",
    "gemini-1.5-flash-latest",
    "gemini-pro",
    "gemini-pro-vision",
  ];

  constructor(config?: ProviderConfig) {
    super("Gemini", "gemini", GeminiProvider.CAPABILITIES, config || {});

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
   * Initialize Gemini provider
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
      throw new Error(
        `Gemini provider initialization failed: ${validation.errors.join(", ")}`
      );
    }

    this.isInitialized = true;
  }

  /**
   * Validate Gemini configuration
   */
  validateConfig(): ValidationResult {
    const errors: string[] = [];
    const warnings: string[] = [];

    if (!this.apiKey) {
      errors.push("API key is required for Gemini provider");
    }

    if (!this.baseUrl) {
      errors.push("Base URL is required for Gemini provider");
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings,
    };
  }

  /**
   * Send prompt to Gemini
   */
  async sendPrompt(request: PromptRequest): Promise<PromptResponse> {
    const validation = this.validateConfig();
    if (!validation.valid) {
      throw new Error(`Provider not initialized: ${validation.errors.join(", ")}`);
    }

    const messages: GeminiMessage[] = request.messages
      .filter((m) => m.role !== "system")
      .map((m) => ({
        role: m.role === "user" ? "user" : "model",
        parts: [
          {
            text: m.content,
          },
        ],
      }));

    const systemPrompt = request.messages.find((m) => m.role === "system")?.content;

    try {
      const body: Record<string, unknown> = {
        contents: messages,
        generation_config: {
          temperature: request.temperature || 0.7,
          top_p: request.topP,
          max_output_tokens: request.maxTokens || this.capabilities.maxTokens,
          stop_sequences: request.stopSequences,
        },
      };

      if (systemPrompt) {
        (body as Record<string, unknown>).system_instruction = {
          parts: [{ text: systemPrompt }],
        };
      }

      const response = await fetch(
        `${this.baseUrl}/models/${this.modelVersion}:generateContent?key=${this.apiKey}`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(body),
        }
      );

      if (!response.ok) {
        const error = await response.json();
        throw new Error(
          `Gemini API error: ${error.error?.message || response.statusText}`
        );
      }

      const data = (await response.json()) as GeminiResponse;

      // Update rate limit info from headers
      this.updateRateLimitInfo(response.headers);

      const candidate = data.candidates[0];
      const content = candidate.content.parts.map((p) => p.text).join("\n");

      return {
        content,
        usage: {
          inputTokens: data.usage_metadata.prompt_token_count,
          outputTokens: data.usage_metadata.candidates_token_count,
          totalTokens: data.usage_metadata.total_token_count,
        },
        stopReason: this.mapStopReason(candidate.finish_reason),
      };
    } catch (error) {
      throw new Error(
        `Gemini sendPrompt failed: ${error instanceof Error ? error.message : String(error)}`
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

    const messages: GeminiMessage[] = request.messages
      .filter((m) => m.role !== "system")
      .map((m) => ({
        role: m.role === "user" ? "user" : "model",
        parts: [
          {
            text: m.content,
          },
        ],
      }));

    const systemPrompt = request.messages.find((m) => m.role === "system")?.content;

    try {
      const body: Record<string, unknown> = {
        contents: messages,
        generation_config: {
          temperature: request.temperature || 0.7,
          top_p: request.topP,
          max_output_tokens: request.maxTokens || this.capabilities.maxTokens,
          stop_sequences: request.stopSequences,
        },
      };

      if (systemPrompt) {
        (body as Record<string, unknown>).system_instruction = {
          parts: [{ text: systemPrompt }],
        };
      }

      const response = await fetch(
        `${this.baseUrl}/models/${this.modelVersion}:streamGenerateContent?key=${this.apiKey}`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(body),
        }
      );

      if (!response.ok) {
        throw new Error(`Gemini API error: ${response.statusText}`);
      }

      if (!response.body) {
        throw new Error("No response body from Gemini API");
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
          if (line) {
            try {
              const event = JSON.parse(line) as GeminiStreamResponse;
              if (event.candidates && event.candidates[0]?.content?.parts) {
                for (const part of event.candidates[0].content.parts) {
                  if (part.text) {
                    yield part.text;
                  }
                }
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
        `Gemini streamPrompt failed: ${error instanceof Error ? error.message : String(error)}`
      );
    }
  }

  /**
   * Get token count for text
   */
  getTokenCount(text: string): TokenCountResult {
    // Gemini approximation: ~1 token per 3.5 characters
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
   * Get supported Gemini models
   */
  getSupportedModels(): string[] {
    return GeminiProvider.GEMINI_MODELS;
  }

  /**
   * Set Gemini model version
   */
  setModelVersion(model: string): void {
    if (GeminiProvider.GEMINI_MODELS.includes(model)) {
      this.modelVersion = model;
    } else {
      throw new Error(`Unsupported Gemini model: ${model}`);
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
    // Gemini uses different rate limit headers
    const rateLimitExceeded = headers.get("x-goog-request-params");

    if (rateLimitExceeded) {
      this.rateLimits = {
        requestsPerMinute: 60,
        tokensPerMinute: 1000000,
        requestsRemaining: 0,
        tokensRemaining: 0,
        resetAt: new Date(),
      };
    }
  }

  /**
   * Map Gemini stop reason to standard format
   */
  private mapStopReason(
    stopReason: string
  ): "end_turn" | "tool_use" | "max_tokens" | "stop_sequence" {
    switch (stopReason) {
      case "STOP":
        return "stop_sequence";
      case "MAX_TOKENS":
        return "max_tokens";
      case "SAFETY":
      case "RECITATION":
      case "OTHER":
      default:
        return "end_turn";
    }
  }
}
