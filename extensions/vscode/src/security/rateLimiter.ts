/**
 * Rate Limiting Module - Advanced Rate Limiting & Throttling
 * Phase 12: Enterprise-Grade Security
 *
 * Provides:
 * - Token bucket algorithm
 * - Sliding window rate limiting
 * - Per-user, per-IP, per-endpoint limits
 * - Adaptive rate limiting
 * - DDoS detection
 * - Circuit breaker pattern
 */

/**
 * Rate limit configuration
 */
export interface RateLimitConfig {
  windowMs: number; // Time window in milliseconds
  maxRequests: number; // Max requests per window
  keyPrefix?: string; // Prefix for rate limit keys
  skipSuccessfulRequests?: boolean; // Don't count successful requests
  skipFailedRequests?: boolean; // Don't count failed requests
}

/**
 * Rate limit status
 */
export interface RateLimitStatus {
  remaining: number;
  limit: number;
  resetTime: number;
  isLimited: boolean;
}

/**
 * Token bucket for rate limiting
 */
interface TokenBucket {
  tokens: number;
  lastRefill: number;
  capacity: number;
  refillRate: number;
}

/**
 * Rate limiter using token bucket algorithm
 */
export class TokenBucketRateLimiter {
  private buckets: Map<string, TokenBucket> = new Map();
  private config: Required<RateLimitConfig>;

  constructor(config: RateLimitConfig) {
    this.config = {
      windowMs: config.windowMs,
      maxRequests: config.maxRequests,
      keyPrefix: config.keyPrefix || 'rl',
      skipSuccessfulRequests: config.skipSuccessfulRequests ?? false,
      skipFailedRequests: config.skipFailedRequests ?? false
    };
  }

  /**
   * Check if request is allowed
   */
  isAllowed(key: string): RateLimitStatus {
    const bucketKey = `${this.config.keyPrefix}:${key}`;
    let bucket = this.buckets.get(bucketKey);

    if (!bucket) {
      bucket = {
        tokens: this.config.maxRequests,
        lastRefill: Date.now(),
        capacity: this.config.maxRequests,
        refillRate: this.config.maxRequests / (this.config.windowMs / 1000)
      };
      this.buckets.set(bucketKey, bucket);
    }

    // Refill tokens based on elapsed time
    const now = Date.now();
    const timePassed = (now - bucket.lastRefill) / 1000;
    const tokensToAdd = timePassed * bucket.refillRate;
    bucket.tokens = Math.min(bucket.capacity, bucket.tokens + tokensToAdd);
    bucket.lastRefill = now;

    const isLimited = bucket.tokens < 1;
    const remaining = Math.floor(bucket.tokens);

    if (!isLimited) {
      bucket.tokens -= 1;
    }

    return {
      remaining: Math.max(0, remaining - 1),
      limit: this.config.maxRequests,
      resetTime: now + ((1 - bucket.tokens) / bucket.refillRate) * 1000,
      isLimited
    };
  }

  /**
   * Reset rate limit for key
   */
  reset(key: string): void {
    const bucketKey = `${this.config.keyPrefix}:${key}`;
    this.buckets.delete(bucketKey);
  }

  /**
   * Reset all rate limits
   */
  resetAll(): void {
    this.buckets.clear();
  }

  /**
   * Get bucket status (for monitoring)
   */
  getStatus(key: string): RateLimitStatus | null {
    const bucketKey = `${this.config.keyPrefix}:${key}`;
    const bucket = this.buckets.get(bucketKey);

    if (!bucket) {
      return null;
    }

    const now = Date.now();
    const timePassed = (now - bucket.lastRefill) / 1000;
    const tokensToAdd = timePassed * bucket.refillRate;
    const currentTokens = Math.min(bucket.capacity, bucket.tokens + tokensToAdd);

    return {
      remaining: Math.floor(currentTokens),
      limit: this.config.maxRequests,
      resetTime: now + ((bucket.capacity - currentTokens) / bucket.refillRate) * 1000,
      isLimited: currentTokens < 1
    };
  }
}

/**
 * Sliding window rate limiter
 */
export class SlidingWindowRateLimiter {
  private windows: Map<string, number[]> = new Map();
  private config: Required<RateLimitConfig>;

  constructor(config: RateLimitConfig) {
    this.config = {
      windowMs: config.windowMs,
      maxRequests: config.maxRequests,
      keyPrefix: config.keyPrefix || 'sw_rl',
      skipSuccessfulRequests: config.skipSuccessfulRequests ?? false,
      skipFailedRequests: config.skipFailedRequests ?? false
    };
  }

  /**
   * Check if request is allowed
   */
  isAllowed(key: string): RateLimitStatus {
    const windowKey = `${this.config.keyPrefix}:${key}`;
    const now = Date.now();
    const windowStart = now - this.config.windowMs;

    let timestamps = this.windows.get(windowKey) || [];

    // Remove old timestamps outside the window
    timestamps = timestamps.filter(t => t > windowStart);

    const isLimited = timestamps.length >= this.config.maxRequests;

    if (!isLimited) {
      timestamps.push(now);
    }

    this.windows.set(windowKey, timestamps);

    return {
      remaining: Math.max(0, this.config.maxRequests - timestamps.length),
      limit: this.config.maxRequests,
      resetTime: timestamps.length > 0 ? timestamps[0] + this.config.windowMs : now + this.config.windowMs,
      isLimited
    };
  }

  /**
   * Reset rate limit for key
   */
  reset(key: string): void {
    const windowKey = `${this.config.keyPrefix}:${key}`;
    this.windows.delete(windowKey);
  }

  /**
   * Reset all rate limits
   */
  resetAll(): void {
    this.windows.clear();
  }

  /**
   * Get window status (for monitoring)
   */
  getStatus(key: string): RateLimitStatus | null {
    const windowKey = `${this.config.keyPrefix}:${key}`;
    const timestamps = this.windows.get(windowKey);

    if (!timestamps) {
      return null;
    }

    const now = Date.now();
    return {
      remaining: Math.max(0, this.config.maxRequests - timestamps.length),
      limit: this.config.maxRequests,
      resetTime: timestamps.length > 0 ? timestamps[0] + this.config.windowMs : now + this.config.windowMs,
      isLimited: timestamps.length >= this.config.maxRequests
    };
  }
}

/**
 * Adaptive rate limiter (adjusts limits based on server load)
 */
export class AdaptiveRateLimiter {
  private baseLimiter: TokenBucketRateLimiter;
  private cpuUsage: number = 0;
  private memoryUsage: number = 0;
  private baseLimit: number;

  constructor(config: RateLimitConfig) {
    this.baseLimit = config.maxRequests;
    this.baseLimiter = new TokenBucketRateLimiter(config);
  }

  /**
   * Set current system metrics
   */
  setSystemMetrics(cpuUsage: number, memoryUsage: number): void {
    this.cpuUsage = cpuUsage;
    this.memoryUsage = memoryUsage;
  }

  /**
   * Check if request is allowed (with adaptive limiting)
   */
  isAllowed(key: string): RateLimitStatus {
    const status = this.baseLimiter.isAllowed(key);

    // Adaptive limit based on system metrics
    const loadFactor = Math.max(this.cpuUsage, this.memoryUsage) / 100;
    const adaptiveLimit = Math.ceil(this.baseLimit * (1 - loadFactor * 0.5));

    return {
      ...status,
      limit: adaptiveLimit,
      isLimited: status.isLimited || status.remaining < adaptiveLimit
    };
  }

  /**
   * Reset rate limit for key
   */
  reset(key: string): void {
    this.baseLimiter.reset(key);
  }

  /**
   * Reset all rate limits
   */
  resetAll(): void {
    this.baseLimiter.resetAll();
  }
}

/**
 * DDoS detector using rate limit metrics
 */
export class DDoSDetector {
  private requestCounts: Map<string, number> = new Map();
  private suspiciousIPs: Set<string> = new Set();
  private threshold: number;
  private timeWindow: number;

  constructor(threshold: number = 100, timeWindowMs: number = 60000) {
    this.threshold = threshold;
    this.timeWindow = timeWindowMs;

    // Cleanup old entries periodically
    setInterval(() => this.cleanup(), this.timeWindow);
  }

  /**
   * Detect potential DDoS attack
   */
  detect(ip: string): boolean {
    const count = (this.requestCounts.get(ip) || 0) + 1;
    this.requestCounts.set(ip, count);

    const isSuspicious = count > this.threshold;

    if (isSuspicious) {
      this.suspiciousIPs.add(ip);
    }

    return isSuspicious;
  }

  /**
   * Check if IP is suspicious
   */
  isSuspicious(ip: string): boolean {
    return this.suspiciousIPs.has(ip);
  }

  /**
   * Get suspicious IPs
   */
  getSuspiciousIPs(): string[] {
    return Array.from(this.suspiciousIPs);
  }

  /**
   * Clear suspicious IP
   */
  clearSuspicious(ip: string): void {
    this.suspiciousIPs.delete(ip);
  }

  /**
   * Cleanup old entries
   */
  private cleanup(): void {
    this.requestCounts.clear();
  }

  /**
   * Reset detector
   */
  reset(): void {
    this.requestCounts.clear();
    this.suspiciousIPs.clear();
  }
}

/**
 * Circuit breaker pattern for resilience
 */
export class CircuitBreaker {
  private state: 'closed' | 'open' | 'half-open' = 'closed';
  private failureCount: number = 0;
  private successCount: number = 0;
  private lastFailureTime: number = 0;
  private threshold: number;
  private timeout: number;

  constructor(failureThreshold: number = 5, timeoutMs: number = 60000) {
    this.threshold = failureThreshold;
    this.timeout = timeoutMs;
  }

  /**
   * Get current state
   */
  getState(): 'closed' | 'open' | 'half-open' {
    if (this.state === 'open' && Date.now() - this.lastFailureTime > this.timeout) {
      this.state = 'half-open';
      this.successCount = 0;
    }
    return this.state;
  }

  /**
   * Record success
   */
  recordSuccess(): void {
    this.failureCount = 0;

    if (this.state === 'half-open') {
      this.successCount++;
      if (this.successCount >= 2) {
        this.state = 'closed';
      }
    }
  }

  /**
   * Record failure
   */
  recordFailure(): void {
    this.failureCount++;
    this.lastFailureTime = Date.now();

    if (this.failureCount >= this.threshold) {
      this.state = 'open';
    }
  }

  /**
   * Check if operation is allowed
   */
  isAllowed(): boolean {
    const state = this.getState();
    return state !== 'open';
  }

  /**
   * Reset circuit breaker
   */
  reset(): void {
    this.state = 'closed';
    this.failureCount = 0;
    this.successCount = 0;
    this.lastFailureTime = 0;
  }
}

/**
 * Rate limiter manager for multiple endpoints
 */
export class RateLimiterManager {
  private limiters: Map<string, TokenBucketRateLimiter> = new Map();
  private defaultConfig: RateLimitConfig;

  constructor(defaultConfig: RateLimitConfig) {
    this.defaultConfig = defaultConfig;
  }

  /**
   * Create limiter for endpoint
   */
  createLimiter(endpoint: string, config?: RateLimitConfig): TokenBucketRateLimiter {
    if (!this.limiters.has(endpoint)) {
      const finalConfig = config || this.defaultConfig;
      this.limiters.set(endpoint, new TokenBucketRateLimiter(finalConfig));
    }
    return this.limiters.get(endpoint)!;
  }

  /**
   * Check if request is allowed
   */
  isAllowed(endpoint: string, key: string): RateLimitStatus {
    const limiter = this.createLimiter(endpoint);
    return limiter.isAllowed(key);
  }

  /**
   * Reset limiter for endpoint
   */
  resetEndpoint(endpoint: string): void {
    this.limiters.delete(endpoint);
  }

  /**
   * Reset all limiters
   */
  resetAll(): void {
    this.limiters.clear();
  }

  /**
   * Get limiter for endpoint
   */
  getLimiter(endpoint: string): TokenBucketRateLimiter | null {
    return this.limiters.get(endpoint) || null;
  }
}

export default TokenBucketRateLimiter;
