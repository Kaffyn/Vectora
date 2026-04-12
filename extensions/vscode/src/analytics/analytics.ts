/**
 * Analytics Engine
 * Core analytics system for Vectora
 */

import { v4 as uuidv4 } from 'uuid';
import {
  AnalyticsEvent,
  EventType,
  AnalyticsConfig,
  AnalyticsStore,
  EventFilter,
  PerformanceMetrics,
  Alert,
  AlertRule,
  AlertCondition,
  HealthStatus,
  DashboardData,
  Recommendation,
} from '../types/analytics';
import { EventTracker } from './eventTracker';
import { MetricsCollector } from './metricsCollector';

export class Analytics {
  private static instance: Analytics;
  private sessionId: string;
  private store: AnalyticsStore;
  private config: AnalyticsConfig;
  private eventTracker: EventTracker;
  private metricsCollector: MetricsCollector;
  private alertRules: Map<string, AlertRule>;
  private flushInterval?: NodeJS.Timer;
  private startTime: number;

  private constructor(config: Partial<AnalyticsConfig> = {}) {
    this.sessionId = uuidv4();
    this.startTime = Date.now();

    // Default configuration
    this.config = {
      enabled: true,
      trackingLevel: 'detailed',
      retentionDays: 30,
      sessionIdStrategy: 'session',
      privacyLevel: 'normal',
      batchSize: 50,
      flushIntervalMs: 30000, // 30 seconds
      storageBackend: 'both',
      exportFormats: ['json', 'csv'],
      ...config,
    };

    this.store = {
      events: [],
      metrics: new Map(),
      alerts: [],
      health: this.getDefaultHealth(),
      lastFlush: Date.now(),
    };

    this.alertRules = new Map();
    this.eventTracker = new EventTracker(this.sessionId);
    this.metricsCollector = new MetricsCollector();

    // Initialize default alert rules
    this.initializeDefaultAlerts();

    // Start auto-flush
    if (this.config.enabled) {
      this.startAutoFlush();
    }
  }

  static getInstance(config?: Partial<AnalyticsConfig>): Analytics {
    if (!Analytics.instance) {
      Analytics.instance = new Analytics(config);
    }
    return Analytics.instance;
  }

  /**
   * Track an analytics event
   */
  trackEvent(type: EventType, metadata: Record<string, unknown>): AnalyticsEvent {
    if (!this.config.enabled) {
      return this.createDummyEvent(type, metadata);
    }

    const event = this.eventTracker.createEvent(type, metadata, this.sessionId);
    this.store.events.push(event);

    // Check alert conditions
    this.checkAlertConditions(event);

    // Flush if batch size reached
    if (this.store.events.length >= this.config.batchSize) {
      this.flush();
    }

    return event;
  }

  /**
   * Track chat event
   */
  trackChat(data: {
    messageLength: number;
    tokenCount?: number;
    provider?: string;
    model?: string;
    responseTime?: number;
    error?: string;
  }): AnalyticsEvent {
    const type: EventType = data.error ? 'chat.error' : 'chat.receive';
    return this.trackEvent(type, data);
  }

  /**
   * Track RAG event
   */
  trackRAG(data: {
    query: string;
    resultCount: number;
    relevanceScore?: number;
    executionTime: number;
    cacheHit?: boolean;
  }): AnalyticsEvent {
    return this.trackEvent('rag.retrieve', data);
  }

  /**
   * Track provider event
   */
  trackProvider(data: {
    provider: string;
    previousProvider?: string;
    type: 'switch' | 'fallback' | 'error';
    reason?: string;
    error?: string;
  }): AnalyticsEvent {
    const type: EventType = `provider.${data.type}` as EventType;
    return this.trackEvent(type, {
      provider: data.provider,
      previousProvider: data.previousProvider,
      reason: data.reason,
      error: data.error,
    });
  }

  /**
   * Record performance metrics
   */
  recordMetrics(key: string, metrics: Partial<PerformanceMetrics>): void {
    const current = this.store.metrics.get(key) || this.getDefaultMetrics();
    this.store.metrics.set(key, { ...current, ...metrics } as PerformanceMetrics);
  }

  /**
   * Get metrics for a period
   */
  getMetrics(
    startTime: number,
    endTime: number,
  ): PerformanceMetrics {
    const events = this.getEvents({
      startTime,
      endTime,
    });

    return this.metricsCollector.aggregateMetrics(events);
  }

  /**
   * Get events matching filter
   */
  getEvents(filter?: EventFilter): AnalyticsEvent[] {
    if (!filter) return this.store.events;

    return this.store.events.filter((event) => {
      if (filter.startTime && event.timestamp < filter.startTime) return false;
      if (filter.endTime && event.timestamp > filter.endTime) return false;
      if (filter.types && !filter.types.includes(event.type)) return false;
      if (filter.sessionId && event.sessionId !== filter.sessionId) return false;
      if (filter.userId && event.userId !== filter.userId) return false;
      return true;
    });
  }

  /**
   * Create an alert
   */
  createAlert(
    ruleId: string,
    severity: 'info' | 'warning' | 'error' | 'critical',
    title: string,
    message: string,
  ): Alert {
    const alert: Alert = {
      id: uuidv4(),
      ruleId,
      severity,
      title,
      message,
      timestamp: Date.now(),
      resolved: false,
    };

    this.store.alerts.push(alert);
    this.eventTracker.logAlert(alert);

    return alert;
  }

  /**
   * Resolve an alert
   */
  resolveAlert(alertId: string): void {
    const alert = this.store.alerts.find((a) => a.id === alertId);
    if (alert) {
      alert.resolved = true;
      alert.resolvedAt = Date.now();
    }
  }

  /**
   * Get active alerts
   */
  getActiveAlerts(): Alert[] {
    return this.store.alerts.filter((a) => !a.resolved);
  }

  /**
   * Add alert rule
   */
  addAlertRule(rule: AlertRule): void {
    this.alertRules.set(rule.id, rule);
  }

  /**
   * Remove alert rule
   */
  removeAlertRule(ruleId: string): void {
    this.alertRules.delete(ruleId);
  }

  /**
   * Get health status
   */
  getHealth(): HealthStatus {
    return this.store.health;
  }

  /**
   * Update health status
   */
  updateHealth(health: Partial<HealthStatus>): void {
    this.store.health = {
      ...this.store.health,
      ...health,
      timestamp: Date.now(),
    };
  }

  /**
   * Get dashboard data
   */
  getDashboardData(
    startTime: number = Date.now() - 24 * 60 * 60 * 1000,
    endTime: number = Date.now(),
  ): DashboardData {
    const events = this.getEvents({ startTime, endTime });
    const metrics = this.getMetrics(startTime, endTime);

    return {
      period: { start: startTime, end: endTime },
      overview: {
        totalChats: events.filter((e) => e.type === 'chat.receive').length,
        totalRAGQueries: events.filter((e) => e.type === 'rag.retrieve').length,
        totalTokens: metrics.tokenInput + metrics.tokenOutput,
        totalErrors: events.filter(
          (e) => e.type === 'chat.error' || e.type === 'provider.error',
        ).length,
      },
      performance: metrics,
      usage: this.metricsCollector.calculateUsageMetrics(events),
      topFeatures: this.metricsCollector.getTopFeatures(events, 5),
      costBreakdown: this.metricsCollector.getCostBreakdown(events),
      health: this.store.health,
      alerts: this.getActiveAlerts(),
      recommendations: this.generateRecommendations(metrics, events),
    };
  }

  /**
   * Export data
   */
  async exportData(
    format: 'json' | 'csv' = 'json',
    startTime?: number,
    endTime?: number,
  ): Promise<string> {
    const events = this.getEvents({
      startTime: startTime || Date.now() - 30 * 24 * 60 * 60 * 1000,
      endTime: endTime || Date.now(),
    });

    if (format === 'json') {
      return JSON.stringify(events, null, 2);
    } else if (format === 'csv') {
      return this.eventsToCsv(events);
    }

    throw new Error(`Unsupported export format: ${format}`);
  }

  /**
   * Clear old data
   */
  clearOldData(retentionDays: number = this.config.retentionDays): void {
    const cutoffTime = Date.now() - retentionDays * 24 * 60 * 60 * 1000;

    this.store.events = this.store.events.filter((e) => e.timestamp > cutoffTime);

    this.store.alerts = this.store.alerts.filter((a) => a.timestamp > cutoffTime);
  }

  /**
   * Flush events to storage
   */
  async flush(): Promise<void> {
    if (this.store.events.length === 0) return;

    try {
      // Store events in local storage (summarized)
      this.storeSummary();

      // Store in IndexedDB if configured
      if (
        this.config.storageBackend === 'indexeddb' ||
        this.config.storageBackend === 'both'
      ) {
        await this.storeInIndexedDB();
      }

      this.store.lastFlush = Date.now();
    } catch (error) {
      console.error('Failed to flush analytics:', error);
    }
  }

  /**
   * Get session info
   */
  getSessionInfo(): {
    sessionId: string;
    startTime: number;
    uptime: number;
    eventCount: number;
  } {
    return {
      sessionId: this.sessionId,
      startTime: this.startTime,
      uptime: Date.now() - this.startTime,
      eventCount: this.store.events.length,
    };
  }

  /**
   * Reset analytics
   */
  reset(): void {
    this.store = {
      events: [],
      metrics: new Map(),
      alerts: [],
      health: this.getDefaultHealth(),
      lastFlush: Date.now(),
    };
    this.sessionId = uuidv4();
  }

  // Private methods

  private initializeDefaultAlerts(): void {
    // High memory usage
    this.addAlertRule({
      id: 'high-memory',
      name: 'High Memory Usage',
      condition: { type: 'high_memory', threshold: 85 },
      action: { type: 'notification', title: 'Memory Warning', message: 'Memory usage is above 85%' },
      enabled: true,
      createdAt: Date.now(),
    });

    // High error rate
    this.addAlertRule({
      id: 'high-error-rate',
      name: 'High Error Rate',
      condition: { type: 'high_error_rate', threshold: 10 },
      action: { type: 'notification', title: 'Error Rate Warning', message: 'Error rate is above 10%' },
      enabled: true,
      createdAt: Date.now(),
    });

    // Rate limit warning
    this.addAlertRule({
      id: 'rate-limit',
      name: 'Rate Limit Warning',
      condition: { type: 'rate_limit', threshold: 10 },
      action: { type: 'notification', title: 'Rate Limit', message: 'Approaching rate limit' },
      enabled: true,
      createdAt: Date.now(),
    });
  }

  private checkAlertConditions(event: AnalyticsEvent): void {
    for (const rule of this.alertRules.values()) {
      if (!rule.enabled) continue;

      const shouldAlert = this.evaluateAlertCondition(event, rule.condition);
      if (shouldAlert) {
        const action = rule.action;
        if (action.type === 'notification') {
          this.createAlert(rule.id, 'warning', action.title, action.message);
        }
      }
    }
  }

  private evaluateAlertCondition(
    event: AnalyticsEvent,
    condition: AlertCondition,
  ): boolean {
    switch (condition.type) {
      case 'high_memory':
        return (
          event.metadata.memoryUsageMB &&
          (event.metadata.memoryUsageMB as number) > condition.threshold
        );
      case 'high_error_rate':
        return event.type.endsWith('.error');
      case 'rate_limit':
        return event.type === 'provider.error' &&
          event.metadata.error?.toString().includes('rate limit');
      case 'provider_down':
        return (
          event.type === 'provider.error' &&
          event.metadata.provider === condition.provider
        );
      case 'storage_quota':
        return (
          event.metadata.usagePercent &&
          (event.metadata.usagePercent as number) > condition.threshold
        );
      case 'performance_degradation':
        return (
          event.metadata.responseTime &&
          (event.metadata.responseTime as number) > condition.threshold
        );
      default:
        return false;
    }
  }

  private generateRecommendations(
    metrics: PerformanceMetrics,
    _events: AnalyticsEvent[],
  ): Recommendation[] {
    const recommendations: Recommendation[] = [];

    // Cache optimization
    if (metrics.cacheHitRate < 0.3) {
      recommendations.push({
        id: 'cache-optimization',
        priority: 'medium',
        category: 'performance',
        title: 'Low Cache Hit Rate',
        description: 'Your cache hit rate is below 30%',
        suggestion: 'Consider using RAG for frequently accessed content',
        estimatedBenefit: 'Reduce latency by 40-60%',
      });
    }

    // Error rate
    if (metrics.errorRate > 0.05) {
      recommendations.push({
        id: 'error-rate',
        priority: 'high',
        category: 'reliability',
        title: 'High Error Rate',
        description: 'Your error rate is above 5%',
        suggestion: 'Review error logs and provider health status',
        estimatedBenefit: 'Improve reliability by up to 95%',
      });
    }

    // Memory usage
    if (metrics.memoryUsageMB > 500) {
      recommendations.push({
        id: 'memory-optimization',
        priority: 'medium',
        category: 'performance',
        title: 'High Memory Usage',
        description: 'Memory usage is above 500MB',
        suggestion: 'Clear cache or use alternative providers',
        estimatedBenefit: 'Reduce memory usage by 30-50%',
      });
    }

    // Cost optimization
    const totalCost = metrics.costUSD;
    if (totalCost > 10) {
      recommendations.push({
        id: 'cost-optimization',
        priority: 'low',
        category: 'cost',
        title: 'High API Costs',
        description: `Total cost is $${totalCost.toFixed(2)}`,
        suggestion: 'Switch to more cost-effective providers or batch requests',
        estimatedBenefit: 'Reduce costs by 20-40%',
      });
    }

    return recommendations;
  }

  private async storeInIndexedDB(): Promise<void> {
    try {
      const db = await this.openIndexedDB();
      const tx = db.transaction('analytics', 'readwrite');
      const store = tx.objectStore('analytics');

      for (const event of this.store.events) {
        await store.add(event);
      }

      await new Promise((resolve, reject) => {
        tx.oncomplete = resolve;
        tx.onerror = reject;
      });
    } catch (error) {
      console.warn('Failed to store in IndexedDB:', error);
    }
  }

  private openIndexedDB(): Promise<IDBDatabase> {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open('VectoraAnalytics', 1);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => resolve(request.result);

      request.onupgradeneeded = (event) => {
        const db = (event.target as IDBOpenDBRequest).result;
        if (!db.objectStoreNames.contains('analytics')) {
          db.createObjectStore('analytics', { keyPath: 'id' });
        }
      };
    });
  }

  private storeSummary(): void {
    const summary = {
      sessionId: this.sessionId,
      eventCount: this.store.events.length,
      timestamp: Date.now(),
      period: {
        start: this.startTime,
        end: Date.now(),
      },
      eventTypes: this.getEventTypeCounts(),
    } as const;

    try {
      const existing = localStorage.getItem('vectora-analytics-summary');
      const summaries = existing ? JSON.parse(existing) : [];
      summaries.push(summary);

      // Keep only last 24 hours
      const cutoff = Date.now() - 24 * 60 * 60 * 1000;
      const filtered = summaries.filter((s: { timestamp: number }) => s.timestamp > cutoff);

      localStorage.setItem('vectora-analytics-summary', JSON.stringify(filtered));
    } catch (error) {
      console.warn('Failed to store analytics summary:', error);
    }
  }

  private getEventTypeCounts(): Record<string, number> {
    const counts: Record<string, number> = {};
    for (const event of this.store.events) {
      counts[event.type] = (counts[event.type] || 0) + 1;
    }
    return counts;
  }

  private eventsToCsv(events: AnalyticsEvent[]): string {
    if (events.length === 0) return 'timestamp,type,sessionId,metadata\n';

    const headers = ['timestamp', 'type', 'sessionId', 'metadata'];
    const rows = events.map((e) => [
      new Date(e.timestamp).toISOString(),
      e.type,
      e.sessionId,
      JSON.stringify(e.metadata),
    ]);

    const csvContent = [
      headers.join(','),
      ...rows.map((r) => r.map((v) => `"${v}"`).join(',')),
    ].join('\n');

    return csvContent;
  }

  private startAutoFlush(): void {
    this.flushInterval = setInterval(() => {
      this.flush();
    }, this.config.flushIntervalMs);
  }

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

  private getDefaultHealth(): HealthStatus {
    return {
      timestamp: Date.now(),
      providers: {},
      network: { latency: 0, connected: true, lastCheck: Date.now() },
      system: { uptime: 0, memoryUsage: 0, cpuUsage: 0, errorCount: 0 },
      storage: {
        quotaBytes: 0,
        usedBytes: 0,
        remainingBytes: 0,
        usagePercent: 0,
      },
    };
  }

  private createDummyEvent(
    type: EventType,
    metadata: Record<string, unknown>,
  ): AnalyticsEvent {
    return {
      id: uuidv4(),
      type,
      timestamp: Date.now(),
      metadata,
      sessionId: this.sessionId,
    };
  }
}
