/**
 * Analytics Types
 * Type-safe analytics for Vectora VS Code Extension
 */

export type EventType =
  | 'chat.send'
  | 'chat.receive'
  | 'chat.error'
  | 'rag.search'
  | 'rag.retrieve'
  | 'rag.rank'
  | 'provider.switch'
  | 'provider.fallback'
  | 'provider.error'
  | 'ui.action'
  | 'system.startup'
  | 'system.error'
  | 'system.warning'
  | 'cache.hit'
  | 'cache.miss'
  | 'memory.warning'
  | 'rate.limit.warning'
  | 'token.usage';

export type AlertSeverity = 'info' | 'warning' | 'error' | 'critical';

export interface AnalyticsEvent {
  id: string;
  type: EventType;
  timestamp: number;
  duration?: number;
  metadata: Record<string, unknown>;
  userId?: string;
  sessionId: string;
}

export interface ChatEvent extends AnalyticsEvent {
  type: 'chat.send' | 'chat.receive' | 'chat.error';
  metadata: {
    messageLength: number;
    tokenCount?: number;
    provider?: string;
    model?: string;
    error?: string;
    responseTime?: number;
  };
}

export interface RAGEvent extends AnalyticsEvent {
  type: 'rag.search' | 'rag.retrieve' | 'rag.rank';
  metadata: {
    query: string;
    resultCount: number;
    relevanceScore?: number;
    executionTime: number;
    cacheHit?: boolean;
  };
}

export interface ProviderEvent extends AnalyticsEvent {
  type: 'provider.switch' | 'provider.fallback' | 'provider.error';
  metadata: {
    provider: string;
    previousProvider?: string;
    reason?: string;
    error?: string;
  };
}

export interface UIActionEvent extends AnalyticsEvent {
  type: 'ui.action';
  metadata: {
    action: string;
    component: string;
    targetId?: string;
    value?: unknown;
  };
}

export interface PerformanceMetrics {
  responseTime: number; // ms
  tokenInput: number;
  tokenOutput: number;
  costUSD: number;
  cacheHitRate: number; // 0-1
  errorRate: number; // 0-1
  memoryUsageMB: number;
  cpuUsagePercent: number;
}

export interface HealthStatus {
  timestamp: number;
  providers: Record<string, ProviderHealth>;
  network: NetworkHealth;
  system: SystemHealth;
  storage: StorageHealth;
}

export interface ProviderHealth {
  available: boolean;
  lastCheck: number;
  errorCount: number;
  successCount: number;
  averageResponseTime: number;
  rateLimitRemaining?: number;
}

export interface NetworkHealth {
  latency: number;
  connected: boolean;
  lastCheck: number;
}

export interface SystemHealth {
  uptime: number;
  memoryUsage: number;
  cpuUsage: number;
  errorCount: number;
}

export interface StorageHealth {
  quotaBytes: number;
  usedBytes: number;
  remainingBytes: number;
  usagePercent: number;
}

export interface AlertConfig {
  enabled: boolean;
  severity: AlertSeverity;
  threshold?: number;
  cooldownMs?: number;
}

export interface AlertRule {
  id: string;
  name: string;
  condition: AlertCondition;
  action: AlertAction;
  enabled: boolean;
  createdAt: number;
}

export type AlertCondition =
  | { type: 'high_memory'; threshold: number }
  | { type: 'high_error_rate'; threshold: number }
  | { type: 'rate_limit'; threshold: number }
  | { type: 'provider_down'; provider: string }
  | { type: 'storage_quota'; threshold: number }
  | { type: 'performance_degradation'; threshold: number };

export type AlertAction =
  | { type: 'notification'; title: string; message: string }
  | { type: 'log'; level: 'info' | 'warn' | 'error' }
  | { type: 'metric'; metric: string };

export interface Alert {
  id: string;
  ruleId: string;
  severity: AlertSeverity;
  title: string;
  message: string;
  timestamp: number;
  resolved: boolean;
  resolvedAt?: number;
}

export interface UsageMetrics {
  chatMessages: number;
  ragQueries: number;
  tokensUsed: number;
  providerBreakdown: Record<string, number>;
  featureUsage: Record<string, number>;
  uniqueUsers?: number;
  sessionDuration: number; // ms
}

export interface AnalyticsConfig {
  enabled: boolean;
  trackingLevel: 'none' | 'basic' | 'detailed';
  retentionDays: number;
  sessionIdStrategy: 'device' | 'user' | 'session';
  privacyLevel: 'strict' | 'normal' | 'detailed';
  batchSize: number;
  flushIntervalMs: number;
  storageBackend: 'localStorage' | 'indexeddb' | 'both';
  exportFormats: ('json' | 'csv')[];
}

export interface DashboardData {
  period: { start: number; end: number };
  overview: {
    totalChats: number;
    totalRAGQueries: number;
    totalTokens: number;
    totalErrors: number;
  };
  performance: PerformanceMetrics;
  usage: UsageMetrics;
  topFeatures: Array<{ feature: string; count: number; percentage: number }>;
  costBreakdown: Array<{ provider: string; cost: number; percentage: number }>;
  health: HealthStatus;
  alerts: Alert[];
  recommendations: Recommendation[];
}

export interface Recommendation {
  id: string;
  priority: 'low' | 'medium' | 'high';
  category: 'performance' | 'cost' | 'reliability' | 'usage';
  title: string;
  description: string;
  suggestion: string;
  estimatedBenefit: string;
  action?: string;
}

export interface ExportData {
  format: 'json' | 'csv';
  data: unknown;
  filename: string;
  timestamp: number;
}

export interface AnalyticsStore {
  events: AnalyticsEvent[];
  metrics: Map<string, PerformanceMetrics>;
  alerts: Alert[];
  health: HealthStatus;
  lastFlush: number;
}

export interface EventFilter {
  startTime?: number;
  endTime?: number;
  types?: EventType[];
  sessionId?: string;
  userId?: string;
  limit?: number;
  offset?: number;
}

export interface MetricsAggregation {
  period: string; // 'hour' | 'day' | 'week' | 'month'
  data: {
    timestamp: number;
    metrics: PerformanceMetrics;
  }[];
}

export interface TrackingConfig {
  enabled: boolean;
  anonymizeIp: boolean;
  cookieConsent: boolean;
  doNotTrack: boolean;
  dataRetention: number; // days
}
