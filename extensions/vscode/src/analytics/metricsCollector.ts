/**
 * Metrics Collector
 * Collects and aggregates performance metrics
 */

import {
  AnalyticsEvent,
  PerformanceMetrics,
  UsageMetrics,
  MetricsAggregation,
} from '../types/analytics';

export class MetricsCollector {
  private metrics: Map<string, PerformanceMetrics> = new Map();

  /**
   * Record a metric
   */
  recordMetric(key: string, metric: Partial<PerformanceMetrics>): void {
    const current = this.metrics.get(key) || this.getDefaultMetrics();
    this.metrics.set(key, { ...current, ...metric } as PerformanceMetrics);
  }

  /**
   * Get a metric by key
   */
  getMetric(key: string): PerformanceMetrics {
    return this.metrics.get(key) || this.getDefaultMetrics();
  }

  /**
   * Aggregate metrics from events
   */
  aggregateMetrics(events: AnalyticsEvent[]): PerformanceMetrics {
    if (events.length === 0) {
      return this.getDefaultMetrics();
    }

    let totalResponseTime = 0;
    let totalInputTokens = 0;
    let totalOutputTokens = 0;
    let totalCost = 0;
    let cacheHits = 0;
    let cacheMisses = 0;
    let errors = 0;
    let totalMemory = 0;
    let totalCPU = 0;

    for (const event of events) {
      // Response time
      if (event.metadata.responseTime) {
        totalResponseTime += event.metadata.responseTime as number;
      }

      // Token usage
      if (event.metadata.inputTokens) {
        totalInputTokens += event.metadata.inputTokens as number;
      }
      if (event.metadata.outputTokens) {
        totalOutputTokens += event.metadata.outputTokens as number;
      }

      // Cost
      if (event.metadata.costUSD) {
        totalCost += event.metadata.costUSD as number;
      }

      // Cache metrics
      if (event.type === 'cache.hit') {
        cacheHits++;
      } else if (event.type === 'cache.miss') {
        cacheMisses++;
      }

      // Error tracking
      if (event.type.endsWith('.error')) {
        errors++;
      }

      // Memory and CPU
      if (event.metadata.memoryUsageMB) {
        totalMemory += event.metadata.memoryUsageMB as number;
      }
      if (event.metadata.cpuUsagePercent) {
        totalCPU += event.metadata.cpuUsagePercent as number;
      }
    }

    const cacheTotal = cacheHits + cacheMisses;
    const cacheHitRate = cacheTotal > 0 ? cacheHits / cacheTotal : 0;
    const errorRate = events.length > 0 ? errors / events.length : 0;
    const avgResponseTime = events.length > 0 ? totalResponseTime / events.length : 0;
    const avgMemory = events.length > 0 ? totalMemory / events.length : 0;
    const avgCPU = events.length > 0 ? totalCPU / events.length : 0;

    return {
      responseTime: avgResponseTime,
      tokenInput: totalInputTokens,
      tokenOutput: totalOutputTokens,
      costUSD: totalCost,
      cacheHitRate,
      errorRate,
      memoryUsageMB: avgMemory,
      cpuUsagePercent: avgCPU,
    };
  }

  /**
   * Calculate usage metrics
   */
  calculateUsageMetrics(events: AnalyticsEvent[]): UsageMetrics {
    const chatEvents = events.filter((e) => e.type === 'chat.receive');
    const ragEvents = events.filter((e) => e.type === 'rag.retrieve');
    const tokenEvents = events.filter((e) => e.type === 'token.usage');

    let totalTokens = 0;
    const providerBreakdown: Record<string, number> = {};
    const featureUsage: Record<string, number> = {};

    // Count tokens
    for (const event of tokenEvents) {
      const inputTokens = (event.metadata.inputTokens as number) || 0;
      const outputTokens = (event.metadata.outputTokens as number) || 0;
      totalTokens += inputTokens + outputTokens;

      const provider = event.metadata.provider as string;
      if (provider) {
        providerBreakdown[provider] = (providerBreakdown[provider] || 0) + 1;
      }
    }

    // Feature usage
    featureUsage['chat'] = chatEvents.length;
    featureUsage['rag'] = ragEvents.length;
    featureUsage['provider_switches'] = events.filter(
      (e) => e.type === 'provider.switch',
    ).length;

    // Session duration (from first to last event)
    let sessionDuration = 0;
    if (events.length > 0) {
      const sortedEvents = [...events].sort((a, b) => a.timestamp - b.timestamp);
      sessionDuration = sortedEvents[sortedEvents.length - 1].timestamp - sortedEvents[0].timestamp;
    }

    return {
      chatMessages: chatEvents.length,
      ragQueries: ragEvents.length,
      tokensUsed: totalTokens,
      providerBreakdown,
      featureUsage,
      sessionDuration,
    };
  }

  /**
   * Get top features
   */
  getTopFeatures(
    events: AnalyticsEvent[],
    limit: number = 5,
  ): Array<{ feature: string; count: number; percentage: number }> {
    const featureCounts: Record<string, number> = {};

    for (const event of events) {
      const feature = this.getFeatureFromEvent(event);
      if (feature) {
        featureCounts[feature] = (featureCounts[feature] || 0) + 1;
      }
    }

    const total = Object.values(featureCounts).reduce((a, b) => a + b, 0);

    return Object.entries(featureCounts)
      .map(([feature, count]) => ({
        feature,
        count,
        percentage: total > 0 ? (count / total) * 100 : 0,
      }))
      .sort((a, b) => b.count - a.count)
      .slice(0, limit);
  }

  /**
   * Get cost breakdown by provider
   */
  getCostBreakdown(
    events: AnalyticsEvent[],
  ): Array<{ provider: string; cost: number; percentage: number }> {
    const costByProvider: Record<string, number> = {};

    for (const event of events) {
      if (event.metadata.costUSD && event.metadata.provider) {
        const provider = event.metadata.provider as string;
        costByProvider[provider] = (costByProvider[provider] || 0) + (event.metadata.costUSD as number);
      }
    }

    const totalCost = Object.values(costByProvider).reduce((a, b) => a + b, 0);

    return Object.entries(costByProvider)
      .map(([provider, cost]) => ({
        provider,
        cost,
        percentage: totalCost > 0 ? (cost / totalCost) * 100 : 0,
      }))
      .sort((a, b) => b.cost - a.cost);
  }

  /**
   * Aggregate metrics over time
   */
  aggregateOverTime(
    events: AnalyticsEvent[],
    period: 'hour' | 'day' | 'week' | 'month',
  ): MetricsAggregation {
    const periodMs = this.getPeriodMs(period);
    const grouped: Map<number, AnalyticsEvent[]> = new Map();

    // Group events by period
    for (const event of events) {
      const periodStart = Math.floor(event.timestamp / periodMs) * periodMs;
      if (!grouped.has(periodStart)) {
        grouped.set(periodStart, []);
      }
      grouped.get(periodStart)!.push(event);
    }

    // Aggregate metrics for each period
    const data = Array.from(grouped.entries())
      .map(([timestamp, periodEvents]) => ({
        timestamp,
        metrics: this.aggregateMetrics(periodEvents),
      }))
      .sort((a, b) => a.timestamp - b.timestamp);

    return {
      period,
      data,
    };
  }

  /**
   * Calculate cache efficiency
   */
  calculateCacheEfficiency(events: AnalyticsEvent[]): {
    hitCount: number;
    missCount: number;
    hitRate: number;
    averageSavings: number; // ms
  } {
    const hitCount = events.filter((e) => e.type === 'cache.hit').length;
    const missCount = events.filter((e) => e.type === 'cache.miss').length;
    const total = hitCount + missCount;
    const hitRate = total > 0 ? hitCount / total : 0;

    // Estimate average savings (cache hits are typically 10-50x faster)
    const avgResponseTime =
      events
        .filter((e) => e.metadata.responseTime)
        .reduce((sum, e) => sum + ((e.metadata.responseTime as number) || 0), 0) /
      Math.max(events.filter((e) => e.metadata.responseTime).length, 1);

    const averageSavings = Math.round(avgResponseTime * 0.7); // Assume 70% faster

    return {
      hitCount,
      missCount,
      hitRate,
      averageSavings,
    };
  }

  /**
   * Get provider performance
   */
  getProviderPerformance(
    events: AnalyticsEvent[],
  ): Array<{
    provider: string;
    requestCount: number;
    errorCount: number;
    errorRate: number;
    averageResponseTime: number;
    averageCost: number;
  }> {
    const providerStats: Record<
      string,
      {
        requestCount: number;
        errorCount: number;
        totalResponseTime: number;
        totalCost: number;
      }
    > = {};

    for (const event of events) {
      if (!event.metadata.provider) continue;

      const provider = event.metadata.provider as string;
      if (!providerStats[provider]) {
        providerStats[provider] = {
          requestCount: 0,
          errorCount: 0,
          totalResponseTime: 0,
          totalCost: 0,
        };
      }

      providerStats[provider].requestCount++;

      if (event.type.endsWith('.error')) {
        providerStats[provider].errorCount++;
      }

      if (event.metadata.responseTime) {
        providerStats[provider].totalResponseTime += event.metadata.responseTime as number;
      }

      if (event.metadata.costUSD) {
        providerStats[provider].totalCost += event.metadata.costUSD as number;
      }
    }

    return Object.entries(providerStats)
      .map(([provider, stats]) => ({
        provider,
        requestCount: stats.requestCount,
        errorCount: stats.errorCount,
        errorRate: stats.requestCount > 0 ? stats.errorCount / stats.requestCount : 0,
        averageResponseTime:
          stats.requestCount > 0 ? stats.totalResponseTime / stats.requestCount : 0,
        averageCost: stats.requestCount > 0 ? stats.totalCost / stats.requestCount : 0,
      }))
      .sort((a, b) => b.requestCount - a.requestCount);
  }

  // Private methods

  private getDefaultMetrics(): PerformanceMetrics {
    return {
      responseTime: 0,
      tokenInput: 0,
      tokenOutput: 0,
      costUSD: 0,
      cacheHitRate: 0,
      errorRate: 0,
      memoryUsageMB: 0,
      cpuUsagePercent: 0,
    };
  }

  private getFeatureFromEvent(event: AnalyticsEvent): string | null {
    if (event.type.startsWith('chat.')) return 'chat';
    if (event.type.startsWith('rag.')) return 'rag';
    if (event.type.startsWith('provider.')) return 'providers';
    if (event.type === 'ui.action') return (event.metadata.action as string) || 'ui';
    if (event.type.startsWith('cache.')) return 'cache';
    return null;
  }

  private getPeriodMs(period: 'hour' | 'day' | 'week' | 'month'): number {
    switch (period) {
      case 'hour':
        return 60 * 60 * 1000;
      case 'day':
        return 24 * 60 * 60 * 1000;
      case 'week':
        return 7 * 24 * 60 * 60 * 1000;
      case 'month':
        return 30 * 24 * 60 * 60 * 1000;
    }
  }
}
