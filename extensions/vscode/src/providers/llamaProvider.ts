/**
 * Llama Provider - Phase 9
 * Implementação do provider Llama (via Ollama ou Groq)
 * Suporta Llama 2, Llama 3 com suporte a local e cloud
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

interface LlamaMessage {
  role: "user" | "assistant" | "system";
  content: string;
}

interface LlamaResponse {
  model: string;
  created_at: string;
  message: {
    role: string;
    content: string;
  };
  done: boolean;
  total_duration?: number;
  load_duration?: number;
  prompt_eval_count?: number;
  prompt_eval_duration?: number;
  eval_count?: number;
  eval_duration?: number;
}

interface LlamaStreamResponse {
  model: string;
  created_at: string;
  message?: {
    role: string;
    content: string;
  };
  response?: string;
  done: boolean;
}

// ============================================================================
// Llama Provider Implementation
// ============================================================================

export class LlamaProvider extends BaseProvider {
  private baseUrl: string = "http://localhost:11434";
  private modelVersion: string = "llama2";
  private mode: "local" | "groq" = "local";
  private apiKey: string = "";
  private rateLimits: RateLimitInfo | null = null;

  // Capabilities for Llama (conservative estimates)
  private static CAPABILITIES: ProviderCapabilities = {
    streaming: true,
    vision: false,
    functionCalling: false,
    embedding: false,
    caching: false,
    maxTokens: 2048,
    contextWindow: 4096,
    costPer1kInputTokens: 0.0,
    costPer1kOutputTokens: 0.0,
  };

  // Supported Llama models
  private static LLAMA_MODELS = [
    "llama2",
    "llama2-uncensored",
    "neural-chat",
    "mistral",
    "dolphin-mixtral",
  ];

  constructor(config?: ProviderConfig) {
    super("Llama", "llama", LlamaProvider.CAPABILITIES, config || {});

    if (config) {
      this.config = config;
      if (config.baseUrl) {
        this.baseUrl = config.baseUrl;
      }
      if (config.apiKey) {
        this.apiKey = config.apiKey;
        this.mode = "groq";
      }
    }
  }

  /**
   * Initialize Llama provider
   */
  async initialize(config: ProviderConfig): Promise<void> {
    this.config = config;

    if (config.baseUrl) {
      this.baseUrl = config.baseUrl;
    }

    if (config.apiKey) {
      this.apiKey = config.apiKey;
      this.mode = "groq";
    }

    // For local mode, try to connect to verify Ollama is running
    if (this.mode === "local") {
      await this.checkOllamaConnection();
    }

    this.isInitialized = true;
  }

  /**
   * Validate Llama configuration
   */
  validateConfig(): ValidationResult {
    const errors: string[] = [];
    const warnings: string[] = [];

    if (this.mode === "groq" && !this.apiKey) {
      errors.push("API key is required for Groq Llama provider");
    }

    if (this.mode === "local" && !this.baseUrl) {
      errors.push("Base URL is required for local Llama provider");
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings,
    };
  }

  /**
   * Send prompt to Llama
   */
  async sendPrompt(request: PromptRequest): Promise<PromptResponse> {
    const validation = this.validateConfig();
    if (!validation.valid) {
      throw new Error(`Provider not initialized: ${validation.errors.join(", ")}`);
    }

    const messages: LlamaMessage[] = request.messages.map((m) => ({
      role: m.role,
      content: m.content,
    }));

    // Build conversation history
    const conversationText = messages.map((m) => `${m.role}: ${m.content}`).join("\n");

    try {
      const headers: Record<string, string> = {
        "Content-Type": "application/json",
      };

      if (this.mode === "groq" && this.apiKey) {
        headers["Authorization"] = `Bearer ${this.apiKey}`;
      }

      const endpoint =
        this.mode === "groq"
          ? "https://api.groq.com/openai/v1/chat/completions"
          : `${this.baseUrl}/api/chat`;

      const body =
        this.mode === "groq"
          ? JSON.stringify({
              model: this.modelVersion,
              messages,
              temperature: request.temperature || 0.7,
              max_tokens: request.maxTokens || this.capabilities.maxTokens,
              top_p: request.topP,
              stream: false,
            })
          : JSON.stringify({
              model: this.modelVersion,
              messages,
              stream: false,
              temperature: request.temperature || 0.7,
            });

      const response = await fetch(endpoint, {
        method: "POST",
        headers,
        body,
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(
          `Llama API error: ${error.error?.message || response.statusText}`
        );
      }

      const data = (await response.json()) as LlamaResponse | { choices: Array<{ message: { content: string } }> };

      let content = "";
      let promptTokens = 0;
      let completionTokens = 0;

      if ("message" in data) {
        // Ollama format
        content = (data as LlamaResponse).message.content;
        promptTokens = (data as LlamaResponse).prompt_eval_count || 0;
        completionTokens = (data as LlamaResponse).eval_count || 0;
      } else if ("choices" in data) {
        // Groq format
        content = (data as { choices: Array<{ message: { content: string } }> }).choices[0].message.content;
      }

      return {
        content,
        usage: {
          inputTokens: promptTokens,
          outputTokens: completionTokens,
          totalTokens: promptTokens + completionTokens,
        },
        stopReason: "end_turn",
      };
    } catch (error) {
      throw new Error(
        `Llama sendPrompt failed: ${error instanceof Error ? error.message : String(error)}`
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

    const messages: LlamaMessage[] = request.messages.map((m) => ({
      role: m.role,
      content: m.content,
    }));

    try {
      const headers: Record<string, string> = {
        "Content-Type": "application/json",
      };

      if (this.mode === "groq" && this.apiKey) {
        headers["Authorization"] = `Bearer ${this.apiKey}`;
      }

      const endpoint =
        this.mode === "groq"
          ? "https://api.groq.com/openai/v1/chat/completions"
          : `${this.baseUrl}/api/chat`;

      const body =
        this.mode === "groq"
          ? JSON.stringify({
              model: this.modelVersion,
              messages,
              temperature: request.temperature || 0.7,
              max_tokens: request.maxTokens || this.capabilities.maxTokens,
              top_p: request.topP,
              stream: true,
            })
          : JSON.stringify({
              model: this.modelVersion,
              messages,
              stream: true,
              temperature: request.temperature || 0.7,
            });

      const response = await fetch(endpoint, {
        method: "POST",
        headers,
        body,
      });

      if (!response.ok) {
        throw new Error(`Llama API error: ${response.statusText}`);
      }

      if (!response.body) {
        throw new Error("No response body from Llama API");
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

          if (this.mode === "groq") {
            // Groq uses OpenAI format
            if (line.startsWith("data: ") && line !== "data: [DONE]") {
              try {
                const event = JSON.parse(line.slice(6));
                if (event.choices?.[0]?.delta?.content) {
                  yield event.choices[0].delta.content;
                }
              } catch (e) {
                // Ignore parsing errors
              }
            }
          } else {
            // Ollama format
            try {
              const event = JSON.parse(line) as LlamaStreamResponse;
              if (event.response) {
                yield event.response;
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
        `Llama streamPrompt failed: ${error instanceof Error ? error.message : String(error)}`
      );
    }
  }

  /**
   * Get token count for text
   */
  getTokenCount(text: string): TokenCountResult {
    // Llama approximation: ~1 token per 4 characters
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
   * Get supported Llama models
   */
  getSupportedModels(): string[] {
    return LlamaProvider.LLAMA_MODELS;
  }

  /**
   * Set Llama model version
   */
  setModelVersion(model: string): void {
    this.modelVersion = model;
  }

  /**
   * Get current model version
   */
  getModelVersion(): string {
    return this.modelVersion;
  }

  /**
   * Set provider mode (local or groq)
   */
  setMode(mode: "local" | "groq"): void {
    this.mode = mode;
  }

  /**
   * Get current mode
   */
  getMode(): "local" | "groq" {
    return this.mode;
  }

  // ============================================================================
  // Private Methods
  // ============================================================================

  /**
   * Check if Ollama is running locally
   */
  private async checkOllamaConnection(): Promise<void> {
    try {
      const response = await fetch(`${this.baseUrl}/api/tags`, {
        method: "GET",
        timeout: 5000,
      });

      if (!response.ok) {
        throw new Error("Ollama is not responding");
      }
    } catch (error) {
      throw new Error(
        `Cannot connect to Ollama at ${this.baseUrl}. Is it running? ${
          error instanceof Error ? error.message : String(error)
        }`
      );
    }
  }
}
