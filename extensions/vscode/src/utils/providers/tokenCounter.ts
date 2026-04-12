/**
 * Token Counter - Phase 9
 * Contagem de tokens por provider com aproximações e validações
 */

import { ProviderType } from "../../providers/baseProvider";

// ============================================================================
// Token Counting Utilities
// ============================================================================

/**
 * Base token counter with provider-specific algorithms
 */
export class TokenCounter {
  /**
   * Count tokens for Claude
   * Aproximação: ~1 token per 3-4 characters
   */
  static countClaudeTokens(text: string): number {
    const cleaned = text.trim();
    // Claude uses BPE tokenizer
    // Conservative estimate: 1 token per 3.5 chars
    return Math.ceil(cleaned.length / 3.5);
  }

  /**
   * Count tokens for OpenAI models
   * Aproximação: ~1 token per 4 characters
   */
  static countOpenAITokens(text: string): number {
    const cleaned = text.trim();
    // OpenAI uses BPE tokenizer similar to GPT
    // Conservative estimate: 1 token per 4 chars
    return Math.ceil(cleaned.length / 4);
  }

  /**
   * Count tokens for Llama models
   * Aproximação: ~1 token per 4 characters
   */
  static countLlamaTokens(text: string): number {
    const cleaned = text.trim();
    // Llama uses SentencePiece tokenizer
    // Similar to OpenAI: ~1 token per 4 chars
    return Math.ceil(cleaned.length / 4);
  }

  /**
   * Count tokens for Gemini
   * Aproximação: ~1 token per 3.5 characters
   */
  static countGeminiTokens(text: string): number {
    const cleaned = text.trim();
    // Gemini uses different tokenizer
    // Estimate: 1 token per 3.5 chars
    return Math.ceil(cleaned.length / 3.5);
  }

  /**
   * Count tokens for any provider
   */
  static countTokens(text: string, provider: ProviderType): number {
    switch (provider) {
      case "claude":
        return this.countClaudeTokens(text);
      case "openai":
        return this.countOpenAITokens(text);
      case "llama":
        return this.countLlamaTokens(text);
      case "gemini":
        return this.countGeminiTokens(text);
      default:
        // Default fallback
        return Math.ceil(text.trim().length / 4);
    }
  }

  /**
   * Count tokens in messages
   */
  static countMessagesTokens(
    messages: Array<{ role: string; content: string }>,
    provider: ProviderType
  ): number {
    let total = 0;

    // Add overhead for message format
    const messageOverhead = 4; // tokens for role/formatting

    for (const message of messages) {
      total += messageOverhead;
      total += this.countTokens(message.content, provider);
    }

    return total;
  }

  /**
   * Estimate cost for token usage
   */
  static estimateCost(
    inputTokens: number,
    outputTokens: number,
    provider: ProviderType,
    costPerInputK: number,
    costPerOutputK: number
  ): {
    inputCost: number;
    outputCost: number;
    totalCost: number;
  } {
    const inputCost = (inputTokens / 1000) * costPerInputK;
    const outputCost = (outputTokens / 1000) * costPerOutputK;

    return {
      inputCost: parseFloat(inputCost.toFixed(6)),
      outputCost: parseFloat(outputCost.toFixed(6)),
      totalCost: parseFloat((inputCost + outputCost).toFixed(6)),
    };
  }

  /**
   * Format tokens for display
   */
  static formatTokens(tokens: number): string {
    if (tokens < 1000) {
      return `${tokens}`;
    }
    return `${(tokens / 1000).toFixed(1)}K`;
  }

  /**
   * Format cost for display
   */
  static formatCost(cost: number): string {
    if (cost < 0.01) {
      return `$${cost.toFixed(4)}`;
    }
    return `$${cost.toFixed(2)}`;
  }

  /**
   * Calculate time estimate for processing
   */
  static estimateProcessingTime(
    tokens: number,
    provider: ProviderType
  ): {
    minMs: number;
    maxMs: number;
    avgMs: number;
  } {
    // Tokens per second by provider (conservative estimates)
    const tokensPerSecond: Record<ProviderType, { min: number; avg: number; max: number }> = {
      claude: { min: 50, avg: 80, max: 150 },
      openai: { min: 40, avg: 60, max: 120 },
      llama: { min: 30, avg: 50, max: 100 },
      gemini: { min: 60, avg: 100, max: 200 },
    };

    const rates = tokensPerSecond[provider] || tokensPerSecond.openai;

    return {
      minMs: (tokens / rates.max) * 1000,
      avgMs: (tokens / rates.avg) * 1000,
      maxMs: (tokens / rates.min) * 1000,
    };
  }

  /**
   * Format time estimate for display
   */
  static formatTimeEstimate(ms: number): string {
    if (ms < 1000) {
      return `${Math.round(ms)}ms`;
    }
    return `${(ms / 1000).toFixed(1)}s`;
  }

  /**
   * Check if text fits in context window
   */
  static fits(
    text: string,
    provider: ProviderType,
    contextWindow: number
  ): {
    fits: boolean;
    tokens: number;
    remaining: number;
    percentage: number;
  } {
    const tokens = this.countTokens(text, provider);
    const remaining = contextWindow - tokens;
    const percentage = (tokens / contextWindow) * 100;

    return {
      fits: tokens <= contextWindow,
      tokens,
      remaining: Math.max(0, remaining),
      percentage: parseFloat(percentage.toFixed(2)),
    };
  }
}

// ============================================================================
// Advanced Token Counting
// ============================================================================

export interface TokenCountAnalysis {
  totalTokens: number;
  tokensByMessage: number[];
  averagePerMessage: number;
  maxPerMessage: number;
  minPerMessage: number;
  estimatedCost: {
    inputCost: number;
    outputCost: number;
    totalCost: number;
  };
  estimatedProcessingTime: {
    minMs: number;
    maxMs: number;
    avgMs: number;
  };
}

export class AdvancedTokenCounter {
  /**
   * Analyze token usage in messages
   */
  static analyzeMessages(
    messages: Array<{ role: string; content: string }>,
    provider: ProviderType,
    costPerInputK: number,
    costPerOutputK: number
  ): TokenCountAnalysis {
    const tokensByMessage = messages.map((msg) =>
      TokenCounter.countTokens(msg.content, provider)
    );

    const totalTokens = tokensByMessage.reduce((a, b) => a + b, 0);
    const avgPerMessage = totalTokens / messages.length;
    const maxPerMessage = Math.max(...tokensByMessage);
    const minPerMessage = Math.min(...tokensByMessage);

    const estimatedCost = TokenCounter.estimateCost(
      totalTokens,
      0,
      provider,
      costPerInputK,
      costPerOutputK
    );

    const estimatedProcessingTime = TokenCounter.estimateProcessingTime(
      totalTokens,
      provider
    );

    return {
      totalTokens,
      tokensByMessage,
      averagePerMessage: parseFloat(avgPerMessage.toFixed(2)),
      maxPerMessage,
      minPerMessage,
      estimatedCost,
      estimatedProcessingTime,
    };
  }

  /**
   * Find optimization opportunities
   */
  static findOptimizations(
    messages: Array<{ role: string; content: string }>,
    provider: ProviderType,
    threshold: number = 0.8
  ): Array<{
    messageIndex: number;
    tokens: number;
    suggestion: string;
    potentialSavings: number;
  }> {
    const analysis = this.analyzeMessages(
      messages,
      provider,
      0,
      0
    );
    const suggestions = [];

    for (let i = 0; i < messages.length; i++) {
      const tokens = analysis.tokensByMessage[i];
      const usage = (tokens / analysis.totalTokens) * 100;

      if (usage > threshold * 100) {
        suggestions.push({
          messageIndex: i,
          tokens,
          suggestion: `Message ${i} uses ${usage.toFixed(1)}% of tokens. Consider summarizing or splitting.`,
          potentialSavings: Math.round(tokens * 0.3), // Assume 30% potential savings
        });
      }
    }

    return suggestions;
  }

  /**
   * Estimate tokens for generation
   */
  static estimateGenerationTokens(
    prompt: string,
    expectedLength: "short" | "medium" | "long",
    provider: ProviderType
  ): {
    inputTokens: number;
    estimatedOutputTokens: number;
    totalTokens: number;
  } {
    const inputTokens = TokenCounter.countTokens(prompt, provider);

    // Estimate output based on expected length
    const outputMultipliers: Record<string, number> = {
      short: 0.5,
      medium: 2,
      long: 5,
    };

    const estimatedOutputTokens = Math.round(
      inputTokens * outputMultipliers[expectedLength]
    );

    return {
      inputTokens,
      estimatedOutputTokens,
      totalTokens: inputTokens + estimatedOutputTokens,
    };
  }
}
