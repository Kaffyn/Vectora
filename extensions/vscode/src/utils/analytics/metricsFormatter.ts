/**
 * Metrics Formatter
 * Utility functions for formatting analytics metrics
 */

import {
  PerformanceMetrics,
  HealthStatus,
  Alert,
  UsageMetrics,
  Recommendation,
} from '../../types/analytics';

export class MetricsFormatter {
  /**
   * Format response time
   */
  static formatResponseTime(ms: number): string {
    if (ms < 1000) {
      return `${Math.round(ms)}ms`;
    }
    return `${(ms / 1000).toFixed(2)}s`;
  }

  /**
   * Format memory usage
   */
  static formatMemory(bytes: number): string {
    const units = ['B', 'KB', 'MB', 'GB'];
    let size = bytes;
    let unitIndex = 0;

    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024;
      unitIndex++;
    }

    return `${size.toFixed(1)} ${units[unitIndex]}`;
  }

  /**
   * Format percentage
   */
  static formatPercent(value: number): string {
    return `${(value * 100).toFixed(1)}%`;
  }

  /**
   * Format cost in USD
   */
  static formatCost(usd: number): string {
    if (usd < 0.01) {
      return `$${(usd * 1000).toFixed(2)}m`; // millicents
    }
    return `$${usd.toFixed(4)}`;
  }

  /**
   * Format tokens
   */
  static formatTokens(count: number): string {
    if (count >= 1000000) {
      return `${(count / 1000000).toFixed(1)}M`;
    }
    if (count >= 1000) {
      return `${(count / 1000).toFixed(1)}K`;
    }
    return String(count);
  }

  /**
   * Format duration
   */
  static formatDuration(ms: number): string {
    const seconds = Math.floor(ms / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (days > 0) {
      return `${days}d ${hours % 24}h`;
    }
    if (hours > 0) {
      return `${hours}h ${minutes % 60}m`;
    }
    if (minutes > 0) {
      return `${minutes}m ${seconds % 60}s`;
    }
    return `${seconds}s`;
  }

  /**
   * Format timestamp
   */
  static formatTimestamp(time: number): string {
    const date = new Date(time);
    return date.toLocaleString();
  }

  /**
   * Format date short
   */
  static formatDateShort(time: number): string {
    const date = new Date(time);
    const today = new Date();

    if (date.toDateString() === today.toDateString()) {
      return date.toLocaleTimeString([], {
        hour: '2-digit',
        minute: '2-digit',
      });
    }

    return date.toLocaleDateString([], {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  /**
   * Format health status
   */
  static formatHealthStatus(status: 'healthy' | 'degraded' | 'critical'): string {
    switch (status) {
      case 'healthy':
        return '✓ Healthy';
      case 'degraded':
        return '⚠ Degraded';
      case 'critical':
        return '✗ Critical';
    }
  }

  /**
   * Get health status color
   */
  static getHealthStatusColor(status: 'healthy' | 'degraded' | 'critical'): string {
    switch (status) {
      case 'healthy':
        return '#10b981'; // green
      case 'degraded':
        return '#f59e0b'; // amber
      case 'critical':
        return '#ef4444'; // red
    }
  }

  /**
   * Format alert severity
   */
  static formatAlertSeverity(severity: string): string {
    switch (severity) {
      case 'info':
        return 'ℹ️ Info';
      case 'warning':
        return '⚠️ Warning';
      case 'error':
        return '❌ Error';
      case 'critical':
        return '🔴 Critical';
      default:
        return severity;
    }
  }

  /**
   * Get severity color
   */
  static getSeverityColor(severity: string): string {
    switch (severity) {
      case 'info':
        return '#3b82f6'; // blue
      case 'warning':
        return '#f59e0b'; // amber
      case 'error':
        return '#ef4444'; // red
      case 'critical':
        return '#dc2626'; // dark red
      default:
        return '#6b7280'; // gray
    }
  }

  /**
   * Format metrics summary
   */
  static formatMetricsSummary(metrics: PerformanceMetrics): {
    responseTime: string;
    memory: string;
    cpu: string;
    tokens: string;
    cost: string;
    cacheHit: string;
    errorRate: string;
  } {
    return {
      responseTime: this.formatResponseTime(metrics.responseTime),
      memory: this.formatMemory(metrics.memoryUsageMB * 1024 * 1024),
      cpu: `${metrics.cpuUsagePercent.toFixed(1)}%`,
      tokens: this.formatTokens(metrics.tokenInput + metrics.tokenOutput),
      cost: this.formatCost(metrics.costUSD),
      cacheHit: this.formatPercent(metrics.cacheHitRate),
      errorRate: this.formatPercent(metrics.errorRate),
    };
  }

  /**
   * Format usage metrics summary
   */
  static formatUsageSummary(usage: UsageMetrics): {
    chats: string;
    ragQueries: string;
    tokens: string;
    duration: string;
  } {
    return {
      chats: String(usage.chatMessages),
      ragQueries: String(usage.ragQueries),
      tokens: this.formatTokens(usage.tokensUsed),
      duration: this.formatDuration(usage.sessionDuration),
    };
  }

  /**
   * Format recommendation
   */
  static formatRecommendation(rec: Recommendation): {
    icon: string;
    title: string;
    description: string;
    action: string;
  } {
    const icons: Record<string, string> = {
      performance: '⚡',
      cost: '💰',
      reliability: '🛡️',
      usage: '📊',
    };

    const priorityPrefix =
      rec.priority === 'high'
        ? '🔴'
        : rec.priority === 'medium'
          ? '🟡'
          : '🟢';

    return {
      icon: `${priorityPrefix} ${icons[rec.category] || '💡'}`,
      title: rec.title,
      description: rec.description,
      action: rec.suggestion,
    };
  }

  /**
   * Create CSV line
   */
  static createCSVLine(values: (string | number | boolean)[]): string {
    return values
      .map((v) => {
        const str = String(v);
        return str.includes(',') || str.includes('"') ? `"${str.replace(/"/g, '""')}"` : str;
      })
      .join(',');
  }

  /**
   * Format number with thousand separators
   */
  static formatNumber(num: number, decimals: number = 0): string {
    return new Intl.NumberFormat('en-US', {
      minimumFractionDigits: decimals,
      maximumFractionDigits: decimals,
    }).format(num);
  }

  /**
   * Format large number (K, M, B)
   */
  static formatLargeNumber(num: number): string {
    if (num >= 1000000000) {
      return `${(num / 1000000000).toFixed(1)}B`;
    }
    if (num >= 1000000) {
      return `${(num / 1000000).toFixed(1)}M`;
    }
    if (num >= 1000) {
      return `${(num / 1000).toFixed(1)}K`;
    }
    return String(num);
  }

  /**
   * Get trend indicator
   */
  static getTrendIndicator(
    current: number,
    previous: number,
  ): { symbol: string; text: string; improvement: number } {
    if (current === previous) {
      return { symbol: '→', text: 'unchanged', improvement: 0 };
    }

    const improvement = ((current - previous) / previous) * 100;
    const symbol = current > previous ? '↑' : '↓';

    return {
      symbol,
      text: `${Math.abs(improvement).toFixed(1)}%`,
      improvement,
    };
  }
}
