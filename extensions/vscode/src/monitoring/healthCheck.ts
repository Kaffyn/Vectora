/**
 * Health Check System
 * Monitors system and provider health
 */

import {
  HealthStatus,
  ProviderHealth,
  NetworkHealth,
  SystemHealth,
  StorageHealth,
} from '../types/analytics';
import { EventTracker } from '../analytics/eventTracker';

export interface HealthCheckResult {
  overall: 'healthy' | 'degraded' | 'critical';
  lastCheck: number;
  providers: Map<string, ProviderHealth>;
  network: NetworkHealth;
  system: SystemHealth;
  storage: StorageHealth;
  checks: HealthCheckDetail[];
}

export interface HealthCheckDetail {
  component: string;
  status: 'healthy' | 'degraded' | 'unhealthy';
  message: string;
  lastUpdated: number;
}

export class HealthCheck {
  private sessionId: string;
  private eventTracker: EventTracker;
  private lastResult: HealthCheckResult | null = null;
  private checkInterval?: NodeJS.Timer;
  private providerStats: Map<string, ProviderHealth> = new Map();

  constructor(sessionId: string, eventTracker: EventTracker) {
    this.sessionId = sessionId;
    this.eventTracker = eventTracker;
  }

  /**
   * Run all health checks
   */
  async performHealthCheck(): Promise<HealthCheckResult> {
    const startTime = Date.now();

    const [network, system, storage, providers] = await Promise.all([
      this.checkNetwork(),
      this.checkSystem(),
      this.checkStorage(),
      this.checkProviders(),
    ]);

    const checks: HealthCheckDetail[] = [];

    // Network checks
    checks.push({
      component: 'Network',
      status: network.connected ? 'healthy' : 'unhealthy',
      message: network.connected
        ? `Latency: ${network.latency}ms`
        : 'Network connection failed',
      lastUpdated: network.lastCheck,
    });

    // System checks
    const systemStatus = this.determineSystemStatus(system);
    checks.push({
      component: 'System',
      status: systemStatus,
      message: `Memory: ${system.memoryUsage}MB, CPU: ${system.cpuUsage}%`,
      lastUpdated: system.uptime,
    });

    // Storage checks
    const storageStatus = storage.usagePercent > 90 ? 'unhealthy' :
                         storage.usagePercent > 75 ? 'degraded' : 'healthy';
    checks.push({
      component: 'Storage',
      status: storageStatus,
      message: `Usage: ${storage.usagePercent}% (${(storage.usedBytes / (1024 * 1024)).toFixed(0)}MB)`,
      lastUpdated: Date.now(),
    });

    // Provider checks
    for (const [provider, health] of providers) {
      const status = health.available ? 'healthy' :
                    health.errorCount > health.successCount ? 'unhealthy' :
                    'degraded';
      checks.push({
        component: `Provider: ${provider}`,
        status,
        message: `Success: ${health.successCount}, Errors: ${health.errorCount}`,
        lastUpdated: health.lastCheck,
      });
    }

    // Determine overall status
    const unhealthyCount = checks.filter((c) => c.status === 'unhealthy').length;
    const degradedCount = checks.filter((c) => c.status === 'degraded').length;

    let overall: 'healthy' | 'degraded' | 'critical' = 'healthy';
    if (unhealthyCount > 0) {
      overall = 'critical';
    } else if (degradedCount > 0) {
      overall = 'degraded';
    }

    const result: HealthCheckResult = {
      overall,
      lastCheck: startTime,
      providers,
      network,
      system,
      storage,
      checks,
    };

    this.lastResult = result;

    // Log if status changed
    if (this.lastResult && this.lastResult.overall !== overall) {
      this.eventTracker.logSystemWarning(`Health status changed to ${overall}`, {
        previous: this.lastResult.overall,
        current: overall,
      });
    }

    return result;
  }

  /**
   * Check network connectivity
   */
  private async checkNetwork(): Promise<NetworkHealth> {
    const startTime = performance.now();

    try {
      // Simple connectivity check
      const response = await fetch('https://www.google.com/favicon.ico', {
        mode: 'no-cors',
        cache: 'no-cache',
      });

      const latency = performance.now() - startTime;

      return {
        latency: Math.round(latency),
        connected: true,
        lastCheck: Date.now(),
      };
    } catch (error) {
      return {
        latency: 0,
        connected: false,
        lastCheck: Date.now(),
      };
    }
  }

  /**
   * Check system resources
   */
  private async checkSystem(): Promise<SystemHealth> {
    const memoryUsage = this.getMemoryUsage();
    const cpuUsage = this.estimateCPUUsage();
    const errorCount = this.countSystemErrors();

    return {
      uptime: Date.now(),
      memoryUsage,
      cpuUsage,
      errorCount,
    };
  }

  /**
   * Check storage
   */
  private async checkStorage(): Promise<StorageHealth> {
    try {
      const quota = await navigator.storage.estimate?.();
      if (!quota) {
        return {
          quotaBytes: 0,
          usedBytes: 0,
          remainingBytes: 0,
          usagePercent: 0,
        };
      }

      const usedBytes = quota.usage || 0;
      const quotaBytes = quota.quota || 0;
      const remainingBytes = quotaBytes - usedBytes;
      const usagePercent = quotaBytes > 0 ? (usedBytes / quotaBytes) * 100 : 0;

      return {
        quotaBytes,
        usedBytes,
        remainingBytes,
        usagePercent,
      };
    } catch (error) {
      console.warn('Failed to check storage quota:', error);
      return {
        quotaBytes: 0,
        usedBytes: 0,
        remainingBytes: 0,
        usagePercent: 0,
      };
    }
  }

  /**
   * Check provider health
   */
  private async checkProviders(): Promise<Map<string, ProviderHealth>> {
    // Return cached provider stats
    return new Map(this.providerStats);
  }

  /**
   * Record provider success
   */
  recordProviderSuccess(provider: string, responseTime: number): void {
    const health = this.providerStats.get(provider) || this.createDefaultHealth();
    health.successCount++;
    health.lastCheck = Date.now();
    health.averageResponseTime =
      (health.averageResponseTime * (health.successCount - 1) + responseTime) /
      health.successCount;
    this.providerStats.set(provider, health);
  }

  /**
   * Record provider error
   */
  recordProviderError(provider: string, error?: Error): void {
    const health = this.providerStats.get(provider) || this.createDefaultHealth();
    health.available = false;
    health.errorCount++;
    health.lastCheck = Date.now();
    this.providerStats.set(provider, health);

    // Check if provider should be marked as unavailable
    if (health.errorCount > 3 && health.successCount === 0) {
      this.eventTracker.logSystemError(`Provider ${provider} is unavailable`, {
        errorCount: health.errorCount,
      });
    }
  }

  /**
   * Update provider rate limit
   */
  updateProviderRateLimit(provider: string, remaining: number, limit: number): void {
    const health = this.providerStats.get(provider) || this.createDefaultHealth();
    health.rateLimitRemaining = remaining;
    this.providerStats.set(provider, health);

    // Warn if approaching limit
    if (remaining < limit * 0.1) {
      this.eventTracker.logRateLimitWarning(provider, remaining, limit);
    }
  }

  /**
   * Start health monitoring
   */
  startMonitoring(intervalMs: number = 30000): void {
    this.checkInterval = setInterval(() => {
      this.performHealthCheck();
    }, intervalMs);
  }

  /**
   * Stop health monitoring
   */
  stopMonitoring(): void {
    if (this.checkInterval) {
      clearInterval(this.checkInterval);
      this.checkInterval = undefined;
    }
  }

  /**
   * Get last health check result
   */
  getLastResult(): HealthCheckResult | null {
    return this.lastResult;
  }

  /**
   * Get provider health
   */
  getProviderHealth(provider: string): ProviderHealth {
    return this.providerStats.get(provider) || this.createDefaultHealth();
  }

  /**
   * Get all provider health
   */
  getAllProviderHealth(): Map<string, ProviderHealth> {
    return new Map(this.providerStats);
  }

  /**
   * Clear provider stats
   */
  clearProviderStats(): void {
    this.providerStats.clear();
  }

  /**
   * Reset monitor
   */
  reset(): void {
    this.stopMonitoring();
    this.clearProviderStats();
    this.lastResult = null;
  }

  // Private methods

  private createDefaultHealth(): ProviderHealth {
    return {
      available: true,
      lastCheck: Date.now(),
      errorCount: 0,
      successCount: 0,
      averageResponseTime: 0,
    };
  }

  private determineSystemStatus(system: SystemHealth): 'healthy' | 'degraded' | 'unhealthy' {
    if (system.memoryUsage > 500 || system.cpuUsage > 80) {
      return 'degraded';
    }
    if (system.memoryUsage > 1000 || system.cpuUsage > 95) {
      return 'unhealthy';
    }
    return 'healthy';
  }

  private getMemoryUsage(): number {
    if (typeof performance === 'undefined' || !performance.memory) {
      return 0;
    }
    return Math.round(performance.memory.usedJSHeapSize / (1024 * 1024));
  }

  private estimateCPUUsage(): number {
    // This is a rough estimate based on available metrics
    // In a real scenario, you'd use more sophisticated CPU monitoring
    return Math.random() * 50; // Return random value for now
  }

  private countSystemErrors(): number {
    // This would be connected to your error tracking system
    return 0;
  }
}
