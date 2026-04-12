/**
 * Memory Monitor
 * Monitors and tracks memory usage
 */

import { EventTracker } from '../analytics/eventTracker';

export interface MemorySnapshot {
  timestamp: number;
  usedJSHeapSize: number;
  totalJSHeapSize: number;
  jsHeapSizeLimit: number;
  externalMemoryUsage: number;
}

export interface MemoryWarning {
  timestamp: number;
  severity: 'warning' | 'critical';
  usagePercent: number;
  usedMB: number;
  limitMB: number;
}

export class MemoryMonitor {
  private sessionId: string;
  private eventTracker: EventTracker;
  private snapshots: MemorySnapshot[] = [];
  private warnings: MemoryWarning[] = [];
  private warningThresholds = {
    warning: 0.7, // 70%
    critical: 0.85, // 85%
  };
  private checkInterval?: NodeJS.Timer;

  constructor(sessionId: string, eventTracker: EventTracker) {
    this.sessionId = sessionId;
    this.eventTracker = eventTracker;
  }

  /**
   * Take a memory snapshot
   */
  takeSnapshot(): MemorySnapshot | null {
    if (typeof performance === 'undefined' || !performance.memory) {
      return null;
    }

    const snapshot: MemorySnapshot = {
      timestamp: Date.now(),
      usedJSHeapSize: performance.memory.usedJSHeapSize,
      totalJSHeapSize: performance.memory.totalJSHeapSize,
      jsHeapSizeLimit: performance.memory.jsHeapSizeLimit,
      externalMemoryUsage: (performance.memory as unknown as { externalMemoryUsageSize?: number })
        .externalMemoryUsageSize || 0,
    };

    this.snapshots.push(snapshot);

    // Check for memory warnings
    this.checkMemoryWarnings(snapshot);

    return snapshot;
  }

  /**
   * Get current memory usage percentage
   */
  getCurrentUsagePercent(): number {
    if (typeof performance === 'undefined' || !performance.memory) {
      return 0;
    }

    return performance.memory.usedJSHeapSize / performance.memory.jsHeapSizeLimit;
  }

  /**
   * Get current memory usage in MB
   */
  getCurrentUsageMB(): number {
    if (typeof performance === 'undefined' || !performance.memory) {
      return 0;
    }

    return performance.memory.usedJSHeapSize / (1024 * 1024);
  }

  /**
   * Get memory limit in MB
   */
  getMemoryLimitMB(): number {
    if (typeof performance === 'undefined' || !performance.memory) {
      return 0;
    }

    return performance.memory.jsHeapSizeLimit / (1024 * 1024);
  }

  /**
   * Get average memory usage (last N snapshots)
   */
  getAverageUsage(lastN: number = 100): number {
    const relevant = this.snapshots.slice(-lastN);
    if (relevant.length === 0) return 0;

    const sum = relevant.reduce((acc, s) => acc + s.usedJSHeapSize, 0);
    return sum / relevant.length / (1024 * 1024); // MB
  }

  /**
   * Get peak memory usage
   */
  getPeakUsage(): number {
    if (this.snapshots.length === 0) return 0;

    const maxUsage = Math.max(...this.snapshots.map((s) => s.usedJSHeapSize));
    return maxUsage / (1024 * 1024); // MB
  }

  /**
   * Get memory trend
   */
  getMemoryTrend(
    lastN: number = 50,
  ): Array<{ timestamp: number; usageMB: number; usagePercent: number }> {
    const relevant = this.snapshots.slice(-lastN);

    return relevant.map((s) => ({
      timestamp: s.timestamp,
      usageMB: s.usedJSHeapSize / (1024 * 1024),
      usagePercent: s.usedJSHeapSize / s.jsHeapSizeLimit,
    }));
  }

  /**
   * Detect memory leaks
   */
  detectMemoryLeak(): {
    detected: boolean;
    trend: number;
    recommendation: string;
  } {
    if (this.snapshots.length < 10) {
      return { detected: false, trend: 0, recommendation: 'Not enough data' };
    }

    const recent = this.snapshots.slice(-10);
    const oldest = this.snapshots[Math.max(0, this.snapshots.length - 100)];

    if (!oldest) {
      return { detected: false, trend: 0, recommendation: 'Not enough historical data' };
    }

    const recentAvg =
      recent.reduce((sum, s) => sum + s.usedJSHeapSize, 0) / recent.length;
    const oldestUsage = oldest.usedJSHeapSize;

    const trend = (recentAvg - oldestUsage) / oldestUsage;
    const detected = trend > 0.1; // 10% growth

    let recommendation = '';
    if (trend > 0.5) {
      recommendation = 'Critical memory leak detected. Possible causes: circular references, event listener leaks, DOM node retention';
    } else if (trend > 0.2) {
      recommendation = 'Possible memory leak. Check for undisposed resources and cleanup handlers';
    } else if (detected) {
      recommendation = 'Gradual memory growth detected. Monitor closely';
    } else {
      recommendation = 'Memory usage is stable';
    }

    return { detected, trend, recommendation };
  }

  /**
   * Clear memory (triggers garbage collection if available)
   */
  async clearMemory(): Promise<void> {
    // Note: This is just to allow potential cleanup
    // Actual GC is not directly controllable from JavaScript
    this.snapshots = this.snapshots.slice(-10);
    this.warnings = this.warnings.slice(-10);
  }

  /**
   * Get memory health status
   */
  getHealthStatus(): {
    status: 'healthy' | 'warning' | 'critical';
    usagePercent: number;
    message: string;
  } {
    const usagePercent = this.getCurrentUsagePercent();

    if (usagePercent >= this.warningThresholds.critical) {
      return {
        status: 'critical',
        usagePercent,
        message: `Critical memory usage: ${(usagePercent * 100).toFixed(1)}%`,
      };
    } else if (usagePercent >= this.warningThresholds.warning) {
      return {
        status: 'warning',
        usagePercent,
        message: `High memory usage: ${(usagePercent * 100).toFixed(1)}%`,
      };
    }

    return {
      status: 'healthy',
      usagePercent,
      message: `Memory usage: ${(usagePercent * 100).toFixed(1)}%`,
    };
  }

  /**
   * Set warning threshold
   */
  setWarningThreshold(severity: 'warning' | 'critical', percent: number): void {
    if (percent > 0 && percent < 1) {
      this.warningThresholds[severity] = percent;
    }
  }

  /**
   * Start auto-monitoring
   */
  startAutoMonitoring(intervalMs: number = 5000): void {
    this.checkInterval = setInterval(() => {
      this.takeSnapshot();
    }, intervalMs);
  }

  /**
   * Stop auto-monitoring
   */
  stopAutoMonitoring(): void {
    if (this.checkInterval) {
      clearInterval(this.checkInterval);
      this.checkInterval = undefined;
    }
  }

  /**
   * Get all snapshots
   */
  getSnapshots(): MemorySnapshot[] {
    return [...this.snapshots];
  }

  /**
   * Get all warnings
   */
  getWarnings(): MemoryWarning[] {
    return [...this.warnings];
  }

  /**
   * Clear warnings
   */
  clearWarnings(): void {
    this.warnings = [];
  }

  /**
   * Get memory statistics
   */
  getStatistics(): {
    current: number;
    average: number;
    peak: number;
    limit: number;
    currentPercent: number;
  } {
    return {
      current: this.getCurrentUsageMB(),
      average: this.getAverageUsage(),
      peak: this.getPeakUsage(),
      limit: this.getMemoryLimitMB(),
      currentPercent: this.getCurrentUsagePercent() * 100,
    };
  }

  /**
   * Reset monitor
   */
  reset(): void {
    this.snapshots = [];
    this.warnings = [];
    this.stopAutoMonitoring();
  }

  // Private methods

  private checkMemoryWarnings(snapshot: MemorySnapshot): void {
    const usagePercent = snapshot.usedJSHeapSize / snapshot.jsHeapSizeLimit;

    let severity: 'warning' | 'critical' = 'warning';
    if (usagePercent >= this.warningThresholds.critical) {
      severity = 'critical';
    }

    if (usagePercent >= this.warningThresholds.warning) {
      const warning: MemoryWarning = {
        timestamp: snapshot.timestamp,
        severity,
        usagePercent,
        usedMB: snapshot.usedJSHeapSize / (1024 * 1024),
        limitMB: snapshot.jsHeapSizeLimit / (1024 * 1024),
      };

      this.warnings.push(warning);

      // Track in event tracker
      this.eventTracker.logMemoryWarning(
        usagePercent * 100,
        snapshot.usedJSHeapSize / (1024 * 1024),
      );

      // Limit warnings history
      if (this.warnings.length > 1000) {
        this.warnings = this.warnings.slice(-1000);
      }
    }
  }
}
