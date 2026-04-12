/**
 * Fallback Strategy - Phase 9
 * Estratégia de fallback automático entre providers
 */

import { ProviderType } from "../../providers/baseProvider";

// ============================================================================
// Types
// ============================================================================

export interface FallbackConfig {
  enabled: boolean;
  chain: ProviderType[];
  maxRetries: number;
  retryDelayMs: number;
  exponentialBackoff: boolean;
}

export interface FallbackAttempt {
  provider: ProviderType;
  attempt: number;
  timestamp: Date;
  error?: Error;
  success: boolean;
  durationMs: number;
}

export interface FallbackResult<T> {
  success: boolean;
  data?: T;
  lastError?: Error;
  attempts: FallbackAttempt[];
  totalDurationMs: number;
  successProvider: ProviderType | null;
}

export interface ProviderHealth {
  provider: ProviderType;
  successRate: number;
  lastError?: Error;
  lastErrorTime?: Date;
  failureCount: number;
  successCount: number;
  avgResponseTimeMs: number;
  isHealthy: boolean;
}

// ============================================================================
// Fallback Strategy Implementation
// ============================================================================

export class FallbackStrategy {
  private config: FallbackConfig;
  private healthMetrics: Map<ProviderType, ProviderHealth> = new Map();
  private requestHistory: FallbackAttempt[] = [];
  private historyLimit: number = 1000;

  constructor(config: Partial<FallbackConfig> = {}) {
    this.config = {
      enabled: config.enabled !== false,
      chain: config.chain || ["claude", "openai", "gemini", "llama"],
      maxRetries: config.maxRetries || 3,
      retryDelayMs: config.retryDelayMs || 1000,
      exponentialBackoff: config.exponentialBackoff !== false,
    };
  }

  /**
   * Execute operation with fallback
   */
  async executeWithFallback<T>(
    operation: (provider: ProviderType) => Promise<T>,
    providers?: ProviderType[]
  ): Promise<FallbackResult<T>> {
    if (!this.config.enabled || !providers) {
      providers = this.config.chain;
    }

    const startTime = Date.now();
    const attempts: FallbackAttempt[] = [];
    let lastError: Error | undefined;
    let successProvider: ProviderType | null = null;

    for (let retryAttempt = 0; retryAttempt < this.config.maxRetries; retryAttempt++) {
      for (const provider of providers) {
        const attemptStartTime = Date.now();

        try {
          const result = await operation(provider);

          const durationMs = Date.now() - attemptStartTime;
          const attempt: FallbackAttempt = {
            provider,
            attempt: retryAttempt + 1,
            timestamp: new Date(),
            success: true,
            durationMs,
          };

          attempts.push(attempt);
          this.recordSuccess(provider, durationMs);
          successProvider = provider;

          return {
            success: true,
            data: result,
            attempts,
            totalDurationMs: Date.now() - startTime,
            successProvider,
          };
        } catch (error) {
          const durationMs = Date.now() - attemptStartTime;
          const err = error instanceof Error ? error : new Error(String(error));

          lastError = err;

          const attempt: FallbackAttempt = {
            provider,
            attempt: retryAttempt + 1,
            timestamp: new Date(),
            error: err,
            success: false,
            durationMs,
          };

          attempts.push(attempt);
          this.recordFailure(provider, err);

          // Wait before next attempt if not the last attempt
          if (retryAttempt < this.config.maxRetries - 1 || provider !== providers[providers.length - 1]) {
            const delayMs = this.getRetryDelay(retryAttempt);
            await this.sleep(delayMs);
          }
        }
      }
    }

    return {
      success: false,
      lastError,
      attempts,
      totalDurationMs: Date.now() - startTime,
      successProvider: null,
    };
  }

  /**
   * Get optimal provider based on health metrics
   */
  getBestProvider(providers: ProviderType[]): ProviderType | null {
    if (!providers.length) return null;

    // Filter healthy providers first
    const healthyProviders = providers.filter((p) => {
      const health = this.healthMetrics.get(p);
      return !health || health.isHealthy;
    });

    if (!healthyProviders.length) {
      // Fall back to any provider if all are unhealthy
      return providers[0];
    }

    // Sort by success rate and response time
    return healthyProviders.sort((a, b) => {
      const healthA = this.getOrCreateHealth(a);
      const healthB = this.getOrCreateHealth(b);

      const scoreA = healthA.successRate / (healthA.avgResponseTimeMs / 1000 + 1);
      const scoreB = healthB.successRate / (healthB.avgResponseTimeMs / 1000 + 1);

      return scoreB - scoreA;
    })[0];
  }

  /**
   * Get optimized fallback chain
   */
  getOptimizedChain(providers: ProviderType[]): ProviderType[] {
    return providers.sort((a, b) => {
      const healthA = this.getOrCreateHealth(a);
      const healthB = this.getOrCreateHealth(b);

      // Primary sort: success rate (higher is better)
      const rateCompare = healthB.successRate - healthA.successRate;
      if (rateCompare !== 0) return rateCompare;

      // Secondary sort: response time (lower is better)
      return healthA.avgResponseTimeMs - healthB.avgResponseTimeMs;
    });
  }

  /**
   * Get health of a provider
   */
  getHealth(provider: ProviderType): ProviderHealth | null {
    return this.healthMetrics.get(provider) || null;
  }

  /**
   * Get all provider healths
   */
  getAllHealths(): ProviderHealth[] {
    return Array.from(this.healthMetrics.values());
  }

  /**
   * Reset health metrics for provider
   */
  resetHealth(provider: ProviderType): void {
    this.healthMetrics.delete(provider);
  }

  /**
   * Reset all health metrics
   */
  resetAllHealth(): void {
    this.healthMetrics.clear();
  }

  /**
   * Get request history
   */
  getHistory(): FallbackAttempt[] {
    return [...this.requestHistory];
  }

  /**
   * Get history for specific provider
   */
  getHistoryFor(provider: ProviderType): FallbackAttempt[] {
    return this.requestHistory.filter((a) => a.provider === provider);
  }

  /**
   * Clear request history
   */
  clearHistory(): void {
    this.requestHistory = [];
  }

  /**
   * Update fallback chain
   */
  setChain(chain: ProviderType[]): void {
    this.config.chain = chain;
  }

  /**
   * Set max retries
   */
  setMaxRetries(retries: number): void {
    this.config.maxRetries = retries;
  }

  /**
   * Enable/disable fallback
   */
  setEnabled(enabled: boolean): void {
    this.config.enabled = enabled;
  }

  /**
   * Get current config
   */
  getConfig(): FallbackConfig {
    return { ...this.config };
  }

  /**
   * Generate report
   */
  generateReport(): {
    config: FallbackConfig;
    health: ProviderHealth[];
    history: {
      total: number;
      successful: number;
      failed: number;
      successRate: number;
    };
  } {
    const health = this.getAllHealths();
    const successful = this.requestHistory.filter((a) => a.success).length;
    const failed = this.requestHistory.filter((a) => !a.success).length;
    const total = this.requestHistory.length;

    return {
      config: this.getConfig(),
      health,
      history: {
        total,
        successful,
        failed,
        successRate: total > 0 ? (successful / total) * 100 : 0,
      },
    };
  }

  // ============================================================================
  // Private Methods
  // ============================================================================

  private recordSuccess(provider: ProviderType, responseTimeMs: number): void {
    const health = this.getOrCreateHealth(provider);

    health.successCount += 1;
    health.lastError = undefined;
    health.lastErrorTime = undefined;
    health.failureCount = Math.max(0, health.failureCount - 1); // Reduce failure count on success

    // Update average response time
    const totalTime =
      health.avgResponseTimeMs * (health.successCount - 1) + responseTimeMs;
    health.avgResponseTimeMs = totalTime / health.successCount;

    // Update success rate
    const total = health.successCount + health.failureCount;
    health.successRate = (health.successCount / total) * 100;

    // Provider is healthy if success rate > 70%
    health.isHealthy = health.successRate > 70 && health.failureCount < 3;

    this.healthMetrics.set(provider, health);
  }

  private recordFailure(provider: ProviderType, error: Error): void {
    const health = this.getOrCreateHealth(provider);

    health.failureCount += 1;
    health.lastError = error;
    health.lastErrorTime = new Date();

    // Update success rate
    const total = health.successCount + health.failureCount;
    health.successRate = (health.successCount / total) * 100;

    // Provider is unhealthy if success rate < 50% or too many consecutive failures
    health.isHealthy = health.successRate >= 50 && health.failureCount < 5;

    this.healthMetrics.set(provider, health);
  }

  private getOrCreateHealth(provider: ProviderType): ProviderHealth {
    let health = this.healthMetrics.get(provider);

    if (!health) {
      health = {
        provider,
        successRate: 100,
        failureCount: 0,
        successCount: 0,
        avgResponseTimeMs: 0,
        isHealthy: true,
      };
      this.healthMetrics.set(provider, health);
    }

    return health;
  }

  private getRetryDelay(attempt: number): number {
    if (!this.config.exponentialBackoff) {
      return this.config.retryDelayMs;
    }

    const exponentialDelay = this.config.retryDelayMs * Math.pow(2, attempt);
    const maxDelay = this.config.retryDelayMs * 32; // Max ~30s for retries
    const cappedDelay = Math.min(exponentialDelay, maxDelay);

    // Add jitter (±10%)
    const jitter = cappedDelay * 0.1;
    const randomJitter = (Math.random() - 0.5) * 2 * jitter;

    return Math.max(0, cappedDelay + randomJitter);
  }

  private recordAttempt(attempt: FallbackAttempt): void {
    this.requestHistory.push(attempt);

    // Keep history size manageable
    if (this.requestHistory.length > this.historyLimit) {
      this.requestHistory = this.requestHistory.slice(-this.historyLimit);
    }
  }

  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}

/**
 * Singleton instance
 */
export const fallbackStrategy = new FallbackStrategy({
  chain: ["claude", "openai", "gemini", "llama"],
  maxRetries: 3,
  exponentialBackoff: true,
});
