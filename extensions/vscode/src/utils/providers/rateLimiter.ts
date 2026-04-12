/**
 * Rate Limiter - Phase 9
 * Rate limiting e token management por provider
 */

import { ProviderType } from "../../providers/baseProvider";

// ============================================================================
// Rate Limiter Types
// ============================================================================

export interface RateLimitBucket {
  requestsLimit: number;
  requestsUsed: number;
  tokensLimit: number;
  tokensUsed: number;
  windowStart: Date;
  windowDuration: number; // milliseconds
}

export interface RateLimitStatus {
  requestsRemaining: number;
  tokensRemaining: number;
  requestsResetIn: number; // milliseconds
  tokensResetIn: number; // milliseconds
  canMakeRequest: boolean;
  canUseTokens: (tokens: number) => boolean;
}

export interface AdaptiveBackoffConfig {
  initialDelayMs: number;
  maxDelayMs: number;
  multiplier: number;
  jitterFactor: number;
}

// ============================================================================
// Rate Limiter Implementation
// ============================================================================

export class RateLimiter {
  private buckets: Map<ProviderType, RateLimitBucket> = new Map();
  private requestQueues: Map<ProviderType, Array<() => Promise<unknown>>> = new Map();
  private backoffStates: Map<ProviderType, { attempts: number; lastBackoffMs: number }> = new Map();

  private defaultLimits: Record<ProviderType, { requests: number; tokens: number }> = {
    claude: { requests: 60, tokens: 1000000 },
    openai: { requests: 200, tokens: 90000 },
    llama: { requests: 30, tokens: 300000 },
    gemini: { requests: 60, tokens: 1000000 },
  };

  private defaultBackoffConfig: AdaptiveBackoffConfig = {
    initialDelayMs: 100,
    maxDelayMs: 60000,
    multiplier: 2,
    jitterFactor: 0.1,
  };

  /**
   * Initialize rate limiter for provider
   */
  initializeProvider(
    provider: ProviderType,
    requestsPerMinute: number = this.defaultLimits[provider].requests,
    tokensPerMinute: number = this.defaultLimits[provider].tokens
  ): void {
    this.buckets.set(provider, {
      requestsLimit: requestsPerMinute,
      requestsUsed: 0,
      tokensLimit: tokensPerMinute,
      tokensUsed: 0,
      windowStart: new Date(),
      windowDuration: 60000, // 1 minute
    });

    this.requestQueues.set(provider, []);
    this.backoffStates.set(provider, { attempts: 0, lastBackoffMs: 0 });
  }

  /**
   * Check if request can be made
   */
  canMakeRequest(provider: ProviderType): boolean {
    const bucket = this.getBucket(provider);
    if (!bucket) return true;

    if (this.isWindowExpired(bucket)) {
      this.resetBucket(bucket);
    }

    return bucket.requestsUsed < bucket.requestsLimit;
  }

  /**
   * Check if tokens can be used
   */
  canUseTokens(provider: ProviderType, tokens: number): boolean {
    const bucket = this.getBucket(provider);
    if (!bucket) return true;

    if (this.isWindowExpired(bucket)) {
      this.resetBucket(bucket);
    }

    return bucket.tokensUsed + tokens <= bucket.tokensLimit;
  }

  /**
   * Record request usage
   */
  recordRequest(provider: ProviderType, tokens: number = 0): void {
    const bucket = this.getBucket(provider);
    if (!bucket) return;

    if (this.isWindowExpired(bucket)) {
      this.resetBucket(bucket);
    }

    bucket.requestsUsed += 1;
    bucket.tokensUsed += tokens;
  }

  /**
   * Get rate limit status
   */
  getStatus(provider: ProviderType): RateLimitStatus {
    const bucket = this.getBucket(provider);

    if (!bucket) {
      return {
        requestsRemaining: 999999,
        tokensRemaining: 999999,
        requestsResetIn: 0,
        tokensResetIn: 0,
        canMakeRequest: true,
        canUseTokens: () => true,
      };
    }

    if (this.isWindowExpired(bucket)) {
      this.resetBucket(bucket);
    }

    const timeUntilReset = this.getTimeUntilReset(bucket);

    return {
      requestsRemaining: Math.max(0, bucket.requestsLimit - bucket.requestsUsed),
      tokensRemaining: Math.max(0, bucket.tokensLimit - bucket.tokensUsed),
      requestsResetIn: timeUntilReset,
      tokensResetIn: timeUntilReset,
      canMakeRequest: this.canMakeRequest(provider),
      canUseTokens: (tokens: number) => this.canUseTokens(provider, tokens),
    };
  }

  /**
   * Queue request with rate limiting
   */
  async queueRequest<T>(
    provider: ProviderType,
    callback: () => Promise<T>,
    tokens: number = 0
  ): Promise<T> {
    while (!this.canMakeRequest(provider) || !this.canUseTokens(provider, tokens)) {
      const status = this.getStatus(provider);
      const waitTime = Math.max(status.requestsResetIn, status.tokensResetIn);
      await this.sleep(Math.min(waitTime, 5000)); // Max wait 5 seconds between checks
    }

    this.recordRequest(provider, tokens);
    return callback();
  }

  /**
   * Apply adaptive backoff
   */
  async applyAdaptiveBackoff(
    provider: ProviderType,
    config: Partial<AdaptiveBackoffConfig> = {}
  ): Promise<number> {
    const backoffConfig = { ...this.defaultBackoffConfig, ...config };
    const state = this.backoffStates.get(provider);

    if (!state) {
      this.backoffStates.set(provider, { attempts: 1, lastBackoffMs: 0 });
      return this.calculateBackoffDelay(backoffConfig, 1);
    }

    state.attempts += 1;
    const delayMs = this.calculateBackoffDelay(backoffConfig, state.attempts);
    state.lastBackoffMs = delayMs;

    await this.sleep(delayMs);
    return delayMs;
  }

  /**
   * Reset backoff state
   */
  resetBackoff(provider: ProviderType): void {
    const state = this.backoffStates.get(provider);
    if (state) {
      state.attempts = 0;
      state.lastBackoffMs = 0;
    }
  }

  /**
   * Get backoff state
   */
  getBackoffState(provider: ProviderType): { attempts: number; lastBackoffMs: number } | null {
    return this.backoffStates.get(provider) || null;
  }

  /**
   * Update rate limits from response headers
   */
  updateFromHeaders(provider: ProviderType, headers: Record<string, string>): void {
    const bucket = this.getBucket(provider);
    if (!bucket) return;

    // Parse provider-specific headers
    let requestsRemaining: number | null = null;
    let tokensRemaining: number | null = null;
    let resetDate: Date | null = null;

    switch (provider) {
      case "claude":
        requestsRemaining = parseInt(
          headers["anthropic-ratelimit-requests-remaining"] || "0"
        );
        tokensRemaining = parseInt(
          headers["anthropic-ratelimit-tokens-remaining"] || "0"
        );
        const resetTokens = headers["anthropic-ratelimit-reset-tokens"];
        if (resetTokens) {
          resetDate = new Date(resetTokens);
        }
        break;

      case "openai":
        requestsRemaining = parseInt(
          headers["x-ratelimit-remaining-requests"] || "0"
        );
        tokensRemaining = parseInt(headers["x-ratelimit-remaining-tokens"] || "0");
        const resetSeconds = parseInt(
          headers["x-ratelimit-reset-tokens"] || "0"
        );
        if (resetSeconds > 0) {
          resetDate = new Date(Date.now() + resetSeconds * 1000);
        }
        break;

      case "gemini":
        // Gemini provides less detailed rate limit info
        requestsRemaining = 999999;
        tokensRemaining = 999999;
        break;

      case "llama":
        // Llama/Groq rate limits
        requestsRemaining = parseInt(
          headers["x-ratelimit-remaining-requests"] || "0"
        );
        tokensRemaining = parseInt(headers["x-ratelimit-remaining-tokens"] || "0");
        break;
    }

    if (requestsRemaining !== null) {
      bucket.requestsUsed = bucket.requestsLimit - requestsRemaining;
    }

    if (tokensRemaining !== null) {
      bucket.tokensUsed = bucket.tokensLimit - tokensRemaining;
    }

    if (resetDate) {
      bucket.windowStart = new Date(Date.now() - 30000); // Assume we're 30s into window
    }
  }

  /**
   * Get time until rate limit reset
   */
  getTimeUntilReset(provider: ProviderType): number {
    const bucket = this.getBucket(provider);
    if (!bucket) return 0;

    const timeUntilReset = this.getTimeUntilWindowExpiry(bucket);
    return Math.max(0, timeUntilReset);
  }

  // ============================================================================
  // Private Methods
  // ============================================================================

  private getBucket(provider: ProviderType): RateLimitBucket | null {
    let bucket = this.buckets.get(provider);

    if (!bucket) {
      this.initializeProvider(provider);
      bucket = this.buckets.get(provider) || null;
    }

    return bucket;
  }

  private isWindowExpired(bucket: RateLimitBucket): boolean {
    const now = new Date();
    const elapsed = now.getTime() - bucket.windowStart.getTime();
    return elapsed >= bucket.windowDuration;
  }

  private resetBucket(bucket: RateLimitBucket): void {
    bucket.requestsUsed = 0;
    bucket.tokensUsed = 0;
    bucket.windowStart = new Date();
  }

  private getTimeUntilWindowExpiry(bucket: RateLimitBucket): number {
    const now = new Date();
    const elapsed = now.getTime() - bucket.windowStart.getTime();
    return Math.max(0, bucket.windowDuration - elapsed);
  }

  private calculateBackoffDelay(config: AdaptiveBackoffConfig, attempt: number): number {
    const exponentialDelay = config.initialDelayMs * Math.pow(config.multiplier, attempt - 1);
    const cappedDelay = Math.min(exponentialDelay, config.maxDelayMs);

    // Add jitter
    const jitter = cappedDelay * config.jitterFactor;
    const randomJitter = (Math.random() - 0.5) * 2 * jitter;

    return Math.max(0, cappedDelay + randomJitter);
  }

  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}

/**
 * Singleton instance
 */
export const rateLimiter = new RateLimiter();
